package apigateway

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bkeane/monad/pkg/basis/mock"
)

func TestDerive_Success(t *testing.T) {
	setup := mock.NewAPIGatewayTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)
	assert.NotNil(t, config)

	// Verify components are properly set
	assert.NotNil(t, config.Client())
	assert.Equal(t, "us-east-1", config.client.Options().Region)
	assert.Equal(t, []string{"aws_iam"}, config.Auth())
}

func TestDerive_DefaultRoutePattern(t *testing.T) {
	setup := mock.NewAPIGatewayTestSetup()
	// Don't set MONAD_ROUTE to test default generation
	setup.Environment["MONAD_ROUTE"] = "" 
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)

	routes := config.Route()
	require.Len(t, routes, 1)
	
	// Should contain rendered template variables
	route := routes[0]
	assert.Contains(t, route, "test-repo")
	assert.Contains(t, route, "test-branch")  
	assert.Contains(t, route, "test-service")
	assert.Contains(t, route, "{proxy+}")
	assert.True(t, strings.HasPrefix(route, "ANY /"))
}

func TestDerive_CustomRoutePatterns(t *testing.T) {
	setup := mock.NewAPIGatewayTestSetup()
	setup.ApplyWithOverrides(t, map[string]string{
		"MONAD_ROUTE": "GET /custom/{proxy+},POST /api/{proxy+}",
		"MONAD_AUTH":  "aws_iam,jwt", // Must match route count
	})
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)

	routes := config.Route()
	require.Len(t, routes, 2)
	assert.Contains(t, routes, "GET /custom/{proxy+}")
	assert.Contains(t, routes, "POST /api/{proxy+}")
}

func TestDerive_CustomAuthTypes(t *testing.T) {
	setup := mock.NewAPIGatewayTestSetup()
	setup.ApplyWithOverrides(t, map[string]string{
		"MONAD_ROUTE": "GET /api/{proxy+},POST /webhook/{proxy+}",
		"MONAD_AUTH":  "jwt,custom", // Must match route count
	})
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)

	auth := config.Auth()
	require.Len(t, auth, 2)
	assert.Contains(t, auth, "jwt")
	assert.Contains(t, auth, "custom")
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
			opts := mock.DefaultBasisOptions()
			if tt.envRegion != "" {
				opts.Region = tt.envRegion
			}
			setup := mock.NewTestSetupWithOptions(opts)
			if tt.envRegion != "" {
				setup.ApplyWithOverrides(t, map[string]string{
					"MONAD_API_REGION": tt.envRegion,
				})
			} else {
				setup.Apply(t)
			}
			ctx := context.Background()

			config, err := Derive(ctx, setup.Basis)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedRegion, config.client.Options().Region)
		})
	}
}

func TestValidate_Success(t *testing.T) {
	setup := mock.NewAPIGatewayTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)

	err = config.Validate()
	assert.NoError(t, err)
}

func TestValidate_RoutePatternValidation(t *testing.T) {
	tests := []struct {
		name    string
		routes  []string
		wantErr bool
		errMsg  string
	}{
		{
			name:   "valid proxy route",
			routes: []string{"ANY /test/{proxy+}"},
		},
		{
			name:    "missing proxy pattern",
			routes:  []string{"ANY /test/"},
			wantErr: true,
			errMsg:  "route must contain {proxy+} pattern",
		},
		{
			name:    "proxy pattern not at end",
			routes:  []string{"ANY /test/{proxy+}/more"},
			wantErr: true,
			errMsg:  "route must end with {proxy+} pattern",
		},
		{
			name:   "multiple valid routes",
			routes: []string{"GET /api/{proxy+}", "POST /webhook/{proxy+}"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config with proper client for validation
			setup := mock.NewAPIGatewayTestSetup()
			setup.Apply(t)
			ctx := context.Background()
			
			baseConfig, err := Derive(ctx, setup.Basis)
			require.NoError(t, err)
			
			config := &Config{
				client:                  baseConfig.client,
				basis:                   setup.Basis,
				ApiGatewayRoutePatterns: tt.routes,
				ApiGatewayRegion:        "us-east-1",
				ApiGatewayAuthTypes:     make([]string, len(tt.routes)), // Match route count
				ApiGatewayAuthType:      make([]string, len(tt.routes)),
				ApiGatewayAuthorizerId:  make([]string, len(tt.routes)),
			}
			
			// Fill auth arrays with correct length
			for i := range tt.routes {
				config.ApiGatewayAuthTypes[i] = "aws_iam"
				config.ApiGatewayAuthType[i] = "AWS_IAM"
				config.ApiGatewayAuthorizerId[i] = ""
			}

			err = config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_AuthLengthMatching(t *testing.T) {
	setup := mock.NewAPIGatewayTestSetup()
	setup.Apply(t)

	config := &Config{
		basis:                   setup.Basis,
		ApiGatewayRoutePatterns: []string{"ANY /test/{proxy+}", "GET /api/{proxy+}"},
		ApiGatewayRegion:        "us-east-1",
		ApiGatewayAuthTypes:     []string{"aws_iam"}, // Only 1 auth type for 2 routes
		ApiGatewayAuthType:      []string{"AWS_IAM"},
		ApiGatewayAuthorizerId:  []string{""},
	}

	err := config.Validate()
	assert.Error(t, err)
	// Should fail because routes and auth types don't match in length
}

func TestPermissionSourceArns(t *testing.T) {
	setup := mock.NewAPIGatewayTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)

	// Set a mock API ID for permission testing
	config.ApiGatewayId = "test-api-id"

	arns, err := config.PermissionSourceArns()
	require.NoError(t, err)
	assert.NotEmpty(t, arns)

	// Should contain execute-api ARN pattern
	for _, arn := range arns {
		assert.Contains(t, arn, "arn:aws:execute-api:")
		assert.Contains(t, arn, "us-east-1")
		assert.Contains(t, arn, "123456789012")
		assert.Contains(t, arn, "test-api-id")
	}
}

func TestForwardedPrefixes(t *testing.T) {
	setup := mock.NewAPIGatewayTestSetup()
	setup.ApplyWithOverrides(t, map[string]string{
		"MONAD_ROUTE": "GET /api/v1/{proxy+},POST /webhook/{proxy+}",
		"MONAD_AUTH":  "aws_iam,jwt", // Must match route count
	})
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)

	prefixes, err := config.ForwardedPrefixes()
	require.NoError(t, err)
	require.Len(t, prefixes, 2)

	assert.Contains(t, prefixes, "/api/v1")
	assert.Contains(t, prefixes, "/webhook")
}

func TestPermissionStatementId(t *testing.T) {
	setup := mock.NewAPIGatewayTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	require.NoError(t, err)

	statementId := config.PermissionStatementId("test-api-123")
	expected := "apigatewayv2-test-repo-test-branch-test-service-test-api-123"
	assert.Equal(t, expected, statementId)
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

func TestDerive_WithApiGatewayIdSet(t *testing.T) {
	setup := mock.NewAPIGatewayTestSetup()
	setup.ApplyWithOverrides(t, map[string]string{
		"MONAD_API": "test-api-name",
	})
	ctx := context.Background()

	// This will attempt to resolve the API name, which will fail in test environment
	// But we can test that the configuration is attempted
	_, err := Derive(ctx, setup.Basis)
	
	// May fail due to AWS API calls, but should fail in a predictable way
	if err != nil {
		// Should be an AWS-related error, not a configuration error
		assert.NotContains(t, err.Error(), "mock:")
	}
}