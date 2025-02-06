package tmpl

import (
	"github.com/bkeane/monad/internal/git"
	"github.com/bkeane/substrate/pkg/env"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type Config interface {
	ResourceName() string
	ResourceNamePrefix() string
	ResourcePath() string
	ResourcePathPrefix() string
	TemplateData() (env.Account, env.ECR, env.EventBridge, env.Lambda, env.ApiGateway, git.Git)
}

type Name struct {
	Full   string
	Prefix string
}

type Path struct {
	Full   string
	Prefix string
}

type Resource struct {
	Name Name
	Path Path
}

type TemplateData struct {
	Resource    Resource
	Account     env.Account
	Ecr         env.ECR
	EventBridge env.EventBridge
	Lambda      env.Lambda
	ApiGateway  env.ApiGateway
	Git         git.Git
}

func Init(c Config) *TemplateData {
	caller, ecr, eventBridge, lambda, apiGateway, git := c.TemplateData()

	return &TemplateData{
		Resource: Resource{
			Name: Name{
				Full:   c.ResourceName(),
				Prefix: c.ResourceNamePrefix(),
			},
			Path: Path{
				Full:   c.ResourcePath(),
				Prefix: c.ResourcePathPrefix(),
			},
		},
		Account:     caller,
		Ecr:         ecr,
		EventBridge: eventBridge,
		Lambda:      lambda,
		ApiGateway:  apiGateway,
		Git:         git,
	}
}

func (t TemplateData) Table() (*string, error) {
	s := lipgloss.NewStyle().Foreground(lipgloss.Color("150")).Render
	tbl := table.New()
	tbl.Headers("Interpolation", "Rendered")
	tbl.Row("{{.Resource.Name.Prefix}}", s(t.Resource.Name.Prefix))
	tbl.Row("{{.Resource.Name.Full}}", s(t.Resource.Name.Full))
	tbl.Row("{{.Resource.Path.Prefix}}", s(t.Resource.Path.Prefix))
	tbl.Row("{{.Resource.Path.Full}}", s(t.Resource.Path.Full))
	tbl.Row("{{.Git.BasePath}}", s(t.Git.BasePath))
	tbl.Row("{{.Git.Origin}}", s(t.Git.Origin))
	tbl.Row("{{.Git.Branch}}", s(t.Git.Branch))
	tbl.Row("{{.Git.Sha}}", s(t.Git.Sha))
	tbl.Row("{{.Account.Id}}", s(*t.Account.Id))
	tbl.Row("{{.Account.Name}}", s(t.Account.Name))
	tbl.Row("{{.Account.Region}}", s(*t.Account.Region))
	tbl.Row("{{.Ecr.Id}}", s(*t.Ecr.Id))
	tbl.Row("{{.Ecr.Region}}", s(*t.Ecr.Region))
	tbl.Row("{{.EventBridge.BusName}}", s(t.EventBridge.BusName))
	tbl.Row("{{.EventBridge.Region}}", s(*t.EventBridge.Region))
	tbl.Row("{{.Lambda.Region}}", s(*t.Lambda.Region))
	tbl.Row("{{.ApiGateway.Region}}", s(*t.ApiGateway.Region))
	result := tbl.Render()
	return &result, nil
}
