package route

import (
	"context"
	"fmt"

	"github.com/bkeane/monad/internal/uriopt"
	"github.com/bkeane/monad/pkg/model"
	"github.com/bkeane/monad/pkg/template"

	"github.com/charmbracelet/lipgloss/table"
)

type Render struct {
	File string `arg:"positional" placeholder:"template" help:"string | file://template.tmpl"`
}

type Table struct{}

type Data struct {
	Render *Render `arg:"subcommand:render" help:"render template with data"`
	Table  *Table  `arg:"subcommand:table" help:"display template data as table"`
}

func (d *Data) Route(ctx context.Context, r Root) error {
	// Initialize unified model
	model := &model.Model{}
	if err := model.Process(ctx, r.AwsConfig); err != nil {
		return err
	}

	// Create template engine
	templateData := template.TemplateData{
		Git:         model.Axiom().Git(),
		Service:     model.Axiom().Service(),
		Caller:      model.Axiom().Caller(),
		Resource:    model.Axiom().Resource(),
		CloudWatch:  model.CloudWatch(),
		Lambda:      model.Lambda(),
		IAM:         model.IAM(),
		ECR:         model.ECR(),
		ApiGateway:  model.ApiGateway(),
		EventBridge: model.EventBridge(),
	}
	tmpl := template.Init(templateData)

	switch {
	case d.Render != nil:
		content, err := uriopt.String(d.Render.File)
		if err != nil {
			return fmt.Errorf("failed to read template file: %w", err)
		}

		rendered, err := tmpl.Process(content)
		if err != nil {
			return fmt.Errorf("failed to render template: %w", err)
		}

		fmt.Print(rendered)

	case d.Table != nil:
		return d.draw(model)

	}

	return nil
}

func (d *Data) draw(model *model.Model) error {
	tbl := table.New()
	tbl.Headers("Interpolation", "Result")
	tbl.Row("{{.Git.Branch}}", model.Axiom().Git().Branch())
	tbl.Row("{{.Git.Sha}}", model.Axiom().Git().Sha())
	tbl.Row("{{.Git.Owner}}", model.Axiom().Git().Owner())
	tbl.Row("{{.Git.Repository}}", model.Axiom().Git().Repository())
	tbl.Row("{{.Caller.AccountId}}", model.Axiom().Caller().AccountId())
	tbl.Row("{{.Caller.Region}}", model.Axiom().Caller().Region())
	tbl.Row("{{.ECR.Id}}", model.ECR().Id())
	tbl.Row("{{.ECR.Region}}", model.ECR().Region())
	tbl.Row("{{.Resource.NamePrefix}}", model.Axiom().Resource().NamePrefix())
	tbl.Row("{{.Resource.Name}}", model.Axiom().Resource().Name())
	tbl.Row("{{.Resource.PathPrefix}}", model.Axiom().Resource().PathPrefix())
	tbl.Row("{{.Resource.Path}}", model.Axiom().Resource().Path())
	tbl.Row("{{.Lambda.Region}}", model.Lambda().Region())
	tbl.Row("{{.Lambda.FunctionArn}}", model.Lambda().FunctionArn())
	tbl.Row("{{.IAM.PolicyArn}}", model.IAM().PolicyArn())
	tbl.Row("{{.IAM.RoleArn}}", model.IAM().RoleArn())
	tbl.Row("{{.CloudWatch.LogGroupArn}}", model.CloudWatch().LogGroupArn())
	tbl.Row("{{.ApiGateway.Region}}", model.ApiGateway().Region())
	tbl.Row("{{.ApiGateway.ID}}", model.ApiGateway().ID())
	tbl.Row("{{.EventBridge.Region}}", model.EventBridge().Region())
	tbl.Row("{{.EventBridge.RuleName}}", model.EventBridge().RuleName())
	tbl.Row("{{.EventBridge.BusName}}", model.EventBridge().BusName())
	fmt.Println(tbl.Render())
	return nil
}
