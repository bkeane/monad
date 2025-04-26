package route

import (
	"context"
	"fmt"

	"github.com/bkeane/monad/internal/tmpl"
	"github.com/bkeane/monad/internal/uriopt"
	"github.com/bkeane/monad/pkg/param"

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
	tbl := table.New()
	tbl.Headers("Interpolation", "Result")
	tbl.Row("{{.Git.Branch}}", d.TemplateData.Git.Branch)
	tbl.Row("{{.Git.Sha}}", d.TemplateData.Git.Sha)
	tbl.Row("{{.Git.Owner}}", d.TemplateData.Git.Owner)
	tbl.Row("{{.Git.Repo}}", d.TemplateData.Git.Repo)
	tbl.Row("{{.Caller.AccountId}}", d.TemplateData.Caller.AccountId)
	tbl.Row("{{.Caller.Region}}", d.TemplateData.Caller.Region)
	tbl.Row("{{.Registry.Id}}", d.TemplateData.Registry.Id)
	tbl.Row("{{.Registry.Region}}", d.TemplateData.Registry.Region)
	tbl.Row("{{.Resource.Name.Prefix}}", d.TemplateData.Resource.Name.Prefix)
	tbl.Row("{{.Resource.Name.Full}}", d.TemplateData.Resource.Name.Full)
	tbl.Row("{{.Resource.Path.Prefix}}", d.TemplateData.Resource.Path.Prefix)
	tbl.Row("{{.Resource.Path.Full}}", d.TemplateData.Resource.Path.Full)
	tbl.Row("{{.Lambda.Region}}", d.TemplateData.Lambda.Region)
	tbl.Row("{{.Lambda.FunctionArn}}", d.TemplateData.Lambda.FunctionArn)
	tbl.Row("{{.Lambda.PolicyArn}}", d.TemplateData.Lambda.PolicyArn)
	tbl.Row("{{.Lambda.RoleArn}}", d.TemplateData.Lambda.RoleArn)
	tbl.Row("{{.Cloudwatch.Region}}", d.TemplateData.Cloudwatch.Region)
	tbl.Row("{{.Cloudwatch.LogGroupArn}}", d.TemplateData.Cloudwatch.LogGroupArn)
	tbl.Row("{{.ApiGateway.Region}}", d.TemplateData.ApiGateway.Region)
	tbl.Row("{{.ApiGateway.Id}}", d.TemplateData.ApiGateway.Id)
	tbl.Row("{{.EventBridge.Region}}", d.TemplateData.EventBridge.Region)
	tbl.Row("{{.EventBridge.RuleArn}}", d.TemplateData.EventBridge.RuleArn)
	tbl.Row("{{.EventBridge.BusName}}", d.TemplateData.EventBridge.BusName)
	result := tbl.Render()
	return &result, nil
}
