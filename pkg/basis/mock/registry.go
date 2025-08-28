package mock

import (
	"github.com/bkeane/monad/pkg/basis/registry"
)

// NewMockRegistry creates a realistic mock registry.Basis for testing
func NewMockRegistry() *registry.Basis {
	callerBasis := NewMockCaller()
	gitBasis := NewMockGit()
	serviceBasis := NewMockService()
	
	// Create a registry basis using the same logic as the real one
	registryBasis, _ := registry.Derive(callerBasis, gitBasis, serviceBasis)
	return registryBasis
}

// NewMockRegistryWithAccount creates a mock registry with specific account
func NewMockRegistryWithAccount(accountId string) *registry.Basis {
	callerBasis := NewMockCallerWithAccount(accountId)
	gitBasis := NewMockGit()
	serviceBasis := NewMockService()
	
	registryBasis, _ := registry.Derive(callerBasis, gitBasis, serviceBasis)
	return registryBasis
}

// NewMockRegistrySimple creates a simple mock registry without dependencies
func NewMockRegistrySimple() *registry.Basis {
	return &registry.Basis{
		EcrId:     "123456789012",
		EcrRegion: "us-east-1", 
		EcrImage:  "test-owner/test-repo/test-service:test-branch",
	}
}