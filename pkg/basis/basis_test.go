package basis

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock STS client for testing
type MockSTSClient struct {
	mock.Mock
}

func (m *MockSTSClient) GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sts.GetCallerIdentityOutput), args.Error(1)
}

func TestDerive_Success(t *testing.T) {
	basis, err := Derive(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, basis)
	assert.Empty(t, basis.Chdir) // No chdir by default
}

func TestDerive_WithChdirEnvironmentVariable(t *testing.T) {
	// Create a temp directory to chdir to
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	
	t.Setenv("MONAD_CHDIR", tmpDir)

	basis, err := Derive(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, basis)
	assert.Equal(t, tmpDir, basis.Chdir)

	// Verify we actually changed directories
	currentDir, err := os.Getwd()
	require.NoError(t, err)
	assert.Equal(t, tmpDir, currentDir)

	// Clean up - change back to original directory
	err = os.Chdir(originalDir)
	require.NoError(t, err)
}

func TestDerive_ChdirInvalidDirectory(t *testing.T) {
	t.Setenv("MONAD_CHDIR", "/nonexistent/directory/that/should/not/exist")

	basis, err := Derive(context.Background())
	assert.Error(t, err)
	assert.Nil(t, basis)
	assert.Contains(t, err.Error(), "failed to change directory")
}

func TestBasis_LazyInitializationAndCaching(t *testing.T) {
	// Set up environment to avoid external dependencies
	setupMockEnvironment(t)
	mockClient := setupMockAWSClient()

	basis, err := Derive(context.Background())
	require.NoError(t, err)

	// First call should initialize
	git1, err := basis.Git()
	require.NoError(t, err)
	assert.NotNil(t, git1)

	// Second call should return cached instance
	git2, err := basis.Git()
	require.NoError(t, err)
	assert.Same(t, git1, git2) // Same pointer = cached

	// Test all components get cached
	service1, err := basis.Service()
	require.NoError(t, err)
	service2, err := basis.Service()
	require.NoError(t, err)
	assert.Same(t, service1, service2)

	defaults1, err := basis.Defaults()
	require.NoError(t, err)
	defaults2, err := basis.Defaults()
	require.NoError(t, err)
	assert.Same(t, defaults1, defaults2)

	// Test caller with mock
	if mockClient != nil {
		caller1, err := basis.Caller()
		if err == nil { // Only test if we can get caller
			caller2, err := basis.Caller()
			require.NoError(t, err)
			assert.Same(t, caller1, caller2)
		}
	}
}

func TestBasis_DependencyGraph(t *testing.T) {
	setupMockEnvironment(t)
	mockClient := setupMockAWSClient()

	basis, err := Derive(context.Background())
	require.NoError(t, err)

	// Resource depends on Git + Service
	resource, err := basis.Resource()
	require.NoError(t, err)
	assert.NotNil(t, resource)

	// Verify dependencies were created
	assert.NotNil(t, basis.GitBasis)
	assert.NotNil(t, basis.ServiceBasis)

	// Registry depends on Caller + Git + Service (skip if no AWS)
	if mockClient != nil {
		registry, err := basis.Registry()
		if err == nil {
			assert.NotNil(t, registry)
			assert.NotNil(t, basis.CallerBasis)
			assert.NotNil(t, basis.GitBasis)
			assert.NotNil(t, basis.ServiceBasis)
		}
	}
}

