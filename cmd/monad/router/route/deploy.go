package route

import (
	"context"

	"github.com/bkeane/monad/pkg/param"
	"github.com/bkeane/monad/pkg/saga"
)

type Deploy struct {
	param.Aws
	param.Target
}

func (d *Deploy) Route(ctx context.Context, r Root) error {
	if err := d.Aws.Validate(ctx, r.AwsConfig, r.Git, d.Target); err != nil {
		return err
	}

	if err := d.Target.Validate(r.Git); err != nil {
		return err
	}

	image, err := d.Registry.GetImage(ctx, d.Target.Image, r.Git.Branch)
	if err != nil {
		return err
	}

	if err := saga.Init(ctx, d.Aws).Do(ctx, image); err != nil {
		return err
	}

	return nil
}
