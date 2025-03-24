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
	TemplateInput string       `arg:"--template" placeholder:"template" help:"string | file://template.tmpl"`
	Data          TemplateData `arg:"-"`
}

type TemplateData struct {
	Git struct {
		Branch string
		Sha    string
		Owner  string
		Repo   string
	}

	Caller struct {
		AccountId string
		Region    string
	}

	Registry struct {
		Id     string
		Region string
	}

	Resource struct {
		Name struct {
			Prefix string
			Full   string
		}
		Path struct {
			Prefix string
			Full   string
		}
	}

	Lambda struct {
		Region      string
		FunctionArn string
		PolicyArn   string
		RoleArn     string
	}

	ApiGateway struct {
		Region string
		Id     string
	}

	Cloudwatch struct {
		Region      string
		LogGroupArn string
	}

	EventBridge struct {
		Region string
		Rule   struct {
			Arn string
		}
		Bus struct {
			Name string
		}
	}
}

func (d *Template) Route(ctx context.Context, r Root) error {
	if err := d.Aws.Validate(ctx, r.AwsConfig, r.Git, d.Target); err != nil {
		return err
	}

	if err := d.Target.Validate(r.Git); err != nil {
		return err
	}

	d.Data.Git.Sha = r.Git.Sha
	d.Data.Git.Branch = r.Git.Branch
	d.Data.Git.Owner = r.Git.Owner
	d.Data.Git.Repo = r.Git.Repository
	d.Data.Caller.AccountId = d.Caller.AccountId
	d.Data.Caller.Region = d.Caller.Region
	d.Data.Registry.Id = d.Registry.Id
	d.Data.Registry.Region = d.Registry.Region
	d.Data.Resource.Name.Prefix = d.ResourceNamePrefix()
	d.Data.Resource.Name.Full = d.ResourceName()
	d.Data.Resource.Path.Prefix = d.ResourcePathPrefix()
	d.Data.Resource.Path.Full = d.ResourcePath()
	d.Data.Lambda.Region = d.Aws.Lambda.Region
	d.Data.Lambda.FunctionArn = d.FunctionArn()
	d.Data.Lambda.PolicyArn = d.PolicyArn()
	d.Data.Lambda.RoleArn = d.RoleArn()
	d.Data.Cloudwatch.Region = d.CloudWatch.Region
	d.Data.Cloudwatch.LogGroupArn = d.CloudwatchLogGroupArn()
	d.Data.ApiGateway.Region = d.ApiGateway.Region
	d.Data.ApiGateway.Id = d.ApiGateway.Id
	d.Data.EventBridge.Region = d.EventBridge.Region
	d.Data.EventBridge.Rule.Arn = d.ResourcePath()
	d.Data.EventBridge.Bus.Name = d.EventBridge.BusName

	tbl, err := d.Table()
	if err != nil {
		return err
	}

	if d.TemplateInput != "" {
		template, err := uriopt.Json(d.TemplateInput)
		if err != nil {
			return fmt.Errorf("failed to read provided template input: %w", err)
		}

		rendered, err := tmpl.Template("template", template, d.Data)
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
	tbl.Row("{{.Git.Branch}}", s(d.Data.Git.Branch))
	tbl.Row("{{.Git.Sha}}", s(d.Data.Git.Sha))
	tbl.Row("{{.Git.Owner}}", s(d.Data.Git.Owner))
	tbl.Row("{{.Git.Repo}}", s(d.Data.Git.Repo))
	tbl.Row("{{.Caller.AccountId}}", s(d.Data.Caller.AccountId))
	tbl.Row("{{.Caller.Region}}", s(d.Data.Caller.Region))
	tbl.Row("{{.Registry.Id}}", s(d.Data.Registry.Id))
	tbl.Row("{{.Registry.Region}}", s(d.Data.Registry.Region))
	tbl.Row("{{.Resource.Name.Prefix}}", s(d.Data.Resource.Name.Prefix))
	tbl.Row("{{.Resource.Name.Full}}", s(d.Data.Resource.Name.Full))
	tbl.Row("{{.Resource.Path.Prefix}}", s(d.Data.Resource.Path.Prefix))
	tbl.Row("{{.Resource.Path.Full}}", s(d.Data.Resource.Path.Full))
	tbl.Row("{{.Lambda.Region}}", s(d.Data.Lambda.Region))
	tbl.Row("{{.Lambda.FunctionArn}}", s(d.Data.Lambda.FunctionArn))
	tbl.Row("{{.Lambda.PolicyArn}}", s(d.Data.Lambda.PolicyArn))
	tbl.Row("{{.Lambda.RoleArn}}", s(d.Data.Lambda.RoleArn))
	tbl.Row("{{.Cloudwatch.Region}}", s(d.Data.Cloudwatch.Region))
	tbl.Row("{{.Cloudwatch.LogGroupArn}}", s(d.Data.Cloudwatch.LogGroupArn))
	tbl.Row("{{.ApiGateway.Region}}", s(d.Data.ApiGateway.Region))
	tbl.Row("{{.ApiGateway.Id}}", s(d.Data.ApiGateway.Id))
	tbl.Row("{{.EventBridge.Region}}", s(d.Data.EventBridge.Region))
	tbl.Row("{{.EventBridge.Rule.Arn}}", s(d.Data.EventBridge.Rule.Arn))
	tbl.Row("{{.EventBridge.Bus.Name}}", s(d.Data.EventBridge.Bus.Name))
	result := tbl.Render()
	return &result, nil
}
