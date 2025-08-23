package git

import (
	"os"

	"github.com/bkeane/monad/internal/git"

	v "github.com/go-ozzo/ozzo-validation/v4"
)

//
// Basis
//

type Basis struct {
	cwd        string
	Owner      string `env:"MONAD_OWNER" flag:"--owner" usage:"Git repository owner"`
	Repository string `env:"MONAD_REPO" flag:"--repo" usage:"Git repository name"`
	Branch     string `env:"MONAD_BRANCH" flag:"--branch" usage:"Git branch name"`
	Sha        string `env:"MONAD_SHA" flag:"--sha" usage:"Git commit SHA"`
}

//
// Derive
//

func Derive() (*Basis, error) {
	var err error
	var basis Basis

	basis.cwd, err = os.Getwd()
	if err != nil {
		return nil, err
	}

	if basis.Owner == "" {
		git, err := git.Parse(basis.cwd)
		if err != nil {
			return nil, err
		}
		basis.Owner = git.Owner
	}

	if basis.Repository == "" {
		git, err := git.Parse(basis.cwd)
		if err != nil {
			return nil, err
		}
		basis.Repository = git.Repo
	}

	if basis.Branch == "" {
		git, err := git.Parse(basis.cwd)
		if err != nil {
			return nil, err
		}
		basis.Branch = git.Branch
	}

	if basis.Sha == "" {
		git, err := git.Parse(basis.cwd)
		if err != nil {
			return nil, err
		}
		basis.Sha = git.Sha
	}

	return &basis, nil
}

//
// Validations
//

func (g *Basis) Validate() error {
	return v.ValidateStruct(g,
		v.Field(&g.Owner, v.Required),
		v.Field(&g.Repository, v.Required),
		v.Field(&g.Branch, v.Required),
		v.Field(&g.Sha, v.Required),
	)
}
