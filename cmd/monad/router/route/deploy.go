package route

import (
	"context"

	"github.com/bkeane/monad/pkg/param"
	"github.com/bkeane/monad/pkg/saga"
)

type Deploy struct {
	param.Aws
}

func (d *Deploy) Route(ctx context.Context, r Root) error {
	if err := d.Aws.Validate(ctx, r.AwsConfig, r.GitConfig, r.ServiceConfig); err != nil {
		return err
	}

	if err := saga.Init(ctx, &d.Aws).Do(ctx); err != nil {
		return err
	}

	return nil
}
