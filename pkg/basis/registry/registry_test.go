package registry

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/bkeane/monad/pkg/basis/caller"
	"github.com/bkeane/monad/pkg/basis/git"
	"github.com/bkeane/monad/pkg/basis/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDerive_Success(t *testing.T) {
	// Create mock dependencies
	callerBasis := createMockCaller("123456789012", "us-west-2")
	gitBasis := createMockGit("testowner", "testrepo", "main", "abcd1234")
	serviceBasis := createMockService("myservice")

	basis, err := Derive(callerBasis, gitBasis, serviceBasis)
	require.NoError(t, err)
	assert.NotNil(t, basis)

	// Verify computed values
	assert.Equal(t, "123456789012", basis.Id())
	assert.Equal(t, "us-west-2", basis.Region())
	assert.Equal(t, "testowner/testrepo/myservice:main", basis.Image())
}

func TestDerive_WithEnvironmentVariables(t *testing.T) {
	// Set environment variables
	t.Setenv("MONAD_IMAGE", "custom/image:v1.0.0")
	t.Setenv("MONAD_REGISTRY_ID", "999888777666")
	t.Setenv("MONAD_REGISTRY_REGION", "eu-west-1")

	callerBasis := createMockCaller("123456789012", "us-west-2")
	gitBasis := createMockGit("testowner", "testrepo", "main", "abcd1234")
	serviceBasis := createMockService("myservice")

	basis, err := Derive(callerBasis, gitBasis, serviceBasis)
	require.NoError(t, err)

	// Environment variables should override computed values
	assert.Equal(t, "999888777666", basis.Id())
	assert.Equal(t, "eu-west-1", basis.Region())
	assert.Equal(t, "custom/image:v1.0.0", basis.Image())
}

func TestDerive_PartialEnvironmentVariables(t *testing.T) {
	// Only set image, let ID and region be computed
	t.Setenv("MONAD_IMAGE", "custom/image:latest")
	t.Setenv("MONAD_REGISTRY_ID", "")
	t.Setenv("MONAD_REGISTRY_REGION", "")

	callerBasis := createMockCaller("123456789012", "us-west-2")
	gitBasis := createMockGit("testowner", "testrepo", "main", "abcd1234")
	serviceBasis := createMockService("myservice")

	basis, err := Derive(callerBasis, gitBasis, serviceBasis)
	require.NoError(t, err)

	// Image from env, others computed
	assert.Equal(t, "custom/image:latest", basis.Image())
	assert.Equal(t, "123456789012", basis.Id())
	assert.Equal(t, "us-west-2", basis.Region())
}

func TestDerive_ImageTagGeneration(t *testing.T) {
	tests := []struct {
		name        string
		envImage    string
		gitBranch   string
		expectedImg string
	}{
		{
			name:        "no tag provided, uses branch",
			envImage:    "owner/repo/service",
			gitBranch:   "develop",
			expectedImg: "owner/repo/service:develop",
		},
		{
			name:        "tag already provided, keeps as is",
			envImage:    "owner/repo/service:v1.2.3",
			gitBranch:   "main",
			expectedImg: "owner/repo/service:v1.2.3",
		},
		{
			name:        "empty env, uses computed image",
			envImage:    "",
			gitBranch:   "feature/new",
			expectedImg: "testowner/testrepo/myservice:feature/new",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envImage != "" {
				t.Setenv("MONAD_IMAGE", tt.envImage)
			} else {
				t.Setenv("MONAD_IMAGE", "")
			}

			callerBasis := createMockCaller("123456789012", "us-west-2")
			gitBasis := createMockGit("testowner", "testrepo", tt.gitBranch, "abcd1234")
			serviceBasis := createMockService("myservice")

			basis, err := Derive(callerBasis, gitBasis, serviceBasis)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedImg, basis.Image())
		})
	}
}

