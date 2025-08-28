package resource

import (
	"fmt"
	"testing"

	"github.com/bkeane/monad/pkg/basis/git"
	"github.com/bkeane/monad/pkg/basis/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDerive_Success(t *testing.T) {
	gitBasis := createMockGit("myorg", "myrepo", "main", "abc123def")
	serviceBasis := createMockService("api")

	basis, err := Derive(gitBasis, serviceBasis)
	require.NoError(t, err)
	assert.NotNil(t, basis)

	// Verify computed values
	assert.Equal(t, "myrepo-main", basis.NamePrefix())
	assert.Equal(t, "myrepo-main-api", basis.Name())
	assert.Equal(t, "myrepo/main", basis.PathPrefix())
	assert.Equal(t, "myrepo/main/api", basis.Path())

	// Verify tags
	expectedTags := map[string]string{
		"Monad":   "true",
		"Service": "api",
		"Owner":   "myorg",
		"Repo":    "myrepo",
		"Branch":  "main",
		"Sha":     "abc123def",
	}
	assert.Equal(t, expectedTags, basis.Tags())
}

func TestDerive_NilDependencies(t *testing.T) {
	tests := []struct {
		name         string
		gitBasis     *git.Basis
		serviceBasis *service.Basis
		expectedErr  string
	}{
		{
			name:         "nil git",
			gitBasis:     nil,
			serviceBasis: createMockService("api"),
			expectedErr:  "git basis must not be nil",
		},
		{
			name:         "nil service",
			gitBasis:     createMockGit("owner", "repo", "branch", "sha"),
			serviceBasis: nil,
			expectedErr:  "service basis must not be nil",
		},
		{
			name:         "both nil",
			gitBasis:     nil,
			serviceBasis: nil,
			expectedErr:  "git basis must not be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			basis, err := Derive(tt.gitBasis, tt.serviceBasis)
			assert.Error(t, err)
			assert.Nil(t, basis)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestDerive_NamingConventions(t *testing.T) {
	tests := []struct {
		name            string
		gitOwner        string
		gitRepo         string
		gitBranch       string
		serviceName     string
		expectedPrefix  string
		expectedName    string
		expectedPath    string
	}{
		{
			name:           "standard naming",
			gitOwner:       "company",
			gitRepo:        "backend",
			gitBranch:      "main",
			serviceName:    "users",
			expectedPrefix: "backend-main",
			expectedName:   "backend-main-users",
			expectedPath:   "backend/main/users",
		},
		{
			name:           "feature branch",
			gitOwner:       "org",
			gitRepo:        "frontend",
			gitBranch:      "feature/auth",
			serviceName:    "login",
			expectedPrefix: "frontend-feature/auth",
			expectedName:   "frontend-feature/auth-login",
			expectedPath:   "frontend/feature/auth/login",
		},
		{
			name:           "development branch",
			gitOwner:       "team",
			gitRepo:        "services",
			gitBranch:      "develop",
			serviceName:    "payments",
			expectedPrefix: "services-develop",
			expectedName:   "services-develop-payments",
			expectedPath:   "services/develop/payments",
		},
		{
			name:           "single character components",
			gitOwner:       "a",
			gitRepo:        "b",
			gitBranch:      "c",
			serviceName:    "d",
			expectedPrefix: "b-c",
			expectedName:   "b-c-d",
			expectedPath:   "b/c/d",
		},
		{
			name:           "with hyphens",
			gitOwner:       "my-org",
			gitRepo:        "my-app",
			gitBranch:      "hotfix-v1",
			serviceName:    "user-service",
			expectedPrefix: "my-app-hotfix-v1",
			expectedName:   "my-app-hotfix-v1-user-service",
			expectedPath:   "my-app/hotfix-v1/user-service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gitBasis := createMockGit(tt.gitOwner, tt.gitRepo, tt.gitBranch, "sha123")
			serviceBasis := createMockService(tt.serviceName)

			basis, err := Derive(gitBasis, serviceBasis)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedPrefix, basis.NamePrefix())
			assert.Equal(t, tt.expectedName, basis.Name())
			assert.Equal(t, fmt.Sprintf("%s/%s", tt.gitRepo, tt.gitBranch), basis.PathPrefix())
			assert.Equal(t, tt.expectedPath, basis.Path())
		})
	}
}

func TestDerive_TagGeneration(t *testing.T) {
	gitBasis := createMockGit("testowner", "testapp", "staging", "def456abc")
	serviceBasis := createMockService("web")

	basis, err := Derive(gitBasis, serviceBasis)
	require.NoError(t, err)

	tags := basis.Tags()

	// Test all expected tags are present
	assert.Equal(t, "true", tags["Monad"])
	assert.Equal(t, "web", tags["Service"])
	assert.Equal(t, "testowner", tags["Owner"])
	assert.Equal(t, "testapp", tags["Repo"])
	assert.Equal(t, "staging", tags["Branch"])
	assert.Equal(t, "def456abc", tags["Sha"])

	// Test that all required tags exist
	expectedKeys := []string{"Monad", "Service", "Owner", "Repo", "Branch", "Sha"}
	for _, key := range expectedKeys {
		assert.Contains(t, tags, key, "Tags should contain key: %s", key)
		assert.NotEmpty(t, tags[key], "Tag value for %s should not be empty", key)
	}

	// Test total number of tags
	assert.Len(t, tags, 6)
}

func TestDerive_TagValues(t *testing.T) {
	tests := []struct {
		name        string
		owner       string
		repo        string
		branch      string
		sha         string
		serviceName string
	}{
		{
			name:        "standard values",
			owner:       "company",
			repo:        "project",
			branch:      "main",
			sha:         "abc123",
			serviceName: "api",
		},
		{
			name:        "with special characters",
			owner:       "my-org",
			repo:        "app.service",
			branch:      "feature/new-auth",
			sha:         "1a2b3c4d5e6f",
			serviceName: "user_service",
		},
		{
			name:        "long values",
			owner:       "very-long-organization-name",
			repo:        "extremely-long-repository-name",
			branch:      "very-long-feature-branch-name",
			sha:         "abcdef1234567890abcdef1234567890abcdef12",
			serviceName: "very-long-service-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gitBasis := createMockGit(tt.owner, tt.repo, tt.branch, tt.sha)
			serviceBasis := createMockService(tt.serviceName)

			basis, err := Derive(gitBasis, serviceBasis)
			require.NoError(t, err)

			tags := basis.Tags()
			assert.Equal(t, tt.owner, tags["Owner"])
			assert.Equal(t, tt.repo, tags["Repo"])
			assert.Equal(t, tt.branch, tags["Branch"])
			assert.Equal(t, tt.sha, tags["Sha"])
			assert.Equal(t, tt.serviceName, tags["Service"])
			assert.Equal(t, "true", tags["Monad"])
		})
	}
}

