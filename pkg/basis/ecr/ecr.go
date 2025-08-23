package ecr

import (
	"fmt"
	"strings"

	"github.com/bkeane/monad/pkg/basis/caller"
	"github.com/bkeane/monad/pkg/basis/git"
	"github.com/bkeane/monad/pkg/basis/service"

	v "github.com/go-ozzo/ozzo-validation/v4"
)

// Data

type Basis struct {
	Image          string `env:"MONAD_IMAGE" flag:"--image" usage:"ECR image path:tag"`
	RegistryId     string `env:"MONAD_REGISTRY_ID" flag:"--ecr-id" usage:"ECR registry ID"`
	RegistryRegion string `env:"MONAD_REGISTRY_REGION" flag:"--ecr-region" usage:"ECR registry region"`
}

//
// Derive
//

func Derive(caller *caller.Basis, git *git.Basis, service *service.Basis) (*Basis, error) {
	var basis Basis

	if caller == nil {
		return nil, fmt.Errorf("caller basis must not be nil when deriving ecr basis")
	}

	if git == nil {
		return nil, fmt.Errorf("git basis must not be nil when deriving ecr basis")
	}

	if service == nil {
		return nil, fmt.Errorf("service basis must not be nil when deriving ecr basis")
	}

	if basis.RegistryId == "" {
		basis.RegistryId = caller.AccountId
	}

	if basis.RegistryRegion == "" {
		basis.RegistryRegion = caller.AwsConfig.Region
	}

	if basis.Image == "" {
		basis.Image = fmt.Sprintf("%s/%s/%s:%s", git.Owner, git.Repository, service.Name, git.Branch)
	}

	if !strings.Contains(basis.Image, ":") {
		basis.Image = fmt.Sprintf("%s:%s", basis.Image, git.Branch)
	}

	return &basis, nil
}

//
// Validations
//

func (e *Basis) Validate() error {
	return v.ValidateStruct(e,
		v.Field(&e.Image, v.Required),
		v.Field(&e.RegistryId, v.Required),
		v.Field(&e.RegistryRegion, v.Required),
	)
}