func TestDerive_NilDependencies(t *testing.T) {
	tests := []struct {
		name         string
		callerBasis  *caller.Basis
		gitBasis     *git.Basis
		serviceBasis *service.Basis
		expectedErr  string
	}{
		{
			name:         "nil caller",
			callerBasis:  nil,
			gitBasis:     createMockGit("owner", "repo", "main", "sha"),
			serviceBasis: createMockService("service"),
			expectedErr:  "caller basis must not be nil",
		},
		{
			name:         "nil git",
			callerBasis:  createMockCaller("123456789012", "us-west-2"),
			gitBasis:     nil,
			serviceBasis: createMockService("service"),
			expectedErr:  "git basis must not be nil",
		},
		{
			name:         "nil service",
			callerBasis:  createMockCaller("123456789012", "us-west-2"),
			gitBasis:     createMockGit("owner", "repo", "main", "sha"),
			serviceBasis: nil,
			expectedErr:  "service basis must not be nil",
		},
		{
			name:         "all nil",
			callerBasis:  nil,
			gitBasis:     nil,
			serviceBasis: nil,
			expectedErr:  "caller basis must not be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			basis, err := Derive(tt.callerBasis, tt.gitBasis, tt.serviceBasis)
			assert.Error(t, err)
			assert.Nil(t, basis)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestBasis_Accessors(t *testing.T) {
	basis := &Basis{
		EcrImage:  "my/image:tag",
		EcrId:     "123456789012",
		EcrRegion: "us-east-1",
	}

	assert.Equal(t, "my/image:tag", basis.Image())
	assert.Equal(t, "123456789012", basis.Id())
	assert.Equal(t, "us-east-1", basis.Region())
}

func TestBasis_Validate(t *testing.T) {
	tests := []struct {
		name    string
		basis   *Basis
		wantErr bool
	}{
		{
			name: "valid basis",
			basis: &Basis{
				EcrImage:  "my/image:tag",
				EcrId:     "123456789012",
				EcrRegion: "us-west-2",
			},
			wantErr: false,
		},
		{
			name: "missing image",
			basis: &Basis{
				EcrId:     "123456789012",
				EcrRegion: "us-west-2",
			},
			wantErr: true,
		},
		{
			name: "missing id",
			basis: &Basis{
				EcrImage:  "my/image:tag",
				EcrRegion: "us-west-2",
			},
			wantErr: true,
		},
		{
			name: "missing region",
			basis: &Basis{
				EcrImage: "my/image:tag",
				EcrId:    "123456789012",
			},
			wantErr: true,
		},
		{
			name: "empty image",
			basis: &Basis{
				EcrImage:  "",
				EcrId:     "123456789012",
				EcrRegion: "us-west-2",
			},
			wantErr: true,
		},
		{
			name: "empty id",
			basis: &Basis{
				EcrImage:  "my/image:tag",
				EcrId:     "",
				EcrRegion: "us-west-2",
			},
			wantErr: true,
		},
		{
			name: "empty region",
			basis: &Basis{
				EcrImage:  "my/image:tag",
				EcrId:     "123456789012",
				EcrRegion: "",
			},
			wantErr: true,
		},
		{
			name:    "all empty",
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

func TestDerive_ValidationCalled(t *testing.T) {
	// Test that Derive calls validation and fails when required fields are missing
	// We can force this by creating a scenario where computed values would be empty
	
	// Create caller with empty values that would fail validation
	config := aws.Config{}
	account := ""
	arn := "arn:aws:iam::123456789012:user/test"
	
	callerBasis := &caller.Basis{
		CallerConfig:  &config,
		CallerAccount: &account, // Empty account should cause ID to be empty
		CallerArn:     &arn,
	}
	
	gitBasis := createMockGit("owner", "repo", "main", "sha")
	serviceBasis := createMockService("service")

	basis, err := Derive(callerBasis, gitBasis, serviceBasis)
	assert.Error(t, err)
	assert.Nil(t, basis)
}

func TestDerive_EnvironmentVariablePrecedence(t *testing.T) {
	// Test that env vars take precedence over computed values
	t.Setenv("MONAD_REGISTRY_ID", "env-id")

	callerBasis := createMockCaller("computed-id", "us-west-2")
	gitBasis := createMockGit("owner", "repo", "main", "sha")
	serviceBasis := createMockService("service")

	basis, err := Derive(callerBasis, gitBasis, serviceBasis)
	require.NoError(t, err)

	// Should use env var, not computed value
	assert.Equal(t, "env-id", basis.Id())
}

func TestDerive_DefaultImageGeneration(t *testing.T) {
	// Test the default image generation logic
	t.Setenv("MONAD_IMAGE", "")

	callerBasis := createMockCaller("123456789012", "us-west-2")
	gitBasis := createMockGit("myorg", "myrepo", "develop", "abc123")
	serviceBasis := createMockService("api")

	basis, err := Derive(callerBasis, gitBasis, serviceBasis)
	require.NoError(t, err)

	// Should generate: owner/repo/service:branch
	assert.Equal(t, "myorg/myrepo/api:develop", basis.Image())
}

func TestDerive_ImageFormatVariations(t *testing.T) {
	tests := []struct {
		name           string
		owner          string
		repo           string
		service        string
		branch         string
		expectedImage  string
	}{
		{
			name:          "standard format",
			owner:         "company",
			repo:          "backend",
			service:       "users",
			branch:        "main",
			expectedImage: "company/backend/users:main",
		},
		{
			name:          "with special characters in branch",
			owner:         "org",
			repo:          "app",
			service:       "api",
			branch:        "feature/auth-v2",
			expectedImage: "org/app/api:feature/auth-v2",
		},
		{
			name:          "short names",
			owner:         "a",
			repo:          "b",
			service:       "c",
			branch:        "d",
			expectedImage: "a/b/c:d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("MONAD_IMAGE", "")

			callerBasis := createMockCaller("123456789012", "us-west-2")
			gitBasis := createMockGit(tt.owner, tt.repo, tt.branch, "sha")
			serviceBasis := createMockService(tt.service)

			basis, err := Derive(callerBasis, gitBasis, serviceBasis)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedImage, basis.Image())
		})
	}
}

// Helper functions for creating mock basis objects

func createMockCaller(accountId, region string) *caller.Basis {
	config := aws.Config{Region: region}
	return &caller.Basis{
		CallerConfig:  &config,
		CallerAccount: &accountId,
		CallerArn:     stringPtr("arn:aws:iam::" + accountId + ":user/test"),
	}
}

func createMockGit(owner, repo, branch, sha string) *git.Basis {
	return &git.Basis{
		GitOwner:  owner,
		GitRepo:   repo,
		GitBranch: branch,
		GitSha:    sha,
	}
}

func createMockService(name string) *service.Basis {
	return &service.Basis{
		ServiceName: name,
	}
}

func stringPtr(s string) *string {
	return &s
}