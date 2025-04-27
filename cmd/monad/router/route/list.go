package route

import (
	"context"

	"github.com/bkeane/monad/pkg/param"
	"github.com/bkeane/monad/pkg/state"
	"github.com/charmbracelet/lipgloss"
)

type List struct {
	param.Aws   `arg:"-"`
	AllMonad    bool `arg:"--all-monads"`
	AllBranches bool `arg:"--all-branches"`
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
	if l.AllMonad {
		filtered = services
	} else {
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
	}

	// mutate
	for _, service := range filtered {
		service.Sha = shorten(service.Sha)
	}

	// highlight contextual matches
	ink := lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render
	for _, service := range filtered {
		if service.Service == l.Service.Name {
			service.Service = ink(service.Service)
		}

		if service.Owner == l.Git.Owner {
			service.Owner = ink(service.Owner)
		}

		if service.Repo == l.Git.Repository {
			service.Repo = ink(service.Repo)
		}

		if shorten(service.Sha) == shorten(l.Git.Sha) {
			service.Sha = ink(service.Sha)
		}

		if service.Branch == l.Git.Branch {
			service.Branch = ink(service.Branch)
		}

		if service.Image == l.Service.Image {
			service.Image = ink(service.Image)
		}
	}

	if err := state.DrawTable(ctx, filtered); err != nil {
		return err
	}

	return nil
}

func shorten(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}
