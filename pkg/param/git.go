package param

import (
	"os"

	"github.com/bkeane/monad/internal/git"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/rs/zerolog/log"
)

type Git struct {
	cwd        string `arg:"-" json:"-"`
	Chdir      string `arg:"--chdir" placeholder:"path" default:"cwd"`
	Owner      string `arg:"--owner" placeholder:"name" default:"github.com/<owner>/repo.git"`
	Repository string `arg:"--repo" placeholder:"name" default:"github.com/owner/<repo>.git"`
	Service    string `arg:"--service" placeholder:"name" default:"current directory name"`
	Branch     string `arg:"--branch,env:MONAD_BRANCH" placeholder:"name" default:"current git branch"`
	Sha        string `arg:"--sha,env:MONAD_SHA" placeholder:"sha" default:"current git sha"`
}

func (g *Git) Validate() error {
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

	if g.Service == "" {
		git, err := git.Parse(g.cwd)
		if err != nil {
			return err
		}
		g.Service = git.BasePath
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
		Str("cwd", g.cwd).
		Str("owner", g.Owner).
		Str("repo", g.Repository).
		Str("service", g.Service).
		Str("branch", g.Branch).
		Str("sha", g.Sha).
		Msg("parsed .git")

	return v.ValidateStruct(g,
		v.Field(&g.Chdir, v.Required),
		v.Field(&g.Owner, v.Required),
		v.Field(&g.Repository, v.Required),
		v.Field(&g.Service, v.Required),
		v.Field(&g.Branch, v.Required),
		v.Field(&g.Sha, v.Required),
	)
}
