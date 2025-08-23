package basis

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/bkeane/monad/pkg/basis/caller"
	"github.com/bkeane/monad/pkg/basis/defaults"
	"github.com/bkeane/monad/pkg/basis/ecr"
	"github.com/bkeane/monad/pkg/basis/git"
	"github.com/bkeane/monad/pkg/basis/resource"
	"github.com/bkeane/monad/pkg/basis/service"
	"github.com/rs/zerolog/log"

	env "github.com/caarlos0/env/v11"
	v "github.com/go-ozzo/ozzo-validation/v4"
)

//
// Basis
//

type Basis struct {
	Chdir         string `env:"MONAD_CHDIR" flag:"--chdir" usage:"Change working directory"`
	GitBasis      *git.Basis
	CallerBasis   *caller.Basis
	ServiceBasis  *service.Basis
	EcrBasis      *ecr.Basis
	ResourceBasis *resource.Basis
	DefaultBasis  *defaults.Basis
}

type TemplateData struct {
	Account struct {
		Id     string
		Region string
	}

	Git struct {
		Repo   string
		Owner  string
		Branch string
		Sha    string
	}

	Service struct {
		Name string
	}

	Resource struct {
		Name string
		Path string
	}
}

//
// Derive
//

func Derive(ctx context.Context) (*Basis, error) {
	var err error
	basis := &Basis{}

	if err = env.Parse(basis); err != nil {
		return nil, err
	}

	// Handle chdir first, before other derivations
	if basis.Chdir != "" {
		if err := os.Chdir(basis.Chdir); err != nil {
			return nil, fmt.Errorf("failed to change directory to %s: %w", basis.Chdir, err)
		}
	}

	basis.GitBasis, err = git.Derive()
	if err != nil {
		return nil, err
	}

	basis.CallerBasis, err = caller.Derive(ctx)
	if err != nil {
		return nil, err
	}

	basis.ServiceBasis, err = service.Derive()
	if err != nil {
		return nil, err
	}

	basis.DefaultBasis, err = defaults.Derive()
	if err != nil {
		return nil, err
	}

	basis.ResourceBasis, err = resource.Derive(basis.GitBasis, basis.ServiceBasis)
	if err != nil {
		return nil, err
	}

	basis.EcrBasis, err = ecr.Derive(basis.CallerBasis, basis.GitBasis, basis.ServiceBasis)
	if err != nil {
		return nil, err
	}

	if err = basis.Validate(); err != nil {
		return nil, err
	}

	log.Info().
		Str("owner", basis.GitBasis.Owner).
		Str("repo", basis.GitBasis.Repository).
		Str("branch", basis.GitBasis.Branch).
		Str("sha", truncate(basis.GitBasis.Sha)).
		Msg("git")

	log.Info().
		Str("account", basis.CallerBasis.AccountId).
		Str("region", basis.CallerBasis.AwsConfig.Region).
		Msg("caller")

	log.Info().
		Str("id", basis.EcrBasis.RegistryId).
		Str("region", basis.EcrBasis.RegistryRegion).
		Str("image", basis.EcrBasis.Image).
		Msg("ecr")

	return basis, nil
}

//
// Validations
//

func (b *Basis) Validate() error {
	return v.ValidateStruct(b,
		v.Field(&b.GitBasis),
		v.Field(&b.CallerBasis),
		v.Field(&b.ServiceBasis),
		v.Field(&b.DefaultBasis),
		v.Field(&b.ResourceBasis),
		v.Field(&b.EcrBasis),
	)
}

//
// Accessors
//

func (b *Basis) AwsConfig() aws.Config { return b.CallerBasis.AwsConfig }
func (b *Basis) AccountId() string     { return b.CallerBasis.AccountId }
func (b *Basis) Region() string        { return b.CallerBasis.AwsConfig.Region }

func (b *Basis) Name() string            { return b.ResourceBasis.Name }
func (b *Basis) Path() string            { return b.ResourceBasis.Path }
func (b *Basis) Tags() map[string]string { return b.ResourceBasis.Tags }

func (b *Basis) Image() string          { return b.EcrBasis.Image }
func (b *Basis) RegistryId() string     { return b.EcrBasis.RegistryId }
func (b *Basis) RegistryRegion() string { return b.EcrBasis.RegistryRegion }

func (b *Basis) PolicyTemplate() string { return b.DefaultBasis.Policy }
func (b *Basis) RoleTemplate() string   { return b.DefaultBasis.Role }
func (b *Basis) EnvTemplate() string    { return b.DefaultBasis.Env }

//
// Templating
//

func (b *Basis) Render(input string) (string, error) {
	data := TemplateData{}
	data.Account.Id = b.CallerBasis.AccountId
	data.Account.Region = b.CallerBasis.AwsConfig.Region
	data.Git.Repo = b.GitBasis.Repository
	data.Git.Owner = b.GitBasis.Owner
	data.Git.Branch = b.GitBasis.Branch
	data.Git.Sha = b.GitBasis.Sha
	data.Service.Name = b.ServiceBasis.Name
	data.Resource.Name = b.ResourceBasis.Name
	data.Resource.Path = b.ResourceBasis.Path

	tmpl, err := template.New("template").Parse(input)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

//
// Helpers
//

// truncate shortens git SHA to 7 characters for display
func truncate(s string) string {
	if len(s) <= 7 {
		return s
	}
	return s[:7]
}
