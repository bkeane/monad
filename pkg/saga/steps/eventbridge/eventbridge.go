package eventbridge

import (
	"context"

	eb "github.com/bkeane/monad/pkg/client/eventbridge"
)

type Step struct {
	*eb.Client
}

func Init(client *eb.Client) *Step {
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
