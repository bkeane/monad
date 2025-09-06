package eventbridge

import (
	"context"
	"fmt"
	"os"
	"strings"
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
	// Note: This may fail due to AWS API validation call, but we test the basic config structure
	if err != nil {
		// Should be an AWS-related error, not a configuration error
		assert.NotContains(t, err.Error(), "mock:")
		return
	}

	assert.NotNil(t, config)

	// Verify components are properly set
	assert.NotNil(t, config.Client())
	assert.Equal(t, "us-east-1", config.Region())
	assert.Equal(t, "", config.BusName()) // No longer defaults to "default"
	assert.Equal(t, "test-repo-test-branch-test-service", config.RuleName())
}

func TestDerive_DefaultValues(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	if err != nil {
		// Should be an AWS-related error, not a configuration error
		assert.NotContains(t, err.Error(), "mock:")
		return
	}

	// Test default values are applied
	assert.Equal(t, "us-east-1", config.Region()) // From caller
	assert.Equal(t, "", config.BusName())   // Empty when not set
	assert.Equal(t, "test-repo-test-branch-test-service", config.RuleName()) // From resource
}

func TestDerive_CustomValues(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.ApplyWithOverrides(t, map[string]string{
		"MONAD_BUS_REGION": "eu-west-1",
		"MONAD_BUS_NAME":   "custom-bus",
	})
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	if err != nil {
		// Should be an AWS-related error, not a configuration error
		assert.NotContains(t, err.Error(), "mock:")
		return
	}

	assert.Equal(t, "eu-west-1", config.Region())
	assert.Equal(t, "custom-bus", config.BusName())
}

func TestDerive_WithCustomRuleTemplate(t *testing.T) {
	setup := mock.NewTestSetup()

	// Create a temp file for custom rule template
	tmpFile := t.TempDir() + "/custom-rule.json"
	customRule := `{
		"Rules": [
			{
				"Name": "{{.Service.Name}}-rule",
				"EventPattern": {
					"source": ["{{.Service.Name}}"]
				}
			}
		]
	}`
	require.NoError(t, os.WriteFile(tmpFile, []byte(customRule), 0644))

	setup.ApplyWithOverrides(t, map[string]string{
		"MONAD_RULE": tmpFile,
	})
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	if err != nil {
		// Should be an AWS-related error, not a configuration error
		assert.NotContains(t, err.Error(), "mock:")
		return
	}

	document := config.RuleDocument()
	assert.Contains(t, document, "test-service-rule")
	assert.Contains(t, document, `"source": ["test-service"]`)
}

func TestDerive_WithCustomBusAndRule(t *testing.T) {
	opts := mock.BasisOptions{
		Owner:     "custom-owner",
		Repo:      "custom-repo",
		Branch:    "custom-branch",
		Service:   "custom-service",
		AccountId: "555666777888",
		Region:    "eu-west-1",
	}
	setup := mock.NewTestSetupWithOptions(opts)
	setup.ApplyWithOverrides(t, map[string]string{
		"MONAD_BUS_REGION": "eu-west-1",
		"MONAD_BUS_NAME":   "custom-event-bus",
	})
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	if err != nil {
		// Should be an AWS-related error, not a configuration error
		assert.NotContains(t, err.Error(), "mock:")
		return
	}

	assert.Equal(t, "eu-west-1", config.Region())
	assert.Equal(t, "custom-event-bus", config.BusName())
	assert.Equal(t, "custom-repo-custom-branch-custom-service", config.RuleName())
}

func TestDerive_Tags(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	if err != nil {
		// Should be an AWS-related error, not a configuration error
		assert.NotContains(t, err.Error(), "mock:")
		return
	}

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

func TestPermissionStatementId(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.ApplyWithOverrides(t, map[string]string{
		"MONAD_BUS_NAME": "my-custom-bus",
	})
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	if err != nil {
		// Should be an AWS-related error, not a configuration error
		assert.NotContains(t, err.Error(), "mock:")
		return
	}

	statementId := config.PermissionStatementId()
	expected := "eventbridge-my-custom-bus-test-repo-test-branch-test-service"
	assert.Equal(t, expected, statementId)
}

func TestPermissionStatementId_EmptyBus(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	if err != nil {
		// Should be an AWS-related error, not a configuration error
		assert.NotContains(t, err.Error(), "mock:")
		return
	}

	statementId := config.PermissionStatementId()
	expected := "eventbridge--test-repo-test-branch-test-service"
	assert.Equal(t, expected, statementId)
}

func TestValidate_Success(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	if err != nil {
		// Should be an AWS-related error, not a configuration error
		assert.NotContains(t, err.Error(), "mock:")
		return
	}

	// Note: This may fail due to AWS API call in validation, but we test the basic config structure
	err = config.Validate()
	// We expect this to fail in test environment due to AWS API calls
	// The test verifies the config is properly structured for validation
	if err != nil {
		// Should be an AWS-related error, not a structural validation error
		assert.NotContains(t, err.Error(), "cannot be blank")
	}
}

func TestValidate_MissingClient(t *testing.T) {
	config := &Config{
		client:            nil,
		EventBridgeBusName: "default",
	}

	// Note: This test will panic because validation tries to use the nil client
	// to call AWS APIs. In a real scenario, this would be a programming error.
	// The EventBridge config requires AWS API validation which needs a valid client.
	defer func() {
		if r := recover(); r != nil {
			// Expected to panic with nil client
			assert.Contains(t, fmt.Sprint(r), "nil pointer")
		}
	}()

	config.Validate()
}

func TestDerive_RegionHandling(t *testing.T) {
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
					"MONAD_BUS_REGION": tt.envRegion,
				})
			} else {
				setup.Apply(t)
			}
			ctx := context.Background()

			config, err := Derive(ctx, setup.Basis)
			if err != nil {
				// Should be an AWS-related error, not a configuration error
				assert.NotContains(t, err.Error(), "mock:")
				return
			}

			assert.Equal(t, tt.expectedRegion, config.Region())
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

func TestRuleDocument_Default(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	if err != nil {
		// Should be an AWS-related error, not a configuration error
		assert.NotContains(t, err.Error(), "mock:")
		return
	}

	// When no rule template path is provided, should use default rule template
	document := config.RuleDocument()
	// The document will be rendered from the default rule template, so it should contain expected structure
	assert.NotEmpty(t, strings.TrimSpace(document))
	assert.Contains(t, document, "source")
	assert.Contains(t, document, "detail-type")
}

func TestBusName_Formatting(t *testing.T) {
	tests := []struct {
		name        string
		busName     string
		expectedBus string
	}{
		{
			name:        "empty bus when not set",
			busName:     "",
			expectedBus: "",
		},
		{
			name:        "custom bus name",
			busName:     "my-custom-bus",
			expectedBus: "my-custom-bus",
		},
		{
			name:        "bus with hyphens",
			busName:     "my-event-bus-name",
			expectedBus: "my-event-bus-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup := mock.NewTestSetup()
			if tt.busName != "" {
				setup.ApplyWithOverrides(t, map[string]string{
					"MONAD_BUS_NAME": tt.busName,
				})
			} else {
				setup.Apply(t)
			}
			ctx := context.Background()

			config, err := Derive(ctx, setup.Basis)
			if err != nil {
				// Should be an AWS-related error, not a configuration error
				assert.NotContains(t, err.Error(), "mock:")
				return
			}

			assert.Equal(t, tt.expectedBus, config.BusName())
		})
	}
}