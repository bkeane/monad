package mock

import (
	"github.com/bkeane/monad/pkg/basis/resource"
)

// NewMockResource creates a realistic mock resource.Basis for testing
func NewMockResource() *resource.Basis {
	gitBasis := NewMockGit()
	serviceBasis := NewMockService()
	
	// Create a resource basis using the same logic as the real one
	resourceBasis, _ := resource.Derive(gitBasis, serviceBasis)
	return resourceBasis
}

// NewMockResourceWithName creates a mock resource with specific naming
func NewMockResourceWithName(owner, repo, branch, service string) *resource.Basis {
	gitBasis := NewMockGitWithRepo(owner, repo, branch)
	serviceBasis := NewMockServiceWithName(service)
	
	resourceBasis, _ := resource.Derive(gitBasis, serviceBasis)
	return resourceBasis
}

// NewMockResourceSimple creates a simple mock resource without dependencies
func NewMockResourceSimple() *resource.Basis {
	return &resource.Basis{
		ResourceName: "test-owner-test-repo-test-branch-test-service",
		ResourcePath: "test-owner/test-repo/test-branch/test-service",
		ResourceTags: map[string]string{
			"Owner":     "test-owner",
			"Repo":      "test-repo", 
			"Branch":    "test-branch",
			"Service":   "test-service",
			"ManagedBy": "monad",
		},
	}
}