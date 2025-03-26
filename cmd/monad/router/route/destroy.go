package route

import (
	"context"

	"github.com/bkeane/monad/pkg/param"
	"github.com/bkeane/monad/pkg/saga"
)

type Destroy struct {
	param.Aws
	param.Target
}

func (d *Destroy) Route(ctx context.Context, r Root) error {
	if err := d.Aws.Validate(ctx, r.AwsConfig, r.Git, d.Target); err != nil {
		return err
	}

	if err := saga.Init(ctx, d.Aws).Undo(ctx); err != nil {
		return err
	}

	return nil
}