func TestBasis_Accessors(t *testing.T) {
	tags := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	basis := &Basis{
		ResourceNamePrefix: "test-prefix",
		ResourceName:       "test-name",
		ResourcePathPrefix: "test/path/prefix",
		ResourcePath:       "test/path/full",
		ResourceTags:       tags,
	}

	assert.Equal(t, "test-prefix", basis.NamePrefix())
	assert.Equal(t, "test-name", basis.Name())
	assert.Equal(t, "test/path/prefix", basis.PathPrefix())
	assert.Equal(t, "test/path/full", basis.Path())
	assert.Equal(t, tags, basis.Tags())
}

func TestBasis_Validate(t *testing.T) {
	validTags := map[string]string{"tag": "value"}

	tests := []struct {
		name    string
		basis   *Basis
		wantErr bool
	}{
		{
			name: "valid basis",
			basis: &Basis{
				ResourceNamePrefix: "prefix",
				ResourceName:       "name",
				ResourcePathPrefix: "path/prefix",
				ResourcePath:       "path/full",
				ResourceTags:       validTags,
			},
			wantErr: false,
		},
		{
			name: "missing name prefix",
			basis: &Basis{
				ResourceName:       "name",
				ResourcePathPrefix: "path/prefix",
				ResourcePath:       "path/full",
				ResourceTags:       validTags,
			},
			wantErr: true,
		},
		{
			name: "missing name",
			basis: &Basis{
				ResourceNamePrefix: "prefix",
				ResourcePathPrefix: "path/prefix",
				ResourcePath:       "path/full",
				ResourceTags:       validTags,
			},
			wantErr: true,
		},
		{
			name: "missing path prefix",
			basis: &Basis{
				ResourceNamePrefix: "prefix",
				ResourceName:       "name",
				ResourcePath:       "path/full",
				ResourceTags:       validTags,
			},
			wantErr: true,
		},
		{
			name: "missing path",
			basis: &Basis{
				ResourceNamePrefix: "prefix",
				ResourceName:       "name",
				ResourcePathPrefix: "path/prefix",
				ResourceTags:       validTags,
			},
			wantErr: true,
		},
		{
			name: "missing tags",
			basis: &Basis{
				ResourceNamePrefix: "prefix",
				ResourceName:       "name",
				ResourcePathPrefix: "path/prefix",
				ResourcePath:       "path/full",
			},
			wantErr: true,
		},
		{
			name: "empty name prefix",
			basis: &Basis{
				ResourceNamePrefix: "",
				ResourceName:       "name",
				ResourcePathPrefix: "path/prefix",
				ResourcePath:       "path/full",
				ResourceTags:       validTags,
			},
			wantErr: true,
		},
		{
			name: "empty tags map",
			basis: &Basis{
				ResourceNamePrefix: "prefix",
				ResourceName:       "name",
				ResourcePathPrefix: "path/prefix",
				ResourcePath:       "path/full",
				ResourceTags:       make(map[string]string),
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

func TestDerive_ValidationNotCalled(t *testing.T) {
	// Note: Unlike other packages, this Derive function doesn't call Validate()
	// This is intentional as all computed values will always be valid
	gitBasis := createMockGit("owner", "repo", "branch", "sha")
	serviceBasis := createMockService("service")

	basis, err := Derive(gitBasis, serviceBasis)
	require.NoError(t, err)
	assert.NotNil(t, basis)

	// Verify all fields are populated
	assert.NotEmpty(t, basis.NamePrefix())
	assert.NotEmpty(t, basis.Name())
	assert.NotEmpty(t, basis.PathPrefix())
	assert.NotEmpty(t, basis.Path())
	assert.NotEmpty(t, basis.Tags())

	// Verify validation would pass if called
	assert.NoError(t, basis.Validate())
}

func TestDerive_ConsistentOutput(t *testing.T) {
	// Test that multiple calls with same input produce identical output
	gitBasis := createMockGit("org", "app", "main", "abc123")
	serviceBasis := createMockService("api")

	basis1, err1 := Derive(gitBasis, serviceBasis)
	require.NoError(t, err1)

	basis2, err2 := Derive(gitBasis, serviceBasis)
	require.NoError(t, err2)

	// Should be identical
	assert.Equal(t, basis1.NamePrefix(), basis2.NamePrefix())
	assert.Equal(t, basis1.Name(), basis2.Name())
	assert.Equal(t, basis1.PathPrefix(), basis2.PathPrefix())
	assert.Equal(t, basis1.Path(), basis2.Path())
	assert.Equal(t, basis1.Tags(), basis2.Tags())
}

func TestDerive_EdgeCaseValues(t *testing.T) {
	tests := []struct {
		name        string
		owner       string
		repo        string
		branch      string
		sha         string
		serviceName string
		expectValid bool
	}{
		{
			name:        "empty strings produce valid structure",
			owner:       "",
			repo:        "",
			branch:      "",
			sha:         "",
			serviceName: "",
			expectValid: true, // Empty inputs still produce non-empty formatted strings like "-", "--", "/"
		},
		{
			name:        "spaces in values",
			owner:       "my org",
			repo:        "my repo",
			branch:      "my branch",
			sha:         "sha with spaces",
			serviceName: "my service",
			expectValid: true,
		},
		{
			name:        "unicode characters",
			owner:       "测试",
			repo:        "репо",
			branch:      "brañch",
			sha:         "ʃα",
			serviceName: "sërvice",
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gitBasis := createMockGit(tt.owner, tt.repo, tt.branch, tt.sha)
			serviceBasis := createMockService(tt.serviceName)

			basis, err := Derive(gitBasis, serviceBasis)
			require.NoError(t, err)
			assert.NotNil(t, basis)

			// Test if validation passes
			validationErr := basis.Validate()
			if tt.expectValid {
				assert.NoError(t, validationErr)
			} else {
				assert.Error(t, validationErr)
			}
		})
	}
}

// Helper functions for creating mock basis objects

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