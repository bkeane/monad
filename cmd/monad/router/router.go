package router

import (
	"context"

	"github.com/bkeane/monad/cmd/monad/router/route"
)

func Route(ctx context.Context, r route.Root) error {
	if err := r.Validate(ctx); err != nil {
		return err
	}

	switch {
	case r.Deploy != nil:
		if err := r.Deploy.Route(ctx, r); err != nil {
			return err
		}

	case r.Destroy != nil:
		if err := r.Destroy.Route(ctx, r); err != nil {
			return err
		}

	case r.Compose != nil:
		if err := r.Compose.Route(ctx, r); err != nil {
			return err
		}

	case r.Ecr != nil:
		if err := r.Ecr.Route(ctx, r); err != nil {
			return err
		}

	case r.Init != nil:
		if err := r.Init.Route(ctx, r); err != nil {
			return err
		}
	}

	return nil
}
