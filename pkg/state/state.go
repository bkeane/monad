package state

import (
	"context"

	"github.com/bkeane/monad/pkg/basis"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

//
// StateMetadata
//

type StateMetadata struct {
	Service string
	Owner   string
	Repo    string
	Branch  string
	Sha     string
}

//
// State
//

type State struct {
	basis  *basis.Basis
	client *lambda.Client
}

func Init(basis *basis.Basis) *State {
	return &State{
		basis:  basis,
		client: lambda.NewFromConfig(basis.AwsConfig()),
	}
}

func (s *State) List(ctx context.Context) ([]*StateMetadata, error) {
	list := &lambda.ListFunctionsInput{}
	functions, err := s.client.ListFunctions(ctx, list)
	if err != nil {
		return nil, err
	}

	var services []*StateMetadata
	for _, function := range functions.Functions {
		if metadata := s.extractFromTags(ctx, *function.FunctionArn); metadata != nil {
			services = append(services, metadata)
		}
	}

	return services, nil
}

//
// Helpers
//

func (s *State) extractFromTags(ctx context.Context, functionArn string) *StateMetadata {
	listTags := &lambda.ListTagsInput{
		Resource: &functionArn,
	}
	
	tagsOutput, err := s.client.ListTags(ctx, listTags)
	if err != nil {
		return nil
	}

	tags := tagsOutput.Tags
	if tags == nil {
		return nil
	}

	// Check for Monad tag to identify our functions
	if monad, ok := tags["Monad"]; !ok || monad != "true" {
		return nil
	}

	metadata := &StateMetadata{}

	// Extract required tags
	var ok bool
	if metadata.Service, ok = tags["Service"]; !ok {
		return nil
	}
	if metadata.Owner, ok = tags["Owner"]; !ok {
		return nil
	}
	if metadata.Repo, ok = tags["Repo"]; !ok {
		return nil
	}
	if metadata.Branch, ok = tags["Branch"]; !ok {
		return nil
	}
	if metadata.Sha, ok = tags["Sha"]; !ok {
		return nil
	}

	return metadata
}