package mock

import (
	"bytes"
	"errors"
	"text/template"

	"github.com/bkeane/monad/pkg/basis"
	"github.com/bkeane/monad/pkg/basis/caller"
	"github.com/bkeane/monad/pkg/basis/defaults"
	"github.com/bkeane/monad/pkg/basis/git"
	"github.com/bkeane/monad/pkg/basis/registry"
	"github.com/bkeane/monad/pkg/basis/resource"
	"github.com/bkeane/monad/pkg/basis/service"
)

// Mock error constants for testing
var (
	ErrGitNotFound      = errors.New("mock: git not found")
	ErrCallerNotFound   = errors.New("mock: caller not found")
	ErrServiceNotFound  = errors.New("mock: service not found")
	ErrResourceNotFound = errors.New("mock: resource not found")
	ErrRegistryNotFound = errors.New("mock: registry not found")
	ErrDefaultsNotFound = errors.New("mock: defaults not found")
	ErrRenderFailed     = errors.New("mock: render failed")
)

// MockBasis provides a complete mock implementation of the basis interfaces
type MockBasis struct {
	caller   *caller.Basis
	git      *git.Basis
	service  *service.Basis
	resource *resource.Basis
	registry *registry.Basis
	defaults *defaults.Basis

	// Track method calls for testing
	CallCounts map[string]int
}

// BasisOptions allows customization of mock basis components
type BasisOptions struct {
	Owner     string
	Repo      string
	Branch    string
	Service   string
	AccountId string
	Region    string
}

// DefaultBasisOptions returns sensible defaults for testing
func DefaultBasisOptions() BasisOptions {
	return BasisOptions{
		Owner:     "test-owner",
		Repo:      "test-repo", 
		Branch:    "test-branch",
		Service:   "test-service",
		AccountId: "123456789012",
		Region:    "us-east-1",
	}
}

// NewMockBasis creates a complete mock basis with realistic components
func NewMockBasis() *MockBasis {
	return NewMockBasisWithOptions(DefaultBasisOptions())
}

// NewMockBasisWithOptions creates a mock basis with customized options
func NewMockBasisWithOptions(opts BasisOptions) *MockBasis {
	git := NewMockGitWithRepo(opts.Owner, opts.Repo, opts.Branch)
	caller := NewMockCallerWithAccount(opts.AccountId)
	caller.CallerConfig.Region = opts.Region
	service := NewMockServiceWithName(opts.Service)
	resource := NewMockResourceWithName(opts.Owner, opts.Repo, opts.Branch, opts.Service)
	registry := NewMockRegistryWithAccount(opts.AccountId)
	defaults := NewMockDefaults()

	return &MockBasis{
		caller:     caller,
		git:        git,
		service:    service,
		resource:   resource,
		registry:   registry,
		defaults:   defaults,
		CallCounts: make(map[string]int),
	}
}

// Interface implementation with call tracking

func (m *MockBasis) Git() (*git.Basis, error) {
	m.CallCounts["Git"]++
	return m.git, nil
}

func (m *MockBasis) Caller() (*caller.Basis, error) {
	m.CallCounts["Caller"]++
	return m.caller, nil
}

func (m *MockBasis) Service() (*service.Basis, error) {
	m.CallCounts["Service"]++
	return m.service, nil
}

func (m *MockBasis) Resource() (*resource.Basis, error) {
	m.CallCounts["Resource"]++
	return m.resource, nil
}

func (m *MockBasis) Registry() (*registry.Basis, error) {
	m.CallCounts["Registry"]++
	return m.registry, nil
}

func (m *MockBasis) Defaults() (*defaults.Basis, error) {
	m.CallCounts["Defaults"]++
	return m.defaults, nil
}

func (m *MockBasis) Render(templateStr string) (string, error) {
	m.CallCounts["Render"]++
	
	// Create template data from our mock components
	data := basis.TemplateData{}
	data.Account.Id = m.caller.AccountId()
	data.Account.Region = m.caller.AwsConfig().Region
	data.Git.Repo = m.git.Repo()
	data.Git.Owner = m.git.Owner() 
	data.Git.Branch = m.git.Branch()
	data.Git.Sha = m.git.Sha()
	data.Service.Name = m.service.Name()
	data.Resource.Name = m.resource.Name()
	data.Resource.Path = m.resource.Path()

	tmpl, err := template.New("template").Parse(templateStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Helper methods for testing

func (m *MockBasis) GetCallCount(method string) int {
	return m.CallCounts[method]
}

func (m *MockBasis) ResetCallCounts() {
	m.CallCounts = make(map[string]int)
}

// Mock basis that returns errors for testing failure scenarios
type MockBasisWithErrors struct {
	CallCounts map[string]int
}

func NewMockBasisWithErrors() *MockBasisWithErrors {
	return &MockBasisWithErrors{
		CallCounts: make(map[string]int),
	}
}

func (m *MockBasisWithErrors) Git() (*git.Basis, error) {
	m.CallCounts["Git"]++
	return nil, ErrGitNotFound
}

func (m *MockBasisWithErrors) Caller() (*caller.Basis, error) {
	m.CallCounts["Caller"]++
	return nil, ErrCallerNotFound
}

func (m *MockBasisWithErrors) Service() (*service.Basis, error) {
	m.CallCounts["Service"]++
	return nil, ErrServiceNotFound
}

func (m *MockBasisWithErrors) Resource() (*resource.Basis, error) {
	m.CallCounts["Resource"]++
	return nil, ErrResourceNotFound
}

func (m *MockBasisWithErrors) Registry() (*registry.Basis, error) {
	m.CallCounts["Registry"]++
	return nil, ErrRegistryNotFound
}

func (m *MockBasisWithErrors) Defaults() (*defaults.Basis, error) {
	m.CallCounts["Defaults"]++
	return nil, ErrDefaultsNotFound
}

func (m *MockBasisWithErrors) Render(template string) (string, error) {
	m.CallCounts["Render"]++
	return "", ErrRenderFailed
}

func (m *MockBasisWithErrors) GetCallCount(method string) int {
	return m.CallCounts[method]
}

func (m *MockBasisWithErrors) ResetCallCounts() {
	m.CallCounts = make(map[string]int)
}