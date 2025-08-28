package mock

import "github.com/bkeane/monad/pkg/basis/git"

// NewMockGit creates a realistic mock git.Basis for testing
func NewMockGit() *git.Basis {
	return &git.Basis{
		GitOwner:  "test-owner",
		GitRepo:   "test-repo", 
		GitBranch: "test-branch",
		GitSha:    "abc1234567890def",
	}
}

// NewMockGitWithRepo creates a mock git basis with specific repository details
func NewMockGitWithRepo(owner, repo, branch string) *git.Basis {
	return &git.Basis{
		GitOwner:  owner,
		GitRepo:   repo,
		GitBranch: branch,
		GitSha:    "abc1234567890def",
	}
}

// NewMockGitWithSha creates a mock git basis with specific commit SHA
func NewMockGitWithSha(sha string) *git.Basis {
	return &git.Basis{
		GitOwner:  "test-owner",
		GitRepo:   "test-repo",
		GitBranch: "main",
		GitSha:    sha,
	}
}