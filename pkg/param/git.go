package param

import (
	"os"

	"github.com/bkeane/monad/internal/git"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/rs/zerolog/log"
)

type GitConfig struct {
	cwd        string `arg:"-" json:"-"`
	Chdir      string `arg:"--chdir,env:MONAD_CHDIR" placeholder:"path" help:"change working directory" default:"."`
	Owner      string `arg:"--owner,env:MONAD_OWNER" placeholder:"name" help:"git repository owner" default:"github.com/<owner>/repo.git"`
	Repository string `arg:"--repo,env:MONAD_REPO" placeholder:"name" help:"git repository name" default:"github.com/owner/<repo>.git"`
	Branch     string `arg:"--branch,env:MONAD_BRANCH" placeholder:"name" help:"git repository branch" default:"current git branch"`
	Sha        string `arg:"--sha,env:MONAD_SHA" placeholder:"sha" help:"git repository sha" default:"current git sha"`
}

func (g *GitConfig) Process() error {
	var err error

	if g.Chdir == "" {
		g.Chdir = "."
	}

	if g.Chdir != "." {
		log.Info().Msgf("chdir to %s", g.Chdir)
		if err := os.Chdir(g.Chdir); err != nil {
			return err
		}
	}

	g.cwd, err = os.Getwd()
	if err != nil {
		return err
	}

	if g.Owner == "" {
		git, err := git.Parse(g.cwd)
		if err != nil {
			return err
		}
		g.Owner = git.Owner
	}

	if g.Repository == "" {
		git, err := git.Parse(g.cwd)
		if err != nil {
			return err
		}
		g.Repository = git.Repo
	}

	if g.Branch == "" {
		git, err := git.Parse(g.cwd)
		if err != nil {
			return err
		}
		g.Branch = git.Branch
	}

	if g.Sha == "" {
		git, err := git.Parse(g.cwd)
		if err != nil {
			return err
		}
		g.Sha = git.Sha
	}

	log.Info().
		Str("owner", g.Owner).
		Str("repo", g.Repository).
		Str("branch", g.Branch).
		Str("sha", truncate(g.Sha)).
		Msg("git")

	return g.Validate()
}

func (g *GitConfig) Validate() error {
	return v.ValidateStruct(g,
		v.Field(&g.Chdir, v.Required),
		v.Field(&g.Owner, v.Required),
		v.Field(&g.Repository, v.Required),
		v.Field(&g.Branch, v.Required),
		v.Field(&g.Sha, v.Required),
	)
}

func truncate(s string) string {
	if len(s) <= 7 {
		return s
	}
	return s[:7]
}
