package lambda

import (
	"context"

	"github.com/bkeane/monad/internal/registry"
	lmb "github.com/bkeane/monad/pkg/client/lambda"
)

type Step struct {
	*lmb.Client
}

func Init(client *lmb.Client) *Step {
	return &Step{
		client,
	}
}

func (s *Step) Do(ctx context.Context, image registry.ImagePointer) error {
	return s.Mount(ctx, image)
}

func (s *Step) Undo(ctx context.Context) error {
	return s.Unmount(ctx)
}
