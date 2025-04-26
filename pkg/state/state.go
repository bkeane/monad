package state

import (
	"context"
	"fmt"

	"github.com/bkeane/monad/pkg/param"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/charmbracelet/lipgloss/table"
)

type State struct {
	config param.Aws
}

func Init(ctx context.Context, c param.Aws) *State {
	return &State{
		config: c,
	}
}

func (s *State) List(ctx context.Context) ([]*param.StateMetadata, error) {
	list := &lambda.ListFunctionsInput{}
	functions, err := s.config.Lambda.Client.ListFunctions(ctx, list)
	if err != nil {
		return nil, err
	}

	var services []*param.StateMetadata
	for _, function := range functions.Functions {
		ok, service := match(function)
		if ok {
			services = append(services, service)
		}
	}

	return services, nil
}

func (s *State) DrawTable(ctx context.Context, services []*param.StateMetadata) error {
	tbl := table.New()
	tbl.Headers("Service", "Owner", "Repo", "Branch", "Sha", "Image", "Api", "Bus")
	for _, service := range services {
		tbl.Row(service.Service, service.Owner, service.Repo, service.Branch, service.Sha, service.Image, service.Api, service.Bus)
	}

	fmt.Println(tbl.Render())

	return nil
}

func match(function types.FunctionConfiguration) (bool, *param.StateMetadata) {
	var ok bool
	svc := &param.StateMetadata{}
	env := function.Environment.Variables

	// Required
	if svc.Service, ok = env["MONAD_SERVICE"]; !ok {
		return false, nil
	}

	if svc.Repo, ok = env["MONAD_REPO"]; !ok {
		return false, nil
	}

	if svc.Sha, ok = env["MONAD_SHA"]; !ok {
		return false, nil
	}

	if svc.Branch, ok = env["MONAD_BRANCH"]; !ok {
		return false, nil
	}

	if svc.Owner, ok = env["MONAD_OWNER"]; !ok {
		return false, nil
	}

	if svc.Image, ok = env["MONAD_IMAGE"]; !ok {
		return false, nil
	}

	// Optional
	svc.Api = env["MONAD_API"]
	svc.Bus = env["MONAD_BUS"]

	return true, svc
}
