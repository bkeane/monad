package mock

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMockBasis(t *testing.T) {
	mockBasis := NewMockBasis()
	
	// Test all components can be retrieved without error
	git, err := mockBasis.Git()
	require.NoError(t, err)
	assert.Equal(t, "test-owner", git.Owner())
	assert.Equal(t, "test-repo", git.Repo())
	assert.Equal(t, "test-branch", git.Branch())
	
	caller, err := mockBasis.Caller()
	require.NoError(t, err)
	assert.Equal(t, "123456789012", caller.AccountId())
	assert.Equal(t, "us-east-1", caller.AwsConfig().Region)
	
	service, err := mockBasis.Service()
	require.NoError(t, err)
	assert.Equal(t, "test-service", service.Name())
	
	resource, err := mockBasis.Resource()
	require.NoError(t, err)
	assert.Contains(t, resource.Name(), "test-repo")
	assert.Contains(t, resource.Name(), "test-branch") 
	assert.Contains(t, resource.Name(), "test-service")
	
	defaults, err := mockBasis.Defaults()
	require.NoError(t, err)
	assert.NotEmpty(t, defaults.EnvTemplate())
}

func TestMockBasisWithOptions(t *testing.T) {
	opts := BasisOptions{
		Owner:     "custom-owner",
		Repo:      "custom-repo",
		Branch:    "custom-branch", 
		Service:   "custom-service",
		AccountId: "999888777666",
		Region:    "eu-west-1",
	}
	
	mockBasis := NewMockBasisWithOptions(opts)
	
	git, err := mockBasis.Git()
	require.NoError(t, err)
	assert.Equal(t, "custom-owner", git.Owner())
	assert.Equal(t, "custom-repo", git.Repo())
	assert.Equal(t, "custom-branch", git.Branch())
	
	caller, err := mockBasis.Caller()
	require.NoError(t, err)
	assert.Equal(t, "999888777666", caller.AccountId())
	assert.Equal(t, "eu-west-1", caller.AwsConfig().Region)
	
	service, err := mockBasis.Service()
	require.NoError(t, err)
	assert.Equal(t, "custom-service", service.Name())
}

func TestMockBasisRender(t *testing.T) {
	mockBasis := NewMockBasis()
	
	tests := []struct {
		name     string
		template string
		contains []string
	}{
		{
			name:     "git template variables",
			template: "{{.Git.Owner}}/{{.Git.Repo}}/{{.Git.Branch}}",
			contains: []string{"test-owner", "test-repo", "test-branch"},
		},
		{
			name:     "account template variables", 
			template: "{{.Account.Id}}:{{.Account.Region}}",
			contains: []string{"123456789012", "us-east-1"},
		},
		{
			name:     "service template variables",
			template: "service-{{.Service.Name}}",
			contains: []string{"service-test-service"},
		},
		{
			name:     "api gateway route template",
			template: "ANY /{{.Git.Repo}}/{{.Git.Branch}}/{{.Service.Name}}/{proxy+}",
			contains: []string{"ANY", "/test-repo/test-branch/test-service/", "{proxy+}"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mockBasis.Render(tt.template)
			require.NoError(t, err)
			assert.NotEmpty(t, result)
			
			for _, substr := range tt.contains {
				assert.Contains(t, result, substr, "Template result should contain %s", substr)
			}
		})
	}
}

func TestMockBasisCallTracking(t *testing.T) {
	mockBasis := NewMockBasis()
	
	// Initially no calls
	assert.Equal(t, 0, mockBasis.GetCallCount("Git"))
	assert.Equal(t, 0, mockBasis.GetCallCount("Caller"))
	
	// Make calls and verify tracking
	_, err := mockBasis.Git()
	require.NoError(t, err)
	assert.Equal(t, 1, mockBasis.GetCallCount("Git"))
	
	_, err = mockBasis.Caller()
	require.NoError(t, err) 
	assert.Equal(t, 1, mockBasis.GetCallCount("Caller"))
	
	// Multiple calls increment counter
	_, err = mockBasis.Git()
	require.NoError(t, err)
	assert.Equal(t, 2, mockBasis.GetCallCount("Git"))
	
	// Reset works
	mockBasis.ResetCallCounts()
	assert.Equal(t, 0, mockBasis.GetCallCount("Git"))
	assert.Equal(t, 0, mockBasis.GetCallCount("Caller"))
}

