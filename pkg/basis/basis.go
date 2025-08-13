package basis

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/bkeane/monad/pkg/basis/caller"
	"github.com/bkeane/monad/pkg/basis/git"
	"github.com/bkeane/monad/pkg/basis/service"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	env "github.com/caarlos0/env/v11"
	v "github.com/go-ozzo/ozzo-validation/v4"
)

//go:embed defaults/*
var defaults embed.FS

//
// Basis
//

type Basis struct {
	awsconfig      aws.Config
	git            *git.Data
	caller         *caller.Data
	service        *service.Data
	image          string `env:"MONAD_IMAGE"`
	registryID     string `env:"MONAD_REGISTRY_ID"`
	registryRegion string `env:"MONAD_REGISTRY_REGION"`
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

func Derive(ctx context.Context, basis *Basis) (*Basis, error) {
	var err error

	if basis == nil {
		basis = &Basis{}
	}

	if err = env.Parse(basis); err != nil {
		return nil, err
	}

	basis.awsconfig, err = config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	basis.git, err = git.Derive()
	if err != nil {
		return nil, err
	}

	basis.service, err = service.Derive()
	if err != nil {
		return nil, err
	}

	basis.caller, err = caller.Derive(ctx, basis.awsconfig)
	if err != nil {
		return nil, err
	}

	if err = basis.Validate(); err != nil {
		return nil, err
	}

	return basis, nil
}

//
// Validations
//

func (b *Basis) Validate() error {
	return v.ValidateStruct(b,
		v.Field(&b.awsconfig),
		v.Field(&b.caller),
		v.Field(&b.git),
		v.Field(&b.service, v.Required),
		// v.Field(&b.image, v.Required),
		// v.Field(&b.registryID, v.Required),
		// v.Field(&b.registryRegion, v.Required),
	)
}

//
// Accessors
//

// AwsConfig returns the aws.Config derived from the defaul credential chain
func (b *Basis) AwsConfig() aws.Config {
	return b.awsconfig
}

// AccountId returns the AWS account ID from sts get caller identity
func (b *Basis) AccountId() string {
	return b.caller.AccountId()
}

// Region returns the AWS region from the default credential chain
func (b *Basis) Region() string {
	return b.caller.Region()
}

// ECR

// RegistryID returns the ECR registry ID
func (b *Basis) RegistryId() string {
	if b.registryID == "" {
		return b.caller.AccountId()
	}

	return b.registryID
}

// RegistryRegion returns the ECR registry region
func (b *Basis) RegistryRegion() string {
	if b.registryRegion == "" {
		return b.caller.Region()
	}

	return b.registryRegion
}

// Image returns full image path + tag: {owner}/{repo}/{service}:{branch}
func (b *Basis) Image() string {
	if b.image == "" {
		return fmt.Sprintf("%s/%s/%s:%s", b.git.Owner(), b.git.Repository(), b.service.Name(), b.git.Branch())
	}

	if !strings.Contains(b.image, ":") {
		return fmt.Sprintf("%s:%s", b.image, b.git.Branch())
	}

	return b.image
}

// Resource Naming

// NamePrefix returns resource name prefix: {repo}-{branch}
func (b *Basis) NamePrefix() string {
	return fmt.Sprintf("%s-%s", b.git.Repository(), b.git.Branch())
}

// Name returns full resource name: {repo}-{branch}-{service}
func (b *Basis) Name() string {
	return fmt.Sprintf("%s-%s", b.NamePrefix(), b.service.Name())
}

// PathPrefix returns resource path prefix: {repo}/{branch}
func (b *Basis) PathPrefix() string {
	return fmt.Sprintf("%s/%s", b.git.Repository(), b.git.Branch())
}

// Path returns full resource path: {repo}/{branch}/{service}
func (b *Basis) Path() string {
	return fmt.Sprintf("%s/%s", b.PathPrefix(), b.service.Name())
}

// Tags returns standardized AWS resource tags
func (b *Basis) Tags() map[string]string {
	return map[string]string{
		"Monad":   "true",
		"Service": b.service.Name(),
		"Owner":   b.git.Owner(),
		"Repo":    b.git.Repository(),
		"Branch":  b.git.Branch(),
		"Sha":     b.git.Sha(),
	}
}

// Documents
func (b *Basis) PolicyTemplate() (string, error) {
	return read("defaults/policy.json.tmpl")
}

func (b *Basis) RoleTemplate() (string, error) {
	return read("defaults/role.json.tmpl")
}

func (b *Basis) EnvTemplate() (string, error) {
	return read("defaults/env.tmpl")
}

//
// Helpers
//

func (b *Basis) Render(input string) (string, error) {
	data := TemplateData{}
	data.Account.Id = b.caller.AccountId()
	data.Account.Region = b.caller.Region()
	data.Git.Repo = b.git.Repository()
	data.Git.Owner = b.git.Owner()
	data.Git.Branch = b.git.Branch()
	data.Git.Sha = b.git.Sha()
	data.Service.Name = b.service.Name()
	data.Resource.Name = b.Name()
	data.Resource.Path = b.Path()

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

func read(path string) (string, error) {
	data, err := defaults.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