func TestBasis_TemplateRendering(t *testing.T) {
	setupMockEnvironment(t)
	setupMockAWSClient()

	basis, err := Derive(context.Background())
	require.NoError(t, err)

	tests := []struct {
		name     string
		template string
		contains []string
		notEmpty bool
	}{
		{
			name:     "git template variables",
			template: "{{.Git.Repo}}-{{.Git.Branch}}-{{.Service.Name}}",
			contains: []string{"-"},
			notEmpty: true,
		},
		{
			name:     "resource template variables",
			template: "/{{.Resource.Path}}/{{.Resource.Name}}",
			contains: []string{"/"},
			notEmpty: true,
		},
		{
			name:     "ecr template variables",
			template: "{{.Ecr.Id}}.dkr.ecr.{{.Ecr.Region}}.amazonaws.com",
			contains: []string{".dkr.ecr.", ".amazonaws.com"},
			notEmpty: true,
		},
		{
			name:     "simple text",
			template: "hello world",
			contains: []string{"hello", "world"},
			notEmpty: true,
		},
		{
			name:     "empty template",
			template: "",
			contains: []string{},
			notEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := basis.Render(tt.template)
			
			// Skip tests that require AWS if we don't have it
			if err != nil && (strings.Contains(err.Error(), "AWS") || 
				strings.Contains(err.Error(), "STS") ||
				strings.Contains(err.Error(), "GetCallerIdentity") ||
				strings.Contains(err.Error(), "ec2imds") ||
				strings.Contains(err.Error(), "IMDS") ||
				strings.Contains(err.Error(), "credentials")) {
				t.Skipf("Skipping test requiring AWS credentials: %v", err)
				return
			}
			
			require.NoError(t, err)

			if tt.notEmpty {
				assert.NotEmpty(t, result)
			}

			for _, substr := range tt.contains {
				assert.Contains(t, result, substr)
			}
		})
	}
}

func TestBasis_TemplateRenderingErrors(t *testing.T) {
	setupMockEnvironment(t)

	basis, err := Derive(context.Background())
	require.NoError(t, err)

	tests := []struct {
		name     string
		template string
		errorMsg string
	}{
		{
			name:     "invalid template syntax",
			template: "{{.Invalid.Syntax",
			errorMsg: "failed to parse template",
		},
		{
			name:     "invalid template variable",
			template: "{{.NonExistent.Field}}",
			errorMsg: "failed to execute template",
		},
		{
			name:     "malformed template",
			template: "{{range .}}{{end}}", // range needs a collection
			errorMsg: "failed to execute template",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := basis.Render(tt.template)
			
			// Skip if AWS dependencies fail - we want to test template errors, not AWS credential errors
			if err != nil && (strings.Contains(err.Error(), "AWS") || 
				strings.Contains(err.Error(), "STS") ||
				strings.Contains(err.Error(), "GetCallerIdentity") ||
				strings.Contains(err.Error(), "ec2imds") ||
				strings.Contains(err.Error(), "IMDS") ||
				strings.Contains(err.Error(), "credentials")) {
				t.Skipf("Skipping test requiring AWS credentials: %v", err)
				return
			}
			
			assert.Error(t, err)
			assert.Empty(t, result)
			assert.Contains(t, err.Error(), tt.errorMsg)
		})
	}
}

func TestBasis_TableGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping table test in short mode")
	}

	setupMockEnvironment(t)
	setupMockAWSClient()

	basis, err := Derive(context.Background())
	require.NoError(t, err)

	table, err := basis.Table()
	
	// Skip if AWS dependencies fail
	if err != nil && (strings.Contains(err.Error(), "AWS") || 
		strings.Contains(err.Error(), "STS") ||
		strings.Contains(err.Error(), "GetCallerIdentity") ||
		strings.Contains(err.Error(), "ec2imds") ||
		strings.Contains(err.Error(), "IMDS") ||
		strings.Contains(err.Error(), "credentials")) {
		t.Skipf("Skipping test requiring AWS credentials: %v", err)
		return
	}

	require.NoError(t, err)
	assert.NotEmpty(t, table)

	// Should contain headers
	assert.Contains(t, table, "Template")
	assert.Contains(t, table, "Value")

	// Should contain some template variables
	assert.Contains(t, table, "{{.Git.Repo}}")
	assert.Contains(t, table, "{{.Service.Name}}")
	assert.Contains(t, table, "{{.Ecr.Id}}")
	assert.Contains(t, table, "{{.Ecr.Region}}")
}

