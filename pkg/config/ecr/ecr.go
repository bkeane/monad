package ecr

import (
	"context"
	"strings"

	"github.com/bkeane/monad/internal/registryv2"
	"github.com/bkeane/monad/pkg/basis/caller"
	"github.com/bkeane/monad/pkg/basis/registry"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	v "github.com/go-ozzo/ozzo-validation/v4"
)

type Basis interface {
	Caller() (*caller.Basis, error)
	Registry() (*registry.Basis, error)
}

//
// Convention
//

type Config struct {
	client     *ecr.Client
	registryv2 *registryv2.Client
	caller     *caller.Basis
	registry   *registry.Basis
}

//
// Derive
//

func Derive(ctx context.Context, basis Basis) (*Config, error) {
	var err error
	var cfg Config

	cfg.caller, err = basis.Caller()
	if err != nil {
		return nil, err
	}

	cfg.registry, err = basis.Registry()
	if err != nil {
		return nil, err
	}

	cfg.client = ecr.NewFromConfig(cfg.caller.AwsConfig())

	cfg.registryv2, err = registryv2.InitEcr(ctx, cfg.caller.AwsConfig(), cfg.registry.Id(), cfg.registry.Region())
	if err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

//
// Validations
//

func (c *Config) Validate() error {
	return v.ValidateStruct(c,
		v.Field(&c.registry, v.Required),
		v.Field(&c.registryv2, v.Required),
	)
}

//
// Accessors
//

// Client returns the registry client
func (c *Config) Clients() (*ecr.Client, *registryv2.Client) {
	return c.client, c.registryv2
}

// ImagePath returns the path of the image
func (c *Config) ImagePath() string {
	parts := strings.Split(c.registry.Image(), ":")
	return parts[0]
}

// ImageTag returns the tag of the image
func (c *Config) ImageTag() string {
	parts := strings.Split(c.registry.Image(), ":")
	return parts[1]
}
