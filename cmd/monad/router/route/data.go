package route

import (
	"context"
	"fmt"

	"github.com/bkeane/monad/internal/tmpl"
	"github.com/bkeane/monad/internal/uriopt"
	"github.com/bkeane/monad/pkg/param"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type Render struct {
	File string `arg:"positional" placeholder:"template" help:"string | file://template.tmpl"`
}

type Table struct{}

type Data struct {
	param.Aws `arg:"-"`
	Render    *Render `arg:"subcommand:render" help:"render template with data"`
	Table     *Table  `arg:"subcommand:table" help:"display template data as table"`
}

func (d *Data) Route(ctx context.Context, r Root) error {
	if err := d.Aws.Validate(ctx, r.AwsConfig, r.Git, r.Service); err != nil {
		return err
	}

	switch {
	case d.Render != nil:
		content, err := uriopt.String(d.Render.File)
		if err != nil {
			return fmt.Errorf("failed to read template file: %w", err)
		}

		rendered, err := tmpl.Template("template", content, d.TemplateData)
		if err != nil {
			return fmt.Errorf("failed to render template: %w", err)
		}

		fmt.Print(rendered)

	case d.Table != nil:
		tbl, err := d.DrawTable()
		if err != nil {
			return err
		}
		fmt.Print(*tbl)

	}

	return nil
}

func (d *Data) DrawTable() (*string, error) {
	s := lipgloss.NewStyle().Foreground(lipgloss.Color("150")).Render
	tbl := table.New()
	tbl.Headers("Interpolation", "Result")
	tbl.Row("{{.Git.Branch}}", s(d.TemplateData.Git.Branch))
	tbl.Row("{{.Git.Sha}}", s(d.TemplateData.Git.Sha))
	tbl.Row("{{.Git.Owner}}", s(d.TemplateData.Git.Owner))
	tbl.Row("{{.Git.Repo}}", s(d.TemplateData.Git.Repo))
	tbl.Row("{{.Caller.AccountId}}", s(d.TemplateData.Caller.AccountId))
	tbl.Row("{{.Caller.Region}}", s(d.TemplateData.Caller.Region))
	tbl.Row("{{.Registry.Id}}", s(d.TemplateData.Registry.Id))
	tbl.Row("{{.Registry.Region}}", s(d.TemplateData.Registry.Region))
	tbl.Row("{{.Resource.Name.Prefix}}", s(d.TemplateData.Resource.Name.Prefix))
	tbl.Row("{{.Resource.Name.Full}}", s(d.TemplateData.Resource.Name.Full))
	tbl.Row("{{.Resource.Path.Prefix}}", s(d.TemplateData.Resource.Path.Prefix))
	tbl.Row("{{.Resource.Path.Full}}", s(d.TemplateData.Resource.Path.Full))
	tbl.Row("{{.Lambda.Region}}", s(d.TemplateData.Lambda.Region))
	tbl.Row("{{.Lambda.FunctionArn}}", s(d.TemplateData.Lambda.FunctionArn))
	tbl.Row("{{.Lambda.PolicyArn}}", s(d.TemplateData.Lambda.PolicyArn))
	tbl.Row("{{.Lambda.RoleArn}}", s(d.TemplateData.Lambda.RoleArn))
	tbl.Row("{{.Cloudwatch.Region}}", s(d.TemplateData.Cloudwatch.Region))
	tbl.Row("{{.Cloudwatch.LogGroupArn}}", s(d.TemplateData.Cloudwatch.LogGroupArn))
	tbl.Row("{{.ApiGateway.Region}}", s(d.TemplateData.ApiGateway.Region))
	tbl.Row("{{.ApiGateway.Id}}", s(d.TemplateData.ApiGateway.Id))
	tbl.Row("{{.EventBridge.Region}}", s(d.TemplateData.EventBridge.Region))
	tbl.Row("{{.EventBridge.Rule.Arn}}", s(d.TemplateData.EventBridge.Rule.Arn))
	tbl.Row("{{.EventBridge.Bus.Name}}", s(d.TemplateData.EventBridge.Bus.Name))
	result := tbl.Render()
	return &result, nil
}