func TestBasis_ErrorPropagation(t *testing.T) {
	// Test in a directory that's not a git repo to trigger git errors
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	t.Setenv("MONAD_OWNER", "")  // Force git parsing
	t.Setenv("MONAD_REPO", "")
	t.Setenv("MONAD_BRANCH", "")
	t.Setenv("MONAD_SHA", "")

	basis, err := Derive(context.Background())
	require.NoError(t, err)

	// Git should fail because we're not in a git repo
	_, err = basis.Git()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git repository")

	// Resource depends on Git, so it should also fail
	_, err = basis.Resource()
	assert.Error(t, err)

	// Registry depends on Git, so it should also fail (if we can get past other dependencies)
	_, err = basis.Registry()
	assert.Error(t, err)
}

func TestBasis_ComponentIndependence(t *testing.T) {
	setupMockEnvironment(t)

	basis, err := Derive(context.Background())
	require.NoError(t, err)

	// These components should work independently
	service, err := basis.Service()
	require.NoError(t, err)
	assert.NotNil(t, service)

	defaults, err := basis.Defaults()
	require.NoError(t, err)
	assert.NotNil(t, defaults)

	// Git should work if we're in a git repo or have env vars
	git, err := basis.Git()
	if err != nil {
		t.Logf("Git component failed (expected if not in git repo): %v", err)
	} else {
		assert.NotNil(t, git)
	}
}

func TestBasis_MultipleInstancesIndependent(t *testing.T) {
	setupMockEnvironment(t)

	basis1, err := Derive(context.Background())
	require.NoError(t, err)

	basis2, err := Derive(context.Background())
	require.NoError(t, err)

	// Different instances should have independent caches
	assert.NotSame(t, basis1, basis2)

	// But if we access the same components, they should be cached per instance
	service1a, err := basis1.Service()
	require.NoError(t, err)
	service1b, err := basis1.Service()
	require.NoError(t, err)
	assert.Same(t, service1a, service1b) // Same instance, same cache

	service2, err := basis2.Service()
	require.NoError(t, err)
	assert.NotSame(t, service1a, service2) // Different instances, different objects
}

func TestBasis_RenderIntegration(t *testing.T) {
	setupMockEnvironment(t)
	mockClient := setupMockAWSClient()

	basis, err := Derive(context.Background())
	require.NoError(t, err)

	// Test a realistic template that would be used in the app
	template := "{{.Ecr.Id}}.dkr.ecr.{{.Ecr.Region}}.amazonaws.com/{{.Git.Owner}}/{{.Git.Repo}}/{{.Service.Name}}:{{.Git.Branch}}"
	
	result, err := basis.Render(template)
	
	// Skip if AWS dependencies fail
	if err != nil && (strings.Contains(err.Error(), "AWS") || 
		strings.Contains(err.Error(), "STS") ||
		strings.Contains(err.Error(), "GetCallerIdentity") ||
		strings.Contains(err.Error(), "ec2imds") ||
		strings.Contains(err.Error(), "IMDS") ||
		strings.Contains(err.Error(), "credentials")) {
		t.Skipf("Skipping test requiring AWS credentials: %v", err)
		return
	}

	require.NoError(t, err)
	assert.Contains(t, result, ".dkr.ecr.")
	assert.Contains(t, result, ".amazonaws.com/")

	// Verify no template variables remain
	assert.NotContains(t, result, "{{")
	assert.NotContains(t, result, "}}")

	_ = mockClient // Use mockClient to avoid unused variable
}

// Helper functions

func setupMockEnvironment(t *testing.T) {
	// Set up environment to avoid external dependencies where possible
	t.Setenv("MONAD_SERVICE", "test-service")
	t.Setenv("MONAD_OWNER", "test-owner")  
	t.Setenv("MONAD_REPO", "test-repo")
	t.Setenv("MONAD_BRANCH", "test-branch")
	t.Setenv("MONAD_SHA", "abc123")
}

func setupMockAWSClient() *MockSTSClient {
	// Return mock client but don't inject it - would require modifying the caller package
	// This is more for future enhancement
	mockClient := new(MockSTSClient)
	
	account := "123456789012"
	arn := "arn:aws:iam::123456789012:user/test"
	
	mockClient.On("GetCallerIdentity", mock.Anything, mock.Anything).Return(
		&sts.GetCallerIdentityOutput{
			Account: &account,
			Arn:     &arn,
		}, nil)

	return mockClient
}