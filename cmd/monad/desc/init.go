package desc

import (
	"context"
	"fmt"
	"strings"

	"github.com/bkeane/monad/pkg/basis"
	"github.com/bkeane/monad/pkg/scaffold"
)

// Init returns a description for the init command with available languages
func Init() string {
	// Create a basic scaffold instance to get available languages
	basis, err := basis.Derive(context.Background())
	if err != nil {
		return "Create a new monad project from a template."
	}

	scaffold, err := scaffold.Derive(basis)
	if err != nil {
		return "Create a new monad project from a template."
	}

	languages, err := scaffold.List()
	if err != nil {
		return "Create a new monad project from a template."
	}

	if len(languages) == 0 {
		return "Create a new monad project from a template."
	}

	return fmt.Sprintf("Create a new monad project from a template.\n\nAvailable languages: %s",
		strings.Join(languages, ", "))
}
