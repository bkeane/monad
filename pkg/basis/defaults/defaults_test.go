package defaults

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDerive_Success(t *testing.T) {
	basis, err := Derive()
	require.NoError(t, err)
	assert.NotNil(t, basis)

	// Validate that all templates are loaded and non-empty
	assert.NotEmpty(t, basis.PolicyTemplate())
	assert.NotEmpty(t, basis.RoleTemplate())
	assert.NotEmpty(t, basis.EnvTemplate())

	// Validate that templates contain expected content
	assert.Contains(t, basis.PolicyTemplate(), "Version")
	assert.Contains(t, basis.PolicyTemplate(), "Statement")
	assert.Contains(t, basis.PolicyTemplate(), "logs:")

	assert.Contains(t, basis.RoleTemplate(), "Version")
	assert.Contains(t, basis.RoleTemplate(), "TrustLambda")
	assert.Contains(t, basis.RoleTemplate(), "lambda.amazonaws.com")

	assert.Contains(t, basis.EnvTemplate(), "GIT_REPO")
	assert.Contains(t, basis.EnvTemplate(), "GIT_BRANCH")
	assert.Contains(t, basis.EnvTemplate(), "GIT_SHA")
}

func TestDerive_PolicyTemplate(t *testing.T) {
	basis, err := Derive()
	require.NoError(t, err)

	policy := basis.PolicyTemplate()

	// Should be valid JSON (even with template variables)
	var policyJSON interface{}
	// Replace template variables with dummy values for JSON validation
	validationPolicy := strings.ReplaceAll(policy, "{{.Account.Region}}", "us-east-1")
	validationPolicy = strings.ReplaceAll(validationPolicy, "{{.Account.Id}}", "123456789012")
	validationPolicy = strings.ReplaceAll(validationPolicy, "{{.Resource.Name}}", "test-resource")
	
	err = json.Unmarshal([]byte(validationPolicy), &policyJSON)
	assert.NoError(t, err, "Policy template should be valid JSON after variable substitution")

	// Should contain template variables
	assert.Contains(t, policy, "{{.Account.Region}}")
	assert.Contains(t, policy, "{{.Account.Id}}")
	assert.Contains(t, policy, "{{.Resource.Name}}")

	// Should contain expected AWS IAM policy structure
	assert.Contains(t, policy, "\"Version\": \"2012-10-17\"")
	assert.Contains(t, policy, "\"Statement\"")
	assert.Contains(t, policy, "\"Effect\": \"Allow\"")
	assert.Contains(t, policy, "\"Action\"")
	assert.Contains(t, policy, "\"Resource\"")

	// Should contain CloudWatch Logs permissions
	assert.Contains(t, policy, "logs:*")
	assert.Contains(t, policy, "arn:aws:logs:")
}

func TestDerive_RoleTemplate(t *testing.T) {
	basis, err := Derive()
	require.NoError(t, err)

	role := basis.RoleTemplate()

	// Should be valid JSON
	var roleJSON interface{}
	err = json.Unmarshal([]byte(role), &roleJSON)
	assert.NoError(t, err, "Role template should be valid JSON")

	// Should contain expected AWS IAM trust policy structure
	assert.Contains(t, role, "\"Version\": \"2012-10-17\"")
	assert.Contains(t, role, "\"Statement\"")
	assert.Contains(t, role, "\"Effect\": \"Allow\"")
	assert.Contains(t, role, "\"Principal\"")
	assert.Contains(t, role, "\"Action\": \"sts:AssumeRole\"")

	// Should trust Lambda service
	assert.Contains(t, role, "\"Service\": \"lambda.amazonaws.com\"")
	assert.Contains(t, role, "TrustLambda")
}

func TestDerive_EnvTemplate(t *testing.T) {
	basis, err := Derive()
	require.NoError(t, err)

	env := basis.EnvTemplate()

	// Should contain Git-related environment variables
	expectedVars := []string{
		"GIT_REPO={{.Git.Repo}}",
		"GIT_BRANCH={{.Git.Branch}}",
		"GIT_SHA={{.Git.Sha}}",
	}

	for _, expectedVar := range expectedVars {
		assert.Contains(t, env, expectedVar)
	}

	// Should be in valid key=value format (simple validation)
	lines := strings.Split(strings.TrimSpace(env), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			parts := strings.SplitN(line, "=", 2)
			assert.Len(t, parts, 2, "Each non-comment line should be in key=value format: %s", line)
			assert.NotEmpty(t, parts[0], "Key should not be empty: %s", line)
		}
	}
}

