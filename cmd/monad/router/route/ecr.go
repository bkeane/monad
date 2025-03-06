package route

import (
	"context"

	"github.com/bkeane/monad/pkg/param"
)

type Untag struct{}
type Login struct{}

type Ecr struct {
	param.Registry
	Untag *Untag `arg:"subcommand:untag"`
	Login *Login `arg:"subcommand:login"`
}

func (e *Ecr) Route(ctx context.Context, r Root) error {
	if err := e.Registry.Validate(ctx, r.AwsConfig); err != nil {
		return err
	}

	switch {
	case r.Ecr.Login != nil:
		if err := e.Registry.Login(ctx); err != nil {
			return err
		}
	case r.Ecr.Untag != nil:
		if err := e.Registry.Untag(ctx, r.Git.Owner, r.Git.Repository, r.Git.Service, r.Git.Branch); err != nil {
			return err
		}
	}

	return nil
}
