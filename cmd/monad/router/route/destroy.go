package route

import (
	"context"
	"fmt"

	"github.com/bkeane/monad/pkg/param"
	"github.com/bkeane/monad/pkg/saga"
	"github.com/rs/zerolog/log"
)

type Destroy struct {
	param.Aws `arg:"-"`
	Untag     bool `arg:"--untag" default:"false"`
}

func (d *Destroy) Route(ctx context.Context, r Root) error {
	if err := d.Aws.Validate(ctx, r.AwsConfig, r.Git); err != nil {
		return err
	}

	if err := saga.Init(ctx, d.Aws).Undo(ctx); err != nil {
		return err
	}

	if d.Untag {
		path := fmt.Sprintf("%s/%s/%s", r.Git.Owner, r.Git.Repository, r.Git.Service)
		log.Info().Str("repository", path).Str("tag", r.Git.Branch).Msg("untagging image")
		if err := d.Registry.Client.Untag(ctx, path, r.Git.Branch); err != nil {
			return err
		}
	}

	return nil
}
