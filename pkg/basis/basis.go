package basis

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/bkeane/monad/pkg/basis/caller"
	"github.com/bkeane/monad/pkg/basis/defaults"
	"github.com/bkeane/monad/pkg/basis/git"
	"github.com/bkeane/monad/pkg/basis/registry"
	"github.com/bkeane/monad/pkg/basis/resource"
	"github.com/bkeane/monad/pkg/basis/service"

	env "github.com/caarlos0/env/v11"
	"github.com/charmbracelet/lipgloss/table"
)

//
// Basis
//

type Basis struct {
	Chdir         string `env:"MONAD_CHDIR" flag:"--chdir" usage:"Change working directory" hint:"path"`
	GitBasis      *git.Basis
	CallerBasis   *caller.Basis
	ServiceBasis  *service.Basis
	RegistryBasis *registry.Basis
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

	return basis, nil
}

//
// Accessors
//

func (b *Basis) Git() (*git.Basis, error) {
	var err error

	if b.GitBasis == nil {
		b.GitBasis, err = git.Derive()
		if err != nil {
			return nil, err
		}
	}

	return b.GitBasis, nil
}

func (b *Basis) Caller() (*caller.Basis, error) {
	var err error

	if b.CallerBasis == nil {
		ctx := context.Background()
		b.CallerBasis, err = caller.Derive(ctx)
		if err != nil {
			return nil, err
		}
	}

	return b.CallerBasis, nil
}

func (b *Basis) Service() (*service.Basis, error) {
	var err error

	if b.ServiceBasis == nil {
		b.ServiceBasis, err = service.Derive()
		if err != nil {
			return nil, err
		}
	}

	return b.ServiceBasis, nil
}

func (b *Basis) Resource() (*resource.Basis, error) {
	if b.ResourceBasis == nil {
		gitBasis, err := b.Git()
		if err != nil {
			return nil, err
		}

		serviceBasis, err := b.Service()
		if err != nil {
			return nil, err
		}

		b.ResourceBasis, err = resource.Derive(gitBasis, serviceBasis)
		if err != nil {
			return nil, err
		}
	}

	return b.ResourceBasis, nil
}

func (b *Basis) Registry() (*registry.Basis, error) {
	if b.RegistryBasis == nil {
		callerBasis, err := b.Caller()
		if err != nil {
			return nil, err
		}

		gitBasis, err := b.Git()
		if err != nil {
			return nil, err
		}

		serviceBasis, err := b.Service()
		if err != nil {
			return nil, err
		}

		b.RegistryBasis, err = registry.Derive(callerBasis, gitBasis, serviceBasis)
		if err != nil {
			return nil, err
		}
	}

	return b.RegistryBasis, nil
}

func (b *Basis) Defaults() (*defaults.Basis, error) {
	var err error

	if b.DefaultBasis == nil {
		b.DefaultBasis, err = defaults.Derive()
		if err != nil {
			return nil, err
		}
	}

	return b.DefaultBasis, nil
}

//
// Templating
//

func (b *Basis) Render(input string) (string, error) {
	data := TemplateData{}

	caller, err := b.Caller()
	if err != nil {
		return "", err
	}

	git, err := b.Git()
	if err != nil {
		return "", err
	}

	service, err := b.Service()
	if err != nil {
		return "", err
	}

	resource, err := b.Resource()
	if err != nil {
		return "", err
	}

	data.Account.Id = caller.AccountId()
	data.Account.Region = caller.AwsConfig().Region
	data.Git.Repo = git.Repo()
	data.Git.Owner = git.Owner()
	data.Git.Branch = git.Branch()
	data.Git.Sha = git.Sha()
	data.Service.Name = service.Name()
	data.Resource.Name = resource.Name()
	data.Resource.Path = resource.Path()

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

func (b *Basis) Table() (string, error) {
	vars := []string{
		"{{.Account.Id}}",
		"{{.Account.Region}}",
		"{{.Git.Repo}}",
		"{{.Git.Owner}}",
		"{{.Git.Branch}}",
		"{{.Git.Sha}}",
		"{{.Service.Name}}",
		"{{.Resource.Name}}",
		"{{.Resource.Path}}",
	}

	tbl := table.New()
	tbl.Headers("Template", "Value")

	for _, v := range vars {
		val, err := b.Render(v)
		if err != nil {
			return "", fmt.Errorf("failed to render %s: %w", v, err)
		}
		tbl.Row(v, strings.TrimSpace(val))
	}

	return tbl.Render(), nil
}
