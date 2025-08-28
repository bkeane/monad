package lambda

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bkeane/monad/pkg/basis/mock"
)

func TestDerive_Success(t *testing.T) {
	setup := mock.NewLambdaTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)
	assert.NotNil(t, config)

	// Verify components are properly set
	assert.NotNil(t, config.Client())
	assert.Equal(t, "us-east-1", config.Region())
	assert.Equal(t, int32(128), config.MemorySize())
	assert.Equal(t, int32(3), config.Timeout())
	assert.Equal(t, int32(512), config.EphemeralStorage())
	assert.Equal(t, int32(0), config.Retries())
	assert.Equal(t, "test-repo-test-branch-test-service", config.FunctionName())
}

func TestDerive_WithCustomValues(t *testing.T) {
	setup := mock.NewLambdaTestSetup()
	setup.ApplyWithOverrides(t, map[string]string{
		"MONAD_MEMORY":  "256",
		"MONAD_TIMEOUT": "10",
		"MONAD_STORAGE": "1024",
		"MONAD_RETRIES": "2",
	})
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)

	assert.Equal(t, int32(256), config.MemorySize())
	assert.Equal(t, int32(10), config.Timeout())
	assert.Equal(t, int32(1024), config.EphemeralStorage())
	assert.Equal(t, int32(2), config.Retries())
}

func TestDerive_WithRegionOverride(t *testing.T) {
	setup := mock.NewLambdaTestSetup()
	setup.ApplyWithOverrides(t, map[string]string{
		"MONAD_LAMBDA_REGION": "eu-west-1",
	})
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)

	assert.Equal(t, "eu-west-1", config.Region())
}

func TestDerive_EnvironmentVariables(t *testing.T) {
	setup := mock.NewLambdaTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)

	env := config.Env()
	assert.NotEmpty(t, env)
	
	// Should contain rendered environment variables from template
	found := false
	for key, value := range env {
		if key != "" && value != "" {
			found = true
			break
		}
	}
	assert.True(t, found, "Environment should contain at least one key-value pair")
}

func TestDerive_FunctionArn(t *testing.T) {
	setup := mock.NewLambdaTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)

	arn := config.FunctionArn()
	assert.Contains(t, arn, "arn:aws:lambda:")
	assert.Contains(t, arn, "us-east-1")
	assert.Contains(t, arn, "123456789012")
	assert.Contains(t, arn, "function:")
	assert.Contains(t, arn, "test-repo-test-branch-test-service")
}

func TestDerive_Tags(t *testing.T) {
	setup := mock.NewLambdaTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)

	tags := config.Tags()
	assert.NotEmpty(t, tags)
	
	// Should contain basic resource tags
	assert.Contains(t, tags, "Service")
	assert.Contains(t, tags, "Owner")
	assert.Contains(t, tags, "Repo")
	assert.Contains(t, tags, "Branch")
}

func TestDerive_WithCustomEnvironmentTemplate(t *testing.T) {
	setup := mock.NewLambdaTestSetup()
	
	// Create a temp file for custom env template
	tmpFile := t.TempDir() + "/custom.env"
	customEnv := "CUSTOM_VAR={{.Service.Name}}\nANOTHER_VAR={{.Account.Region}}"
	require.NoError(t, os.WriteFile(tmpFile, []byte(customEnv), 0644))
	
	setup.ApplyWithOverrides(t, map[string]string{
		"MONAD_ENV": tmpFile,
	})
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)

	env := config.Env()
	assert.Contains(t, env, "CUSTOM_VAR")
	assert.Contains(t, env, "ANOTHER_VAR")
	assert.Equal(t, "test-service", env["CUSTOM_VAR"])
	assert.Equal(t, "us-east-1", env["ANOTHER_VAR"])
}

func TestValidate_Success(t *testing.T) {
	setup := mock.NewLambdaTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)

	err = config.Validate()
	assert.NoError(t, err)
}

func TestValidate_Failures(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*testing.T) Basis
		errorMsg string
	}{
		{
			name: "nil caller",
			setup: func(t *testing.T) Basis {
				mockBasis := mock.NewMockBasis()
				// Return nil for caller to trigger validation error
				return &struct {
					*mock.MockBasis
				}{mockBasis}
			},
			errorMsg: "cannot be blank",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				client: nil, // This should cause validation to fail
			}
			
			err := config.Validate()
			assert.Error(t, err)
			if tt.errorMsg != "" {
				assert.Contains(t, err.Error(), tt.errorMsg)
			}
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
	// Create a basis that will fail during template rendering
	mockBasis := mock.NewMockBasisWithErrors()
	ctx := context.Background()

	_, err := Derive(ctx, mockBasis)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock:")
}

func TestDerive_WithZeroValues(t *testing.T) {
	setup := mock.NewLambdaTestSetup()
	setup.ApplyWithOverrides(t, map[string]string{
		"MONAD_MEMORY":  "0",
		"MONAD_TIMEOUT": "0", 
		"MONAD_STORAGE": "0",
		"MONAD_RETRIES": "0",
	})
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)

	// Should apply defaults for zero values
	assert.Equal(t, int32(128), config.MemorySize(), "Should default memory to 128")
	assert.Equal(t, int32(3), config.Timeout(), "Should default timeout to 3")
	assert.Equal(t, int32(512), config.EphemeralStorage(), "Should default storage to 512")
	assert.Equal(t, int32(0), config.Retries(), "Retries can be 0")
}