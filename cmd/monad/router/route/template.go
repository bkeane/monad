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

type Template struct {
	param.Aws `arg:"-"`
	param.Target
	TemplateInput string `arg:"--template" placeholder:"template" help:"string | file://template.tmpl"`
}

func (d *Template) Route(ctx context.Context, r Root) error {
	if err := d.Target.Validate(r.Git); err != nil {
		return err
	}

	if err := d.Aws.Validate(ctx, r.AwsConfig, r.Git, d.Target); err != nil {
		return err
	}

	tbl, err := d.Table()
	if err != nil {
		return err
	}

	if d.TemplateInput != "" {
		template, err := uriopt.Json(d.TemplateInput)
		if err != nil {
			return fmt.Errorf("failed to read provided template input: %w", err)
		}

		rendered, err := tmpl.Template("template", template, d.TemplateData)
		if err != nil {
			return fmt.Errorf("failed to render template: %w", err)
		}

		fmt.Println(rendered)
		return nil
	}

	fmt.Println(*tbl)

	return nil
}

func (d *Template) Table() (*string, error) {
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
