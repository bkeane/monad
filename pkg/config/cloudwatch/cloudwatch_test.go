package cloudwatch

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bkeane/monad/pkg/basis/mock"
)

func TestDerive_Success(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)
	assert.NotNil(t, config)

	// Verify components are properly set
	assert.NotNil(t, config.Client())
	assert.Equal(t, "/aws/lambda/test-repo-test-branch-test-service", config.Name())
	assert.Equal(t, "us-east-1", config.Region())
	assert.Equal(t, int32(14), config.Retention())
}

func TestDerive_LogGroupArn(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)

	arn := config.Arn()
	expected := "arn:aws:logs:us-east-1:123456789012:log-group:/aws/lambda/test-repo-test-branch-test-service"
	assert.Equal(t, expected, arn)
}

func TestDerive_LogGroupName(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)

	name := config.Name()
	expected := "/aws/lambda/test-repo-test-branch-test-service"
	assert.Equal(t, expected, name)
}

func TestDerive_Retention(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.ApplyWithOverrides(t, map[string]string{
		"MONAD_LOG_RETENTION": "30",
	})
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)

	assert.Equal(t, int32(30), config.Retention())
}

func TestDerive_Tags(t *testing.T) {
	setup := mock.NewTestSetup()
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
	
	assert.Equal(t, "test-service", tags["Service"])
	assert.Equal(t, "test-owner", tags["Owner"])
}

func TestValidate_Success(t *testing.T) {
	setup := mock.NewTestSetup()
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
				client:               nil,
				CloudWatchRegion:     "", // Will fail validation
				CloudWatchRetention:  0,  // Will fail validation
			},
		},
		{
			name: "missing region",
			config: &Config{
				CloudWatchRegion:     "", // Will fail validation
				CloudWatchRetention:  14,
			},
		},
		{
			name: "zero retention",
			config: &Config{
				CloudWatchRegion:     "us-east-1",
				CloudWatchRetention:  0, // Will fail validation
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

func TestDerive_WithCustomOptions(t *testing.T) {
	opts := mock.BasisOptions{
		Owner:     "custom-owner",
		Repo:      "custom-repo",
		Branch:    "custom-branch",
		Service:   "custom-service",
		AccountId: "555666777888",
		Region:    "eu-west-1",
	}
	setup := mock.NewTestSetupWithOptions(opts)
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)

	// Verify custom values are used
	assert.Equal(t, "/aws/lambda/custom-repo-custom-branch-custom-service", config.Name())
	
	arn := config.Arn()
	assert.Contains(t, arn, "555666777888")
	assert.Contains(t, arn, "eu-west-1")
	assert.Contains(t, arn, "custom-repo-custom-branch-custom-service")
	
	tags := config.Tags()
	assert.Equal(t, "custom-service", tags["Service"])
	assert.Equal(t, "custom-owner", tags["Owner"])
	assert.Equal(t, "custom-repo", tags["Repo"])
	assert.Equal(t, "custom-branch", tags["Branch"])
}

func TestDerive_ErrorPropagation(t *testing.T) {
	errorSetup := mock.NewErrorTestSetup()
	errorSetup.Apply(t)
	ctx := context.Background()

	_, err := Derive(ctx, errorSetup.Basis)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock:")
}

func TestLogGroupName_Format(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)

	name := config.Name()
	
	// Should follow AWS Lambda log group naming convention
	assert.Equal(t, "/aws/lambda/test-repo-test-branch-test-service", name,
		"Log group name should follow /aws/lambda/{function-name} format")
}

func TestRegionHandling(t *testing.T) {
	tests := []struct {
		name           string
		envRegion      string
		expectedRegion string
	}{
		{
			name:           "uses caller region when not set",
			envRegion:      "",
			expectedRegion: "us-east-1",
		},
		{
			name:           "uses custom region when set",
			envRegion:      "eu-west-1",
			expectedRegion: "eu-west-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup := mock.NewTestSetup()
			if tt.envRegion != "" {
				setup.ApplyWithOverrides(t, map[string]string{
					"MONAD_LOG_REGION": tt.envRegion,
				})
			} else {
				setup.Apply(t)
			}
			ctx := context.Background()

			config, err := Derive(ctx, setup.Basis)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedRegion, config.Region())
		})
	}
}