package git

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDerive_WithEnvironmentVariables(t *testing.T) {
	// Set all environment variables
	t.Setenv("MONAD_OWNER", "testowner")
	t.Setenv("MONAD_REPO", "testrepo")
	t.Setenv("MONAD_BRANCH", "testbranch")
	t.Setenv("MONAD_SHA", "abcd1234")

	basis, err := Derive()
	require.NoError(t, err)
	assert.NotNil(t, basis)

	// Should use env vars, not parse git
	assert.Equal(t, "testowner", basis.Owner())
	assert.Equal(t, "testrepo", basis.Repo())
	assert.Equal(t, "testbranch", basis.Branch())
	assert.Equal(t, "abcd1234", basis.Sha())
}

func TestDerive_WithPartialEnvironmentVariables(t *testing.T) {
	// Only set some environment variables
	t.Setenv("MONAD_OWNER", "envowner")
	t.Setenv("MONAD_REPO", "envrepo")
	// Leave MONAD_BRANCH and MONAD_SHA unset

	// This test assumes we're in a valid git repository
	// If not, we'll skip the test
	basis, err := Derive()
	if err != nil {
		t.Skipf("Skipping test: not in a valid git repository: %v", err)
		return
	}

	assert.NotNil(t, basis)
	// Should use env vars for owner/repo
	assert.Equal(t, "envowner", basis.Owner())
	assert.Equal(t, "envrepo", basis.Repo())
	// Should parse branch/sha from git
	assert.NotEmpty(t, basis.Branch())
	assert.NotEmpty(t, basis.Sha())
}

func TestDerive_FromGitRepository(t *testing.T) {
	// Clear all environment variables
	t.Setenv("MONAD_OWNER", "")
	t.Setenv("MONAD_REPO", "")
	t.Setenv("MONAD_BRANCH", "")
	t.Setenv("MONAD_SHA", "")

	// This test assumes we're in a valid git repository
	basis, err := Derive()
	if err != nil {
		t.Skipf("Skipping test: not in a valid git repository: %v", err)
		return
	}

	assert.NotNil(t, basis)
	// Should parse all values from git
	assert.NotEmpty(t, basis.Owner())
	assert.NotEmpty(t, basis.Repo())
	assert.NotEmpty(t, basis.Branch())
	assert.NotEmpty(t, basis.Sha())

	// Validate the values are reasonable
	assert.Greater(t, len(basis.Sha()), 10, "SHA should be reasonably long")
	assert.NotContains(t, basis.Owner(), "/", "Owner should not contain slashes")
	assert.NotContains(t, basis.Repo(), "/", "Repo should not contain slashes")
}

func TestDerive_OutsideGitRepository(t *testing.T) {
	// Clear environment variables
	t.Setenv("MONAD_OWNER", "")
	t.Setenv("MONAD_REPO", "")
	t.Setenv("MONAD_BRANCH", "")
	t.Setenv("MONAD_SHA", "")

	// Create a temporary directory that's not a git repo
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	// Change to temp directory
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	// Should fail because we're not in a git repository
	basis, err := Derive()
	assert.Error(t, err)
	assert.Nil(t, basis)
	assert.Contains(t, err.Error(), "git repository")
}

func TestDerive_ValidationFailure(t *testing.T) {
	// Set only partial environment variables (missing required fields)
	t.Setenv("MONAD_OWNER", "testowner")
	t.Setenv("MONAD_REPO", "")
	t.Setenv("MONAD_BRANCH", "")
	t.Setenv("MONAD_SHA", "")

	// Create a temporary directory that's not a git repo
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	// Should fail validation because required fields are missing
	basis, err := Derive()
	assert.Error(t, err)
	assert.Nil(t, basis)
}

func TestBasis_Accessors(t *testing.T) {
	basis := &Basis{
		GitOwner:  "testowner",
		GitRepo:   "testrepo",
		GitBranch: "testbranch",
		GitSha:    "abcd1234",
	}

	assert.Equal(t, "testowner", basis.Owner())
	assert.Equal(t, "testrepo", basis.Repo())
	assert.Equal(t, "testbranch", basis.Branch())
	assert.Equal(t, "abcd1234", basis.Sha())
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
				GitOwner:  "owner",
				GitRepo:   "repo",
				GitBranch: "branch",
				GitSha:    "sha",
			},
			wantErr: false,
		},
		{
			name: "missing owner",
			basis: &Basis{
				GitRepo:   "repo",
				GitBranch: "branch",
				GitSha:    "sha",
			},
			wantErr: true,
		},
		{
			name: "missing repo",
			basis: &Basis{
				GitOwner:  "owner",
				GitBranch: "branch",
				GitSha:    "sha",
			},
			wantErr: true,
		},
		{
			name: "missing branch",
			basis: &Basis{
				GitOwner: "owner",
				GitRepo:  "repo",
				GitSha:   "sha",
			},
			wantErr: true,
		},
		{
			name: "missing sha",
			basis: &Basis{
				GitOwner:  "owner",
				GitRepo:   "repo",
				GitBranch: "branch",
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

func TestDerive_EnvironmentVariablePrecedence(t *testing.T) {
	// Test that environment variables take precedence over git parsing
	t.Setenv("MONAD_OWNER", "env_owner")
	// Don't set repo, branch, sha - let them be parsed from git

	basis, err := Derive()
	if err != nil {
		t.Skipf("Skipping test: not in a valid git repository: %v", err)
		return
	}

	// Owner should come from env var
	assert.Equal(t, "env_owner", basis.Owner())
	// Others should come from git parsing
	assert.NotEmpty(t, basis.Repo())
	assert.NotEmpty(t, basis.Branch())
	assert.NotEmpty(t, basis.Sha())
}

func TestDerive_EmptyEnvironmentVariables(t *testing.T) {
	// Test that empty string env vars are treated as unset
	t.Setenv("MONAD_OWNER", "")
	t.Setenv("MONAD_REPO", "nonempty")

	basis, err := Derive()
	if err != nil {
		t.Skipf("Skipping test: not in a valid git repository: %v", err)
		return
	}

	// Owner should be parsed from git (empty env var ignored)
	assert.NotEmpty(t, basis.Owner())
	assert.NotEqual(t, "", basis.Owner())
	// Repo should come from env var
	assert.Equal(t, "nonempty", basis.Repo())
}

// TestDerive_CurrentWorkingDirectory tests that the basis correctly
// captures the current working directory for git operations
func TestDerive_CurrentWorkingDirectory(t *testing.T) {
	t.Setenv("MONAD_OWNER", "testowner")
	t.Setenv("MONAD_REPO", "testrepo")
	t.Setenv("MONAD_BRANCH", "testbranch")
	t.Setenv("MONAD_SHA", "abcd1234")

	originalDir, err := os.Getwd()
	require.NoError(t, err)

	basis, err := Derive()
	require.NoError(t, err)

	// The cwd field should be set to current directory
	assert.Equal(t, originalDir, basis.cwd)
}
