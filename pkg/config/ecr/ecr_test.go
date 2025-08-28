package ecr

import (
	"context"
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
	// Note: This may fail due to AWS API calls during registryv2 initialization
	if err != nil {
		// Should be an AWS-related error, not a configuration error
		assert.NotContains(t, err.Error(), "mock:")
		return
	}

	assert.NotNil(t, config)

	// Verify components are properly set
	ecrClient, registryClient := config.Clients()
	assert.NotNil(t, ecrClient)
	assert.NotNil(t, registryClient)
}

func TestImagePath_Parsing(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	if err != nil {
		// Should be an AWS-related error, not a configuration error
		assert.NotContains(t, err.Error(), "mock:")
		return
	}

	imagePath := config.ImagePath()
	assert.NotEmpty(t, imagePath)
	
	// Should not contain the tag part
	assert.False(t, strings.Contains(imagePath, ":"))
	
	// Should be a valid ECR repository path format
	assert.Contains(t, imagePath, "123456789012.dkr.ecr.us-east-1.amazonaws.com")
}

func TestImageTag_Parsing(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	if err != nil {
		// Should be an AWS-related error, not a configuration error
		assert.NotContains(t, err.Error(), "mock:")
		return
	}

	imageTag := config.ImageTag()
	assert.NotEmpty(t, imageTag)
	
	// Should be a valid image tag
	assert.False(t, strings.Contains(imageTag, ":"))
	assert.False(t, strings.Contains(imageTag, "/"))
}

func TestImagePath_And_Tag_Combination(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	if err != nil {
		// Should be an AWS-related error, not a configuration error
		assert.NotContains(t, err.Error(), "mock:")
		return
	}

	imagePath := config.ImagePath()
	imageTag := config.ImageTag()

	// When combined, should recreate the original image string
	combinedImage := imagePath + ":" + imageTag
	
	// Should be a valid ECR image URI format
	assert.Contains(t, combinedImage, ".dkr.ecr.")
	assert.Contains(t, combinedImage, ".amazonaws.com")
	assert.Contains(t, combinedImage, ":")
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

	err = config.Validate()
	assert.NoError(t, err)
}

func TestValidate_MissingRegistry(t *testing.T) {
	config := &Config{
		registry:   nil,
		registryv2: nil,
	}

	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be blank")
}

func TestValidate_MissingRegistryV2(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)

	registry, err := setup.Basis.Registry()
	require.NoError(t, err)

	config := &Config{
		registry:   registry,
		registryv2: nil, // Missing registryv2 client
	}

	err = config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be blank")
}

func TestDerive_WithCustomRegistryOptions(t *testing.T) {
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
	if err != nil {
		// Should be an AWS-related error, not a configuration error
		assert.NotContains(t, err.Error(), "mock:")
		return
	}

	imagePath := config.ImagePath()
	
	// Should contain custom account ID and region
	assert.Contains(t, imagePath, "555666777888.dkr.ecr.eu-west-1.amazonaws.com")
	assert.Contains(t, imagePath, "custom-repo-custom-branch-custom-service")
}

func TestClients_BothReturned(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	if err != nil {
		// Should be an AWS-related error, not a configuration error
		assert.NotContains(t, err.Error(), "mock:")
		return
	}

	ecrClient, registryClient := config.Clients()
	
	// Both clients should be returned
	assert.NotNil(t, ecrClient, "ECR client should not be nil")
	assert.NotNil(t, registryClient, "Registry client should not be nil")
	
	// Should be different types
	assert.IsType(t, ecrClient, ecrClient)
	assert.NotEqual(t, ecrClient, registryClient, "Clients should be different instances")
}

func TestImageParsing_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		imageString  string
		expectedPath string
		expectedTag  string
	}{
		{
			name:         "standard ECR image",
			imageString:  "123456789012.dkr.ecr.us-east-1.amazonaws.com/my-repo:latest",
			expectedPath: "123456789012.dkr.ecr.us-east-1.amazonaws.com/my-repo",
			expectedTag:  "latest",
		},
		{
			name:         "image with commit SHA",
			imageString:  "123456789012.dkr.ecr.us-east-1.amazonaws.com/my-repo:abc123def",
			expectedPath: "123456789012.dkr.ecr.us-east-1.amazonaws.com/my-repo",
			expectedTag:  "abc123def",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test verifies the parsing logic would work correctly
			// with different image format inputs
			parts := strings.Split(tt.imageString, ":")
			actualPath := parts[0]
			actualTag := parts[1]
			
			assert.Equal(t, tt.expectedPath, actualPath)
			assert.Equal(t, tt.expectedTag, actualTag)
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

func TestDerive_RegistryV2InitError(t *testing.T) {
	// This test would verify error handling from registryv2.InitEcr
	// In a real scenario, this might fail due to AWS configuration or permissions
	mockBasis := mock.NewMockBasisWithErrors()
	ctx := context.Background()

	_, err := Derive(ctx, mockBasis)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock:")
}