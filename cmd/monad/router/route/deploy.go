package route

import (
	"context"
	"fmt"

	"github.com/bkeane/monad/pkg/param"
	"github.com/bkeane/monad/pkg/saga"
)

type Deploy struct {
	param.Aws
}

func (d *Deploy) Route(ctx context.Context, r Root) error {
	if err := d.Aws.Validate(ctx, r.AwsConfig, r.Git); err != nil {
		return err
	}

	path := fmt.Sprintf("%s/%s/%s", r.Git.Owner, r.Git.Repository, r.Git.Service)
	image, err := d.Registry.Client.FromPath(ctx, path, r.Git.Branch)
	if err != nil {
		return err
	}

	if err := saga.Init(ctx, d.Aws).Do(ctx, image); err != nil {
		return err
	}

	return nil
}
