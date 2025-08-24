package resource

import (
	"fmt"

	"github.com/bkeane/monad/pkg/basis/git"
	"github.com/bkeane/monad/pkg/basis/service"
	v "github.com/go-ozzo/ozzo-validation/v4"
)

// Basis

type Basis struct {
	NamePrefix string
	Name       string
	PathPrefix string
	Path       string
	Tags       map[string]string
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

	basis.NamePrefix = fmt.Sprintf("%s-%s", git.Repository, git.Branch)
	basis.Name = fmt.Sprintf("%s-%s", basis.NamePrefix, service.Name)
	basis.PathPrefix = fmt.Sprintf("%s/%s", git.Repository, git.Branch)
	basis.Path = fmt.Sprintf("%s/%s", basis.PathPrefix, service.Name)

	basis.Tags = map[string]string{
		"Monad":   "true",
		"Service": service.Name,
		"Owner":   git.Owner,
		"Repo":    git.Repository,
		"Branch":  git.Branch,
		"Sha":     git.Sha,
	}

	return &basis, nil
}

//
// Validations
//

func (r *Basis) Validate() error {
	return v.ValidateStruct(r,
		v.Field(&r.NamePrefix, v.Required),
		v.Field(&r.Name, v.Required),
		v.Field(&r.PathPrefix, v.Required),
		v.Field(&r.Path, v.Required),
		v.Field(&r.Tags, v.Required),
	)
}
