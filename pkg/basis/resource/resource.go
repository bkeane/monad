package resource

import (
	"fmt"

	"github.com/bkeane/monad/pkg/basis/git"
	"github.com/bkeane/monad/pkg/basis/service"
	v "github.com/go-ozzo/ozzo-validation/v4"
)

// Basis

type Basis struct {
	ResourceNamePrefix string
	ResourceName       string
	ResourcePathPrefix string
	ResourcePath       string
	ResourceTags       map[string]string
}

//
// Derive
//

func Derive(git *git.Basis, service *service.Basis) (*Basis, error) {
	var basis Basis

	if git == nil {
		return nil, fmt.Errorf("git basis must not be nil when deriving resource basis")
	}

	if service == nil {
		return nil, fmt.Errorf("service basis must not be nil when deriving resource basis")
	}

	basis.ResourceNamePrefix = fmt.Sprintf("%s-%s", git.Repo(), git.Branch())
	basis.ResourceName = fmt.Sprintf("%s-%s", basis.ResourceNamePrefix, service.Name())
	basis.ResourcePathPrefix = fmt.Sprintf("%s/%s", git.Repo(), git.Branch())
	basis.ResourcePath = fmt.Sprintf("%s/%s", basis.ResourcePathPrefix, service.Name())

	basis.ResourceTags = map[string]string{
		"Monad":   "true",
		"Service": service.Name(),
		"Owner":   git.Owner(),
		"Repo":    git.Repo(),
		"Branch":  git.Branch(),
		"Sha":     git.Sha(),
	}

	return &basis, nil
}

//
// Validations
//

func (r *Basis) Validate() error {
	return v.ValidateStruct(r,
		v.Field(&r.ResourceNamePrefix, v.Required),
		v.Field(&r.ResourceName, v.Required),
		v.Field(&r.ResourcePathPrefix, v.Required),
		v.Field(&r.ResourcePath, v.Required),
		v.Field(&r.ResourceTags, v.Required),
	)
}

//
// Accessors
//

func (r *Basis) NamePrefix() string {
	return r.ResourceNamePrefix
}

func (r *Basis) Name() string {
	return r.ResourceName
}

func (r *Basis) PathPrefix() string {
	return r.ResourcePathPrefix
}

func (r *Basis) Path() string {
	return r.ResourcePath
}

func (r *Basis) Tags() map[string]string {
	return r.ResourceTags
}