func TestBasis_Accessors(t *testing.T) {
	basis := &Basis{
		Policy: "test-policy",
		Role:   "test-role",
		Env:    "test-env",
	}

	assert.Equal(t, "test-policy", basis.PolicyTemplate())
	assert.Equal(t, "test-role", basis.RoleTemplate())
	assert.Equal(t, "test-env", basis.EnvTemplate())
}

func TestBasis_Validate(t *testing.T) {
	tests := []struct {
		name    string
		basis   *Basis
		wantErr bool
	}{
		{
			name: "valid basis",
			basis: &Basis{
				Policy: "policy-content",
				Role:   "role-content",
				Env:    "env-content",
			},
			wantErr: false,
		},
		{
			name: "missing policy",
			basis: &Basis{
				Role: "role-content",
				Env:  "env-content",
			},
			wantErr: true,
		},
		{
			name: "missing role",
			basis: &Basis{
				Policy: "policy-content",
				Env:    "env-content",
			},
			wantErr: true,
		},
		{
			name: "missing env",
			basis: &Basis{
				Policy: "policy-content",
				Role:   "role-content",
			},
			wantErr: true,
		},
		{
			name: "empty policy",
			basis: &Basis{
				Policy: "",
				Role:   "role-content",
				Env:    "env-content",
			},
			wantErr: true,
		},
		{
			name: "empty role",
			basis: &Basis{
				Policy: "policy-content",
				Role:   "",
				Env:    "env-content",
			},
			wantErr: true,
		},
		{
			name: "empty env",
			basis: &Basis{
				Policy: "policy-content",
				Role:   "role-content",
				Env:    "",
			},
			wantErr: true,
		},
		{
			name:    "all empty",
			basis:   &Basis{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.basis.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRead_Success(t *testing.T) {
	// Test the internal read function with known embedded files
	tests := []struct {
		name     string
		path     string
		contains []string
	}{
		{
			name: "policy template",
			path: "embed/policy.json.tmpl",
			contains: []string{
				"Version",
				"Statement",
				"logs:",
				"{{.Account.Region}}",
				"{{.Resource.Name}}",
			},
		},
		{
			name: "role template",
			path: "embed/role.json.tmpl",
			contains: []string{
				"Version",
				"TrustLambda",
				"lambda.amazonaws.com",
				"sts:AssumeRole",
			},
		},
		{
			name: "env template",
			path: "embed/env.tmpl",
			contains: []string{
				"GIT_REPO=",
				"GIT_BRANCH=",
				"GIT_SHA=",
				"{{.Git.Repo}}",
				"{{.Git.Branch}}",
				"{{.Git.Sha}}",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := read(tt.path)
			require.NoError(t, err)
			assert.NotEmpty(t, content)

			for _, expectedContent := range tt.contains {
				assert.Contains(t, content, expectedContent)
			}
		})
	}
}

func TestRead_FileNotFound(t *testing.T) {
	content, err := read("embed/nonexistent.tmpl")
	assert.Error(t, err)
	assert.Empty(t, content)
	assert.Contains(t, err.Error(), "nonexistent.tmpl")
}

func TestRead_InvalidPath(t *testing.T) {
	content, err := read("invalid/path/file.tmpl")
	assert.Error(t, err)
	assert.Empty(t, content)
}

// TestDerive_TemplateVariables ensures that templates contain all expected
// template variables for proper rendering
func TestDerive_TemplateVariables(t *testing.T) {
	basis, err := Derive()
	require.NoError(t, err)

	// Policy template should contain account and resource variables
	policy := basis.PolicyTemplate()
	policyVars := []string{"{{.Account.Id}}", "{{.Account.Region}}", "{{.Resource.Name}}"}
	for _, v := range policyVars {
		assert.Contains(t, policy, v, "Policy template should contain %s", v)
	}

	// Env template should contain git variables (only those actually present)
	env := basis.EnvTemplate()
	envVars := []string{"{{.Git.Repo}}", "{{.Git.Branch}}", "{{.Git.Sha}}"}
	for _, v := range envVars {
		assert.Contains(t, env, v, "Env template should contain %s", v)
	}

	// Role template doesn't use template variables (it's static)
	// Just verify it's not empty
	role := basis.RoleTemplate()
	assert.NotEmpty(t, role, "Role template should not be empty")
}

// TestDerive_Consistency ensures that Derive returns the same content
// across multiple calls (embedded files should be deterministic)
func TestDerive_Consistency(t *testing.T) {
	basis1, err1 := Derive()
	require.NoError(t, err1)

	basis2, err2 := Derive()
	require.NoError(t, err2)

	// Should return identical content
	assert.Equal(t, basis1.PolicyTemplate(), basis2.PolicyTemplate())
	assert.Equal(t, basis1.RoleTemplate(), basis2.RoleTemplate())
	assert.Equal(t, basis1.EnvTemplate(), basis2.EnvTemplate())
}