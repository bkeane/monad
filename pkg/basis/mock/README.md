# Mock Package for Basis Components

This package provides comprehensive mocks for all basis components, designed to be reusable across all test packages in the monad project.

## Quick Start

```go
import "github.com/bkeane/monad/pkg/basis/mock"

func TestMyComponent(t *testing.T) {
    setup := mock.NewTestSetup()
    setup.Apply(t)
    
    config, err := config.Derive(ctx, setup.Basis)
    // Now you can test with realistic mock data
}
```

## Available Mocks

### Individual Component Mocks
- `NewMockCaller()` - AWS caller with account/region
- `NewMockGit()` - Git repository information  
- `NewMockService()` - Service name configuration
- `NewMockResource()` - Resource naming and tagging
- `NewMockRegistry()` - ECR registry configuration
- `NewMockDefaults()` - Template defaults

### Complete Basis Mock
- `NewMockBasis()` - Complete mock with all components
- `NewMockBasisWithOptions(opts)` - Customizable mock
- `NewMockBasisWithErrors()` - Mock that returns errors

## Customization

```go
opts := mock.BasisOptions{
    Owner:     "my-org", 
    Repo:      "my-repo",
    Branch:    "feature-branch",
    Service:   "my-service",
    AccountId: "555666777888",
    Region:    "eu-west-1",
}

setup := mock.NewTestSetupWithOptions(opts)
```

## Test Environment Setup

The `TestSetup` type handles both mock basis creation and environment variable configuration:

```go
setup := mock.NewTestSetup()
setup.Apply(t) // Sets up environment variables

// Or with overrides
setup.ApplyWithOverrides(t, map[string]string{
    "MONAD_MEMORY": "256",
    "CUSTOM_VAR":   "custom-value",
})
```

## Preset Configurations

For common testing scenarios:

```go
// Optimized for Lambda testing
setup := mock.NewLambdaTestSetup()

// Optimized for API Gateway testing  
setup := mock.NewAPIGatewayTestSetup()

// For testing error scenarios
errorSetup := mock.NewErrorTestSetup()
```

## Call Tracking

All mocks track method calls for caching behavior testing:

```go
mockBasis := mock.NewMockBasis()

// Make some calls
mockBasis.Git()
mockBasis.Caller()

// Verify call counts
assert.Equal(t, 1, mockBasis.GetCallCount("Git"))
assert.Equal(t, 1, mockBasis.GetCallCount("Caller"))

// Reset counters
mockBasis.ResetCallCounts()
```

## Template Rendering

The mock basis provides realistic template rendering:

```go
mockBasis := mock.NewMockBasis()

result, err := mockBasis.Render("{{.Git.Owner}}/{{.Git.Repo}}/{{.Service.Name}}")
// Returns: "test-owner/test-repo/test-service"

result, err := mockBasis.Render("ANY /{{.Git.Repo}}/{{.Git.Branch}}/{{.Service.Name}}/{proxy+}")
// Returns: "ANY /test-repo/test-branch/test-service/{proxy+}"
```

## Usage in Different Packages

### Config Package Testing

```go
// pkg/config/config_test.go
import "github.com/bkeane/monad/pkg/basis/mock"

func TestConfig_Lambda_Caching(t *testing.T) {
    setup := mock.NewLambdaTestSetup()
    setup.Apply(t)
    
    config, err := Derive(ctx, setup.Basis)
    require.NoError(t, err)
    
    // Test caching behavior
    lambda1, err := config.Lambda(ctx)
    lambda2, err := config.Lambda(ctx)
    assert.Same(t, lambda1, lambda2)
    
    // Verify basis methods called only once
    assert.Equal(t, 1, setup.Basis.GetCallCount("Caller"))
}
```

### Step Package Testing

```go
// pkg/step/lambda/lambda_test.go
import "github.com/bkeane/monad/pkg/basis/mock"

func TestLambdaStep_Deploy(t *testing.T) {
    setup := mock.NewLambdaTestSetup()
    setup.Apply(t)
    
    // Create step with mocked dependencies
    step := NewLambdaStep(setup.Basis)
    
    // Test deployment logic
    err := step.Deploy(ctx)
    require.NoError(t, err)
}
```

## Benefits

✅ **Realistic Data**: Mocks return data that passes validation  
✅ **Consistent**: Same mock data across all tests  
✅ **Customizable**: Easy to modify for specific test scenarios  
✅ **Call Tracking**: Built-in support for testing caching behavior  
✅ **Template Support**: Working template rendering with realistic data  
✅ **Environment Setup**: Automated environment variable configuration  
✅ **Reusable**: Share mocks across all packages  

## Error Testing

For testing error scenarios:

```go
errorSetup := mock.NewErrorTestSetup()
errorSetup.Apply(t)

// All basis methods will return errors
_, err := errorSetup.Basis.Git()
assert.Error(t, err)
assert.Equal(t, mock.ErrGitNotFound, err)
```

This mock package solves the fundamental testing challenge by providing realistic, reusable mocks that work across the entire codebase.