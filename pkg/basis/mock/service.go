package mock

import "github.com/bkeane/monad/pkg/basis/service"

// NewMockService creates a realistic mock service.Basis for testing
func NewMockService() *service.Basis {
	return &service.Basis{
		ServiceName: "test-service",
	}
}

// NewMockServiceWithName creates a mock service with specific name
func NewMockServiceWithName(name string) *service.Basis {
	return &service.Basis{
		ServiceName: name,
	}
}