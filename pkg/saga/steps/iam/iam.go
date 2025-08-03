package iam

import (
	"context"

	iam "github.com/bkeane/monad/pkg/client/iam"
)

type Step struct {
	*iam.Client
}

func Init(client *iam.Client) *Step {
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
