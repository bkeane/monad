package registry

import (
	"fmt"
	"strings"

	"github.com/bkeane/monad/pkg/basis/caller"
	"github.com/bkeane/monad/pkg/basis/git"
	"github.com/bkeane/monad/pkg/basis/service"

	"github.com/caarlos0/env/v11"
	v "github.com/go-ozzo/ozzo-validation/v4"
)

// Data

type Basis struct {
	EcrImage  string `env:"MONAD_IMAGE" flag:"--image" usage:"ECR image path:tag" hint:"path:tag"`
	EcrId     string `env:"MONAD_REGISTRY_ID" flag:"--ecr-id" usage:"ECR registry ID" hint:"id"`
	EcrRegion string `env:"MONAD_REGISTRY_REGION" flag:"--ecr-region" usage:"ECR registry region" hint:"name"`
}

//
// Derive
//

func Derive(caller *caller.Basis, git *git.Basis, service *service.Basis) (*Basis, error) {
	var err error
	var basis Basis

	if err = env.Parse(&basis); err != nil {
		return nil, err
	}

	if caller == nil {
		return nil, fmt.Errorf("caller basis must not be nil when deriving ecr basis")
	}

	if git == nil {
		return nil, fmt.Errorf("git basis must not be nil when deriving ecr basis")
	}

	if service == nil {
		return nil, fmt.Errorf("service basis must not be nil when deriving ecr basis")
	}

	if basis.EcrId == "" {
		basis.EcrId = caller.AccountId()
	}

	if basis.EcrRegion == "" {
		basis.EcrRegion = caller.AwsConfig().Region
	}

	if basis.EcrImage == "" {
		basis.EcrImage = fmt.Sprintf("%s/%s/%s:%s", git.Owner(), git.Repo(), service.Name(), git.Branch())
	}

	if !strings.Contains(basis.EcrImage, ":") {
		basis.EcrImage = fmt.Sprintf("%s:%s", basis.Image(), git.Branch())
	}

	err = basis.Validate()
	if err != nil {
		return nil, err
	}

	return &basis, nil
}

//
// Validations
//

func (e *Basis) Validate() error {
	return v.ValidateStruct(e,
		v.Field(&e.EcrImage, v.Required),
		v.Field(&e.EcrId, v.Required),
		v.Field(&e.EcrRegion, v.Required),
	)
}

// Accessors

func (e *Basis) Image() string {
	return e.EcrImage
}

func (e *Basis) Id() string {
	return e.EcrId
}

func (e *Basis) Region() string {
	return e.EcrRegion
}
