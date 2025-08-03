package route

import (
	"context"
	"strings"

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

	img := strings.Split(r.ServiceConfig.Image, ":")

	image, err := d.Registry().Client().GetImage(ctx, img[0], img[1])
	if err != nil {
		return err
	}

	if err := saga.Init(ctx, &d.Aws).Do(ctx, image); err != nil {
		return err
	}

	return nil
}
