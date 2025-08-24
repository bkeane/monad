package state

import (
	"context"
	"sort"

	"github.com/bkeane/monad/pkg/basis"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/charmbracelet/lipgloss/table"
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

func Init(ctx context.Context) (*State, error) {
	basis, err := basis.Derive(ctx)
	if err != nil {
		return nil, err
	}

	return &State{
		basis:  basis,
		client: lambda.NewFromConfig(basis.AwsConfig()),
	}, nil
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

func (s *State) Table(ctx context.Context) (string, error) {
	services, err := s.List(ctx)
	if err != nil {
		return "", err
	}

	// Sort by repo first, then by branch with current branch first within each repo
	currentBranch := s.basis.GitBasis.Branch
	sort.Slice(services, func(i, j int) bool {
		// First group by repo
		if services[i].Repo != services[j].Repo {
			return services[i].Repo < services[j].Repo
		}

		// Within the same repo, current branch comes first
		if services[i].Branch == currentBranch && services[j].Branch != currentBranch {
			return true
		}
		if services[i].Branch != currentBranch && services[j].Branch == currentBranch {
			return false
		}

		// Otherwise sort alphabetically by branch within the same repo
		return services[i].Branch < services[j].Branch
	})

	tbl := table.New()
	tbl.Headers("Service", "Owner", "Repo", "Branch", "Sha")

	for _, service := range services {
		tbl.Row(service.Service, service.Owner, service.Repo, service.Branch, truncate(service.Sha))
	}

	return tbl.Render(), nil
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

// truncate shortens git SHA to 7 characters for display
func truncate(s string) string {
	if len(s) <= 7 {
		return s
	}
	return s[:7]
}
