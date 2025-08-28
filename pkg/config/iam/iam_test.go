package iam

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bkeane/monad/pkg/basis/mock"
)

func TestDerive_Success(t *testing.T) {
	setup := mock.NewIAMTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)
	assert.NotNil(t, config)

	// Verify components are properly set
	assert.NotNil(t, config.Client())
	assert.Equal(t, "test-repo-test-branch-test-service", config.RoleName())
	assert.Equal(t, "test-repo-test-branch-test-service", config.PolicyName())
}

func TestDerive_RoleArn(t *testing.T) {
	setup := mock.NewIAMTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)

	arn := config.RoleArn()
	expected := "arn:aws:iam::123456789012:role/test-repo-test-branch-test-service"
	assert.Equal(t, expected, arn)
}

func TestDerive_PolicyArn(t *testing.T) {
	setup := mock.NewIAMTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)

	arn := config.PolicyArn()
	expected := "arn:aws:iam::123456789012:policy/test-repo-test-branch-test-service"
	assert.Equal(t, expected, arn)
}

func TestDerive_PolicyDocument(t *testing.T) {
	setup := mock.NewIAMTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)

	policy := config.PolicyDocument()
	assert.NotEmpty(t, policy)

	// Should be valid JSON
	var parsed map[string]interface{}
	err = json.Unmarshal([]byte(policy), &parsed)
	require.NoError(t, err)
	
	// Should have basic IAM policy structure
	assert.Contains(t, parsed, "Version")
	assert.Contains(t, parsed, "Statement")
}

func TestDerive_RoleDocument(t *testing.T) {
	setup := mock.NewIAMTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)

	role := config.RoleDocument()
	assert.NotEmpty(t, role)

	// Should be valid JSON
	var parsed map[string]interface{}
	err = json.Unmarshal([]byte(role), &parsed)
	require.NoError(t, err)
	
	// Should have basic IAM role structure
	assert.Contains(t, parsed, "Version")
	assert.Contains(t, parsed, "Statement")
	
	// Should contain assume role policy
	statements, ok := parsed["Statement"].([]interface{})
	require.True(t, ok)
	require.Len(t, statements, 1)
	
	statement := statements[0].(map[string]interface{})
	assert.Equal(t, "Allow", statement["Effect"])
	assert.Equal(t, "sts:AssumeRole", statement["Action"])
}

func TestDerive_WithCustomPolicyTemplate(t *testing.T) {
	setup := mock.NewIAMTestSetup()
	
	// Create a temp file for custom policy template
	tmpFile := t.TempDir() + "/custom-policy.json"
	customPolicy := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": "s3:GetObject",
				"Resource": "arn:aws:s3:::{{.Service.Name}}/*"
			}
		]
	}`
	require.NoError(t, os.WriteFile(tmpFile, []byte(customPolicy), 0644))
	
	setup.ApplyWithOverrides(t, map[string]string{
		"MONAD_POLICY": tmpFile,
	})
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)

	policy := config.PolicyDocument()
	assert.Contains(t, policy, "s3:GetObject")
	assert.Contains(t, policy, "test-service") // Template should be rendered
}

func TestDerive_WithCustomRoleTemplate(t *testing.T) {
	setup := mock.NewIAMTestSetup()
	
	// Create a temp file for custom role template
	tmpFile := t.TempDir() + "/custom-role.json"
	customRole := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {"Service": "{{.Service.Name}}.amazonaws.com"},
				"Action": "sts:AssumeRole"
			}
		]
	}`
	require.NoError(t, os.WriteFile(tmpFile, []byte(customRole), 0644))
	
	setup.ApplyWithOverrides(t, map[string]string{
		"MONAD_ROLE": tmpFile,
	})
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)

	role := config.RoleDocument()
	assert.Contains(t, role, "test-service.amazonaws.com") // Template should be rendered
}

func TestDerive_WithBoundaryPolicy(t *testing.T) {
	tests := []struct {
		name     string
		boundary string
		expected string
	}{
		{
			name:     "boundary policy name",
			boundary: "DeveloperBoundary",
			expected: "arn:aws:iam::123456789012:policy/DeveloperBoundary",
		},
		{
			name:     "boundary policy ARN",
			boundary: "arn:aws:iam::123456789012:policy/CustomBoundary",
			expected: "arn:aws:iam::123456789012:policy/CustomBoundary",
		},
		{
			name:     "empty boundary",
			boundary: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup := mock.NewIAMTestSetup()
			setup.ApplyWithOverrides(t, map[string]string{
				"MONAD_BOUNDARY_POLICY": tt.boundary,
			})
			ctx := context.Background()

			config, err := Derive(ctx, setup.Basis)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, config.BoundaryPolicyArn())
			
			if tt.boundary != "" {
				if strings.HasPrefix(tt.boundary, "arn:") {
					assert.Equal(t, "CustomBoundary", config.BoundaryPolicyName())
				} else {
					assert.Equal(t, tt.boundary, config.BoundaryPolicyName())
				}
			}
		})
	}
}

func TestDerive_ENIRoleConfiguration(t *testing.T) {
	setup := mock.NewIAMTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)

	// Verify ENI role configuration
	assert.Equal(t, "AWSLambdaVPCAccessExecutionRole", config.EniRoleName())
	
	expectedArn := "arn:aws:iam::123456789012:role/AWSLambdaVPCAccessExecutionRole"
	assert.Equal(t, expectedArn, config.EniRoleArn())
	
	expectedPolicyArn := "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
	assert.Equal(t, expectedPolicyArn, config.EniRolePolicyArn())
}

func TestDerive_Tags(t *testing.T) {
	setup := mock.NewIAMTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)

	tags := config.Tags()
	assert.NotEmpty(t, tags)
	
	// Should contain basic resource tags
	foundService := false
	foundOwner := false
	for _, tag := range tags {
		if *tag.Key == "Service" {
			foundService = true
			assert.Equal(t, "test-service", *tag.Value)
		}
		if *tag.Key == "Owner" {
			foundOwner = true
			assert.Equal(t, "test-owner", *tag.Value)
		}
	}
	assert.True(t, foundService, "Should have Service tag")
	assert.True(t, foundOwner, "Should have Owner tag")
}

func TestValidate_Success(t *testing.T) {
	setup := mock.NewIAMTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)

	err = config.Validate()
	assert.NoError(t, err)
}

func TestValidate_Failures(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name: "missing client",
			config: &Config{
				client:            nil,
				IamPolicyDocument: "{}",
				IamRoleDocument:   "{}",
			},
		},
		{
			name: "missing policy document",
			config: &Config{
				IamPolicyDocument: "",
				IamRoleDocument:   "{}",
			},
		},
		{
			name: "missing role document", 
			config: &Config{
				IamPolicyDocument: "{}",
				IamRoleDocument:   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "cannot be blank")
		})
	}
}

func TestDerive_ErrorPropagation(t *testing.T) {
	errorSetup := mock.NewErrorTestSetup()
	errorSetup.Apply(t)
	ctx := context.Background()

	_, err := Derive(ctx, errorSetup.Basis)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock:")
}

func TestDerive_TemplateRenderingError(t *testing.T) {
	mockBasis := mock.NewMockBasisWithErrors()
	ctx := context.Background()

	_, err := Derive(ctx, mockBasis)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock:")
}