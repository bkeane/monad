package mock

import (
	"testing"
)

// TestSetup provides complete test environment with mock basis and environment variables
type TestSetup struct {
	Basis       *MockBasis
	Environment map[string]string
	Options     BasisOptions
}

// NewTestSetup creates a test setup with default configuration
func NewTestSetup() *TestSetup {
	opts := DefaultBasisOptions()
	return &TestSetup{
		Basis:   NewMockBasisWithOptions(opts),
		Options: opts,
		Environment: map[string]string{
			// Git-related
			"MONAD_OWNER":  opts.Owner,
			"MONAD_REPO":   opts.Repo,
			"MONAD_BRANCH": opts.Branch,
			"MONAD_SHA":    "abc1234567890def",
			
			// Service-related
			"MONAD_SERVICE": opts.Service,
			
			// AWS-related
			"AWS_REGION": opts.Region,
			
			// Lambda configuration
			"MONAD_MEMORY":  "128",
			"MONAD_TIMEOUT": "3",
			"MONAD_STORAGE": "512",
			"MONAD_RETRIES": "0",
			
			// API Gateway configuration
			"MONAD_ROUTE": "ANY /test/{proxy+}",
			"MONAD_AUTH":  "aws_iam",
		},
	}
}

// NewTestSetupWithOptions creates a test setup with custom options
func NewTestSetupWithOptions(opts BasisOptions) *TestSetup {
	setup := NewTestSetup()
	setup.Options = opts
	setup.Basis = NewMockBasisWithOptions(opts)
	
	// Update environment to match options
	setup.Environment["MONAD_OWNER"] = opts.Owner
	setup.Environment["MONAD_REPO"] = opts.Repo
	setup.Environment["MONAD_BRANCH"] = opts.Branch
	setup.Environment["MONAD_SERVICE"] = opts.Service
	setup.Environment["AWS_REGION"] = opts.Region
	
	return setup
}

// Apply sets up the test environment by configuring environment variables
func (ts *TestSetup) Apply(t *testing.T) {
	for key, value := range ts.Environment {
		t.Setenv(key, value)
	}
}

// ApplyWithOverrides applies environment with additional overrides
func (ts *TestSetup) ApplyWithOverrides(t *testing.T, overrides map[string]string) {
	// Apply base environment
	ts.Apply(t)
	
	// Apply overrides
	for key, value := range overrides {
		t.Setenv(key, value)
	}
}

// WithEnvironment adds or overrides environment variables
func (ts *TestSetup) WithEnvironment(env map[string]string) *TestSetup {
	for key, value := range env {
		ts.Environment[key] = value
	}
	return ts
}

// WithBasisOptions updates the basis options and recreates the mock basis
func (ts *TestSetup) WithBasisOptions(opts BasisOptions) *TestSetup {
	ts.Options = opts
	ts.Basis = NewMockBasisWithOptions(opts)
	return ts
}

// Preset configurations for common testing scenarios

// NewLambdaTestSetup creates setup optimized for Lambda config testing
func NewLambdaTestSetup() *TestSetup {
	setup := NewTestSetup()
	setup.Environment["MONAD_LAMBDA_REGION"] = setup.Options.Region
	return setup
}

// NewAPIGatewayTestSetup creates setup optimized for API Gateway config testing
func NewAPIGatewayTestSetup() *TestSetup {
	setup := NewTestSetup()
	setup.Environment["MONAD_API_REGION"] = setup.Options.Region
	setup.Environment["MONAD_ROUTE"] = "ANY /test-repo/test-branch/test-service/{proxy+}"
	return setup
}

// NewIAMTestSetup creates setup optimized for IAM config testing
func NewIAMTestSetup() *TestSetup {
	setup := NewTestSetup()
	// IAM doesn't need special setup beyond defaults
	return setup
}

// Error setup for testing failure scenarios
type ErrorTestSetup struct {
	Basis       *MockBasisWithErrors
	Environment map[string]string
}

// NewErrorTestSetup creates a test setup that will cause basis errors
func NewErrorTestSetup() *ErrorTestSetup {
	return &ErrorTestSetup{
		Basis: NewMockBasisWithErrors(),
		Environment: map[string]string{
			"MONAD_SERVICE": "test-service",
		},
	}
}

// Apply sets up the error test environment
func (ets *ErrorTestSetup) Apply(t *testing.T) {
	for key, value := range ets.Environment {
		t.Setenv(key, value)
	}
}