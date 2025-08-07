package route

import (
	"context"
	"fmt"

	statepkg "github.com/bkeane/monad/pkg/state"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type List struct {
	All         bool `arg:"--all" help:"show services of all branches of all repos"`
	AllBranches bool `arg:"--all-branches" help:"show all services of all branches of current repo"`
}

func (l *List) Route(ctx context.Context, r Root) error {
	state, err := statepkg.Init(ctx, r.AwsConfig)
	if err != nil {
		return err
	}
	
	services, err := state.List(ctx)
	if err != nil {
		return err
	}

	var filtered []*statepkg.StateMetadata
	if l.All {
		filtered = services
	} else {
		for _, service := range services {
			if service.Owner == r.Git().Owner() {
				if service.Repo == r.Git().Repository() {
					if service.Branch == r.Git().Branch() {
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

		if service.Api == "" {
			service.Api = "none"
		}

		if service.Bus == "" {
			service.Bus = "none"
		}
	}

	// highlight contextual matches
	ink := lipgloss.NewStyle().Foreground(lipgloss.Color("69")).Render
	for _, service := range filtered {
		if service.Service == r.Service().Name() {
			service.Service = ink(service.Service)
		}

		if service.Owner == r.Git().Owner() {
			service.Owner = ink(service.Owner)
		}

		if service.Repo == r.Git().Repository() {
			service.Repo = ink(service.Repo)
		}

		if shorten(service.Sha) == shorten(r.Git().Sha()) {
			service.Sha = ink(service.Sha)
		}

		if service.Branch == r.Git().Branch() {
			service.Branch = ink(service.Branch)
		}

		if service.Image == r.Service().Image() {
			service.Image = ink(service.Image)
		}
	}

	return draw(filtered)
}

func shorten(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}

func draw(services []*statepkg.StateMetadata) error {
	tbl := table.New()
	tbl.Headers("Service", "Owner", "Repo", "Branch", "Sha", "Image", "Api", "Bus")
	for _, service := range services {
		tbl.Row(service.Service, service.Owner, service.Repo, service.Branch, service.Sha, service.Image, service.Api, service.Bus)
	}

	fmt.Println(tbl.Render())

	return nil
}
