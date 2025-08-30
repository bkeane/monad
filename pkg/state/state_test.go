package state

import (
	"testing"

	"github.com/bkeane/monad/pkg/basis/caller"
	"github.com/bkeane/monad/pkg/basis/git"
	"github.com/bkeane/monad/pkg/basis/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for testing
type MockBasis struct {
	mock.Mock
}

func (m *MockBasis) Caller() (*caller.Basis, error) {
	args := m.Called()
	return args.Get(0).(*caller.Basis), args.Error(1)
}

func (m *MockBasis) Git() (*git.Basis, error) {
	args := m.Called()
	return args.Get(0).(*git.Basis), args.Error(1)
}

func (m *MockBasis) Service() (*service.Basis, error) {
	args := m.Called()
	return args.Get(0).(*service.Basis), args.Error(1)
}

// Helper to create git.Basis with specific values for testing
func createTestGitBasis(owner, repo, branch string) *git.Basis {
	// We'll create a minimal git.Basis - this might need adjustment based on the actual struct
	return &git.Basis{
		GitOwner:  owner,
		GitRepo:   repo, 
		GitBranch: branch,
		GitSha:    "abc123",
	}
}

type MockServiceBasis struct {
	name string
}

func (m *MockServiceBasis) Name() string { return m.name }

func TestMatchesFilter_DefaultFiltering(t *testing.T) {
	// Test that default filtering works by current git context
	gitBasis := createTestGitBasis("testowner", "testrepo", "testbranch")
	
	serviceBasis := &MockServiceBasis{name: "testservice"}
	mockBasis := &MockBasis{}
	mockBasis.On("Service").Return(serviceBasis, nil)
	
	state := &State{
		basis: mockBasis,
		git:   gitBasis,
	}
	
	// Should match when all fields match
	matchingMetadata := &StateMetadata{
		Service: "anyservice", // Service should be ignored
		Owner:   "testowner",
		Repo:    "testrepo",
		Branch:  "testbranch",
		Sha:     "def456",
	}
	
	assert.True(t, state.matchesFilter(matchingMetadata))
	
	// Should not match when owner differs
	nonMatchingOwner := &StateMetadata{
		Service: "anyservice",
		Owner:   "differentowner",
		Repo:    "testrepo", 
		Branch:  "testbranch",
		Sha:     "def456",
	}
	
	assert.False(t, state.matchesFilter(nonMatchingOwner))
	
	// Should not match when repo differs
	nonMatchingRepo := &StateMetadata{
		Service: "anyservice",
		Owner:   "testowner",
		Repo:    "differentrepo",
		Branch:  "testbranch",
		Sha:     "def456",
	}
	
	assert.False(t, state.matchesFilter(nonMatchingRepo))
	
	// Should not match when branch differs
	nonMatchingBranch := &StateMetadata{
		Service: "anyservice",
		Owner:   "testowner",
		Repo:    "testrepo",
		Branch:  "differentbranch",
		Sha:     "def456",
	}
	
	assert.False(t, state.matchesFilter(nonMatchingBranch))
}

func TestMatchesFilter_WildcardFiltering(t *testing.T) {
	// Test that wildcard (*) filtering works
	gitBasis := createTestGitBasis("*", "testrepo", "*")
	
	serviceBasis := &MockServiceBasis{name: "testservice"}
	mockBasis := &MockBasis{}
	mockBasis.On("Service").Return(serviceBasis, nil)
	
	state := &State{
		basis: mockBasis,
		git:   gitBasis,
	}
	
	// Should match any owner/branch when wildcarded, but still filter by repo
	metadata := &StateMetadata{
		Service: "anyservice",
		Owner:   "anyowner",    // Should match due to wildcard
		Repo:    "testrepo",    // Must match
		Branch:  "anybranch",   // Should match due to wildcard
		Sha:     "def456",
	}
	
	assert.True(t, state.matchesFilter(metadata))
	
	// Should not match when non-wildcarded field (repo) differs
	nonMatchingRepo := &StateMetadata{
		Service: "anyservice",
		Owner:   "anyowner",
		Repo:    "differentrepo", // Should not match
		Branch:  "anybranch",
		Sha:     "def456",
	}
	
	assert.False(t, state.matchesFilter(nonMatchingRepo))
}

func TestMatchesFilter_AllWildcards(t *testing.T) {
	// Test that all wildcards shows everything
	gitBasis := createTestGitBasis("*", "*", "*")
	
	serviceBasis := &MockServiceBasis{name: "*"}
	mockBasis := &MockBasis{}
	mockBasis.On("Service").Return(serviceBasis, nil)
	
	state := &State{
		basis: mockBasis,
		git:   gitBasis,
	}
	
	// Should match anything
	metadata := &StateMetadata{
		Service: "anyservice",
		Owner:   "anyowner",
		Repo:    "anyrepo",
		Branch:  "anybranch",
		Sha:     "def456",
	}
	
	assert.True(t, state.matchesFilter(metadata))
}

func TestMatchesFilter_ServiceFilteringDisabled(t *testing.T) {
	// Test that service filtering is effectively disabled
	gitBasis := createTestGitBasis("testowner", "testrepo", "testbranch")
	
	// Service basis returns a different name than metadata service
	serviceBasis := &MockServiceBasis{name: "differentservice"}
	mockBasis := &MockBasis{}
	mockBasis.On("Service").Return(serviceBasis, nil)
	
	state := &State{
		basis: mockBasis,
		git:   gitBasis,
	}
	
	// Should still match even though service names differ
	// This is the key test - service filtering should be ignored
	metadata := &StateMetadata{
		Service: "actualservice", // Different from serviceBasis.name
		Owner:   "testowner",
		Repo:    "testrepo", 
		Branch:  "testbranch",
		Sha:     "def456",
	}
	
	assert.True(t, state.matchesFilter(metadata), 
		"Service filtering should be disabled - should match regardless of service name")
}

func TestMatchesFilter_ServiceBasisError(t *testing.T) {
	// Test that errors from Service() don't break filtering
	gitBasis := createTestGitBasis("testowner", "testrepo", "testbranch")
	
	mockBasis := &MockBasis{}
	mockBasis.On("Service").Return((*service.Basis)(nil), assert.AnError)
	
	state := &State{
		basis: mockBasis,
		git:   gitBasis,
	}
	
	metadata := &StateMetadata{
		Service: "anyservice",
		Owner:   "testowner",
		Repo:    "testrepo",
		Branch:  "testbranch", 
		Sha:     "def456",
	}
	
	// Should still work even if Service() returns an error
	assert.True(t, state.matchesFilter(metadata))
}