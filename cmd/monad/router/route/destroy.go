package route

import (
	"context"

	"github.com/bkeane/monad/pkg/model"
	"github.com/bkeane/monad/pkg/saga"
)

type Destroy struct {
	model.Model `arg:"-"`
}

func (d *Destroy) Route(ctx context.Context, r Root) error {
	// Process the embedded model with CLI args
	if err := d.Model.Process(ctx, r.AwsConfig); err != nil {
		return err
	}

	saga, err := saga.Init(ctx, r.AwsConfig)
	if err != nil {
		return err
	}

	return saga.Undo(ctx)
}
