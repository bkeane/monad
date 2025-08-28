package caller

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockSTSClient implements the STSClient interface for testing
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
	// Create mock client
	mockClient := new(MockSTSClient)
	
	// Set up expectations
	expectedAccount := "123456789012"
	expectedArn := "arn:aws:iam::123456789012:user/testuser"
	
	mockClient.On("GetCallerIdentity", mock.Anything, mock.Anything).Return(
		&sts.GetCallerIdentityOutput{
			Account: &expectedAccount,
			Arn:     &expectedArn,
		}, nil)

	// Test the function
	ctx := context.Background()
	basis, err := Derive(ctx, mockClient)

	// Assertions
	require.NoError(t, err)
	assert.NotNil(t, basis)
	assert.Equal(t, expectedAccount, basis.AccountId())
	assert.Equal(t, expectedArn, basis.Arn())
	assert.NotNil(t, basis.AwsConfig())

	// Verify mock was called
	mockClient.AssertExpectations(t)
}

func TestDerive_GetCallerIdentityError(t *testing.T) {
	// Create mock client
	mockClient := new(MockSTSClient)
	
	// Set up expectations for error
	expectedError := errors.New("AWS credentials not configured")
	mockClient.On("GetCallerIdentity", mock.Anything, mock.Anything).Return(nil, expectedError)

	// Test the function
	ctx := context.Background()
	basis, err := Derive(ctx, mockClient)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, basis)
	assert.Contains(t, err.Error(), "AWS credentials not configured")

	// Verify mock was called
	mockClient.AssertExpectations(t)
}

func TestDerive_ValidationError(t *testing.T) {
	// Create mock client that returns nil values (will fail validation)
	mockClient := new(MockSTSClient)
	
	mockClient.On("GetCallerIdentity", mock.Anything, mock.Anything).Return(
		&sts.GetCallerIdentityOutput{
			Account: nil,  // This will cause validation to fail
			Arn:     nil,  // This will cause validation to fail
		}, nil)

	// Test the function
	ctx := context.Background()
	basis, err := Derive(ctx, mockClient)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, basis)

	// Verify mock was called
	mockClient.AssertExpectations(t)
}

func TestDerive_BackwardCompatibility(t *testing.T) {
	// This test shows that the existing API still works
	// Note: This will make a real AWS call if run outside of CI/without mocks
	// In a real test suite, you'd skip this test in CI environments
	
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	
	// This should still work with the original API
	// (will fail if no AWS credentials, but that's expected)
	basis, err := Derive(ctx)
	
	// We can't assert success because we don't know if AWS creds are available
	// But we can verify the function signature still works
	if err != nil {
		// Expected in CI environments without AWS credentials
		t.Logf("Expected error in test environment: %v", err)
	} else {
		// If it succeeds, validate the result
		assert.NotNil(t, basis)
		assert.NotEmpty(t, basis.AccountId())
		assert.NotEmpty(t, basis.Arn())
	}
}

func TestBasis_Accessors(t *testing.T) {
	account := "123456789012"
	arn := "arn:aws:iam::123456789012:user/testuser"
	config := aws.Config{Region: "us-west-2"}

	basis := &Basis{
		CallerConfig:  &config,
		CallerAccount: &account,
		CallerArn:     &arn,
	}

	assert.Equal(t, config, basis.AwsConfig())
	assert.Equal(t, account, basis.AccountId())
	assert.Equal(t, arn, basis.Arn())
}

func TestBasis_Validate(t *testing.T) {
	account := "123456789012"
	arn := "arn:aws:iam::123456789012:user/testuser"
	config := aws.Config{}

	tests := []struct {
		name    string
		basis   *Basis
		wantErr bool
	}{
		{
			name: "valid basis",
			basis: &Basis{
				CallerConfig:  &config,
				CallerAccount: &account,
				CallerArn:     &arn,
			},
			wantErr: false,
		},
		{
			name: "missing config",
			basis: &Basis{
				CallerAccount: &account,
				CallerArn:     &arn,
			},
			wantErr: true,
		},
		{
			name: "missing account",
			basis: &Basis{
				CallerConfig: &config,
				CallerArn:    &arn,
			},
			wantErr: true,
		},
		{
			name: "missing arn",
			basis: &Basis{
				CallerConfig:  &config,
				CallerAccount: &account,
			},
			wantErr: true,
		},
		{
			name:    "all missing",
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

// Example of how to test specific AWS error scenarios
func TestDerive_AWSErrorScenarios(t *testing.T) {
	tests := []struct {
		name          string
		mockError     error
		expectedError string
	}{
		{
			name:          "access denied",
			mockError:     errors.New("AccessDenied: User is not authorized to perform: sts:GetCallerIdentity"),
			expectedError: "AccessDenied",
		},
		{
			name:          "network error",
			mockError:     errors.New("no such host"),
			expectedError: "no such host",
		},
		{
			name:          "invalid credentials",
			mockError:     errors.New("InvalidUserID.NotFound: The AWS Access Key Id you provided does not exist in our records"),
			expectedError: "InvalidUserID.NotFound",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockSTSClient)
			mockClient.On("GetCallerIdentity", mock.Anything, mock.Anything).Return(nil, tt.mockError)

			ctx := context.Background()
			basis, err := Derive(ctx, mockClient)

			assert.Error(t, err)
			assert.Nil(t, basis)
			assert.Contains(t, err.Error(), tt.expectedError)

			mockClient.AssertExpectations(t)
		})
	}
}