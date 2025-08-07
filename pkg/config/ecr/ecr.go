package ecr

import (
	"context"
	"strings"

	"github.com/bkeane/monad/internal/registryv2"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	v "github.com/go-ozzo/ozzo-validation/v4"
)

type Basis interface {
	AwsConfig() aws.Config
	RegistryId() string
	RegistryRegion() string
	Image() string
}

//
// Convention
//

type Config struct {
	basis      Basis
	ecr        *ecr.Client
	registryv2 *registryv2.Client
}

//
// Derive
//

func Derive(ctx context.Context, basis Basis) (*Config, error) {
	var err error
	var cfg Config

	cfg.basis = basis
	cfg.ecr = ecr.NewFromConfig(basis.AwsConfig())

	cfg.registryv2, err = registryv2.InitEcr(ctx, basis.AwsConfig(), basis.RegistryId(), basis.RegistryRegion())
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
		v.Field(&c.ecr, v.Required),
		v.Field(&c.registryv2, v.Required),
	)
}

//
// Accessors
//

// Client returns the registry client
func (c *Config) Clients() (*ecr.Client, *registryv2.Client) {
	return c.ecr, c.registryv2
}

// ImagePath returns the path of the image
func (c *Config) ImagePath() string {
	parts := strings.Split(c.basis.Image(), ":")
	return parts[0]
}

// ImageTag returns the tag of the image
func (c *Config) ImageTag() string {
	parts := strings.Split(c.basis.Image(), ":")
	return parts[1]
}
