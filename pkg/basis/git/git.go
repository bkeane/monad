package git

import (
	"os"

	"github.com/bkeane/monad/internal/git"

	"github.com/caarlos0/env/v11"
	v "github.com/go-ozzo/ozzo-validation/v4"
)

//
// Basis
//

type Basis struct {
	cwd       string
	GitOwner  string `env:"MONAD_OWNER" flag:"--owner" usage:"Git repository owner" hint:"name"`
	GitRepo   string `env:"MONAD_REPO" flag:"--repo" usage:"Git repository name" hint:"name"`
	GitBranch string `env:"MONAD_BRANCH" flag:"--branch" usage:"Git branch name" hint:"name"`
	GitSha    string `env:"MONAD_SHA" flag:"--sha" usage:"Git commit SHA" hint:"hash"`
}

//
// Derive
//

func Derive() (*Basis, error) {
	var err error
	var basis Basis

	if err = env.Parse(&basis); err != nil {
		return nil, err
	}

	basis.cwd, err = os.Getwd()
	if err != nil {
		return nil, err
	}

	if basis.GitOwner == "" {
		git, err := git.Parse(basis.cwd)
		if err != nil {
			return nil, err
		}
		basis.GitOwner = git.Owner
	}

	if basis.GitRepo == "" {
		git, err := git.Parse(basis.cwd)
		if err != nil {
			return nil, err
		}
		basis.GitRepo = git.Repo
	}

	if basis.GitBranch == "" {
		git, err := git.Parse(basis.cwd)
		if err != nil {
			return nil, err
		}
		basis.GitBranch = git.Branch
	}

	if basis.GitSha == "" {
		git, err := git.Parse(basis.cwd)
		if err != nil {
			return nil, err
		}
		basis.GitSha = git.Sha
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

func (g *Basis) Validate() error {
	return v.ValidateStruct(g,
		v.Field(&g.GitOwner, v.Required),
		v.Field(&g.GitRepo, v.Required),
		v.Field(&g.GitBranch, v.Required),
		v.Field(&g.GitSha, v.Required),
	)
}

//
// Accessors
//

func (g *Basis) Owner() string {
	return g.GitOwner
}

func (g *Basis) Repo() string {
	return g.GitRepo
}

func (g *Basis) Branch() string {
	return g.GitBranch
}

func (g *Basis) Sha() string {
	return g.GitSha
}
