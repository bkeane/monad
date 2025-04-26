package route

import (
	"context"

	"github.com/bkeane/monad/pkg/param"
	"github.com/bkeane/monad/pkg/state"
)

type List struct {
	param.Aws   `arg:"-"`
	AllBranches bool `arg:"--all"`
}

func (l *List) Route(ctx context.Context, r Root) error {
	if err := l.Aws.Validate(ctx, r.AwsConfig, r.Git, r.Service); err != nil {
		return err
	}

	state := state.Init(ctx, l.Aws)
	services, err := state.List(ctx)
	if err != nil {
		return err
	}

	var filtered []*param.StateMetadata
	// match := lipgloss.NewStyle().Foreground(lipgloss.Color("150")).Render
	for _, service := range services {
		if service.Owner == l.Git.Owner {
			if service.Repo == l.Git.Repository {
				if service.Branch == l.Git.Branch {
					filtered = append(filtered, service)
				} else if l.AllBranches {
					filtered = append(filtered, service)
				}
			}
		}
	}

	if err := state.DrawTable(ctx, filtered); err != nil {
		return err
	}

	return nil
}
