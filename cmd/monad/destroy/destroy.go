package destroy

import (
	"context"
	"path/filepath"

	"github.com/bkeane/monad/internal/git"
	"github.com/bkeane/monad/pkg/config/source"
	"github.com/bkeane/monad/pkg/event"
	"github.com/bkeane/substrate/pkg/registry"
	"github.com/bkeane/substrate/pkg/substrate"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type Dockerfile struct {
	Path string `arg:"positional" help:"path to dockerfile" default:"./Dockerfile"`
}

type Options struct {
	Substrate   string   `arg:"-s,--substrate" help:"substrate name" default:"platform"`
	Destination []string `arg:"-d,--destination" help:"destroy in these accounts"`
	Context     string   `arg:"-c,--context" help:"set path context" default:"."`
}

type Overrides struct {
	Branch *string `arg:"-b,--branch" help:"target alternate branch"`
	Sha    *string `arg:"-s,--sha" help:"target alternate sha"`
}

type Root struct {
	Options
	Overrides
	Dockerfile
	source *source.Config
}

func (r *Root) Route(ctx context.Context, awsconfig aws.Config) (*string, error) {
	var release registry.ImagePointer
	var err error

	path := filepath.Join(r.Options.Context, r.Dockerfile.Path)

	git, err := git.Git{}.Parse(path)
	if err != nil {
		return nil, err
	}

	substrate, err := substrate.Parse(ctx, awsconfig, r.Options.Substrate)
	if err != nil {
		return nil, err
	}

	r.source, err = source.Config{}.Parse(ctx, awsconfig, git, substrate)
	if err != nil {
		return nil, err
	}

	switch {
	case r.Overrides.Branch != nil:
		release, err = substrate.ECR.FetchByName(ctx, r.source.ImagePath(), *r.Overrides.Branch)
		if err != nil {
			return nil, err
		}
	case r.Overrides.Sha != nil:
		release, err = substrate.ECR.FetchByName(ctx, r.source.ImagePath(), *r.Overrides.Sha)
		if err != nil {
			return nil, err
		}
	default:
		release, err = substrate.ECR.FetchByName(ctx, r.source.ImagePath(), git.Branch)
		if err != nil {
			return nil, err
		}
	}

	msg := &event.DestroyRequest{
		ImageUri: release.Uri,
	}

	if r.Options.Destination == nil {
		r.Options.Destination = []string{substrate.Source}
	}

	return substrate.Tx(ctx, msg, r.Options.Destination, []string{})
}
