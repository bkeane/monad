package apigateway

import (
	"context"

	gw "github.com/bkeane/monad/pkg/client/apigateway"
)

type Step struct {
	*gw.Client
}

func Init(client *gw.Client) *Step {
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
