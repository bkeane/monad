package git

import (
	"os"

	"github.com/bkeane/monad/internal/git"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/rs/zerolog/log"
)

//
// Data
//

type Data struct {
	cwd        string
	owner      string
	repository string
	branch     string
	sha        string
}

//
// Derive
//

func Derive() (*Data, error) {
	var err error
	var data Data

	data.cwd, err = os.Getwd()
	if err != nil {
		return nil, err
	}

	if data.owner == "" {
		git, err := git.Parse(data.cwd)
		if err != nil {
			return nil, err
		}
		data.owner = git.Owner
	}

	if data.repository == "" {
		git, err := git.Parse(data.cwd)
		if err != nil {
			return nil, err
		}
		data.repository = git.Repo
	}

	if data.branch == "" {
		git, err := git.Parse(data.cwd)
		if err != nil {
			return nil, err
		}
		data.branch = git.Branch
	}

	if data.sha == "" {
		git, err := git.Parse(data.cwd)
		if err != nil {
			return nil, err
		}
		data.sha = git.Sha
	}

	log.Info().
		Str("owner", data.owner).
		Str("repo", data.repository).
		Str("branch", data.branch).
		Str("sha", truncate(data.sha)).
		Msg("git")

	if err := data.Validate(); err != nil {
		return nil, err
	}

	return &data, nil
}

//
// Validations
//

func (g *Data) Validate() error {
	return v.ValidateStruct(g,
		v.Field(&g.owner, v.Required),
		v.Field(&g.repository, v.Required),
		v.Field(&g.branch, v.Required),
		v.Field(&g.sha, v.Required),
	)
}

//
// Accessors
//

func (g *Data) Owner() string      { return g.owner }      // Git repository owner
func (g *Data) Repository() string { return g.repository } // Git repository name
func (g *Data) Branch() string     { return g.branch }     // Git branch name
func (g *Data) Sha() string        { return g.sha }        // Git commit SHA

//
// Helpers
//

// truncate shortens git SHA to 7 characters for display
func truncate(s string) string {
	if len(s) <= 7 {
		return s
	}
	return s[:7]
}
