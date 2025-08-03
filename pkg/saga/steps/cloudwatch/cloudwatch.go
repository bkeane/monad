package cloudwatch

import (
	"context"

	cw "github.com/bkeane/monad/pkg/client/cloudwatch"
)

type Step struct {
	*cw.Client
}

func Init(client *cw.Client) *Step {
	return &Step{
		client,
	}
}

func (s *Step) Do(ctx context.Context) error {
	return s.Mount(ctx)
}

func (s *Step) Undo(ctx context.Context) error {
	return s.Unmount(ctx)
}
