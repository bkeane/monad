package mock

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/bkeane/monad/pkg/basis/caller"
	"github.com/stretchr/testify/mock"
)

// MockSTSClient provides a mock STS client for testing
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

// NewMockCaller creates a realistic mock caller.Basis for testing
func NewMockCaller() *caller.Basis {
	return &caller.Basis{
		CallerConfig: &aws.Config{
			Region: "us-east-1",
		},
		CallerAccount: aws.String("123456789012"),
		CallerArn:     aws.String("arn:aws:iam::123456789012:user/test-user"),
	}
}

// NewMockCallerWithRegion creates a mock caller with specific region
func NewMockCallerWithRegion(region string) *caller.Basis {
	return &caller.Basis{
		CallerConfig: &aws.Config{
			Region: region,
		},
		CallerAccount: aws.String("123456789012"),
		CallerArn:     aws.String("arn:aws:iam::123456789012:user/test-user"),
	}
}

// NewMockCallerWithAccount creates a mock caller with specific account ID
func NewMockCallerWithAccount(accountId string) *caller.Basis {
	return &caller.Basis{
		CallerConfig: &aws.Config{
			Region: "us-east-1",
		},
		CallerAccount: aws.String(accountId),
		CallerArn:     aws.String("arn:aws:iam::" + accountId + ":user/test-user"),
	}
}

// SetupMockSTSClient creates a mock STS client with realistic responses
func SetupMockSTSClient() *MockSTSClient {
	mockClient := new(MockSTSClient)
	
	account := "123456789012"
	arn := "arn:aws:iam::123456789012:user/test-user"
	
	mockClient.On("GetCallerIdentity", mock.Anything, mock.Anything).Return(
		&sts.GetCallerIdentityOutput{
			Account: &account,
			Arn:     &arn,
		}, nil)

	return mockClient
}

// SetupMockSTSClientWithError creates a mock STS client that returns errors
func SetupMockSTSClientWithError() *MockSTSClient {
	mockClient := new(MockSTSClient)
	mockClient.On("GetCallerIdentity", mock.Anything, mock.Anything).Return(nil, errors.New("mock STS error"))
	return mockClient
}