func TestMockBasisWithErrors(t *testing.T) {
	mockBasis := NewMockBasisWithErrors()
	
	// All methods should return errors
	_, err := mockBasis.Git()
	assert.Error(t, err)
	assert.Equal(t, ErrGitNotFound, err)
	
	_, err = mockBasis.Caller()
	assert.Error(t, err)
	assert.Equal(t, ErrCallerNotFound, err)
	
	_, err = mockBasis.Service()
	assert.Error(t, err)
	assert.Equal(t, ErrServiceNotFound, err)
	
	_, err = mockBasis.Render("test")
	assert.Error(t, err)
	assert.Equal(t, ErrRenderFailed, err)
	
	// Call tracking still works
	assert.Equal(t, 1, mockBasis.GetCallCount("Git"))
	assert.Equal(t, 1, mockBasis.GetCallCount("Caller"))
	assert.Equal(t, 1, mockBasis.GetCallCount("Service"))
	assert.Equal(t, 1, mockBasis.GetCallCount("Render"))
}

func TestTestSetup(t *testing.T) {
	setup := NewTestSetup()
	
	// Verify setup has reasonable defaults
	assert.NotNil(t, setup.Basis)
	assert.NotEmpty(t, setup.Environment)
	assert.Equal(t, "test-owner", setup.Options.Owner)
	assert.Equal(t, "test-service", setup.Options.Service)
	
	// Test applying environment
	setup.Apply(t)
	
	// Test that the mock basis works after environment setup
	git, err := setup.Basis.Git()
	require.NoError(t, err)
	assert.Equal(t, "test-owner", git.Owner())
}

func TestTestSetupWithOverrides(t *testing.T) {
	setup := NewTestSetup()
	
	overrides := map[string]string{
		"MONAD_MEMORY": "256",
		"CUSTOM_VAR":   "custom-value",
	}
	
	setup.ApplyWithOverrides(t, overrides)
	
	// Verify base environment was applied
	git, err := setup.Basis.Git()
	require.NoError(t, err)
	assert.Equal(t, "test-owner", git.Owner())
}

func TestPresetConfigurations(t *testing.T) {
	t.Run("Lambda setup", func(t *testing.T) {
		setup := NewLambdaTestSetup()
		setup.Apply(t)
		
		assert.Contains(t, setup.Environment, "MONAD_LAMBDA_REGION")
		assert.Contains(t, setup.Environment, "MONAD_MEMORY")
		assert.Contains(t, setup.Environment, "MONAD_TIMEOUT")
	})
	
	t.Run("API Gateway setup", func(t *testing.T) {
		setup := NewAPIGatewayTestSetup()
		setup.Apply(t)
		
		assert.Contains(t, setup.Environment, "MONAD_API_REGION")
		assert.Contains(t, setup.Environment, "MONAD_ROUTE")
		route := setup.Environment["MONAD_ROUTE"]
		assert.Contains(t, route, "{proxy+}")
	})
	
	t.Run("Error setup", func(t *testing.T) {
		errorSetup := NewErrorTestSetup()
		errorSetup.Apply(t)
		
		_, err := errorSetup.Basis.Git()
		assert.Error(t, err)
	})
}

func TestMockSTSClient(t *testing.T) {
	mockClient := SetupMockSTSClient()
	
	result, err := mockClient.GetCallerIdentity(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "123456789012", *result.Account)
	assert.Contains(t, *result.Arn, "test-user")
	
	// Test error client
	errorClient := SetupMockSTSClientWithError()
	_, err = errorClient.GetCallerIdentity(context.Background(), nil)
	assert.Error(t, err)
}