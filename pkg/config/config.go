package config

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/bkeane/monad/pkg/config/apigateway"
	"github.com/bkeane/monad/pkg/config/cloudwatch"
	"github.com/bkeane/monad/pkg/config/ecr"
	"github.com/bkeane/monad/pkg/config/eventbridge"
	"github.com/bkeane/monad/pkg/config/iam"
	"github.com/bkeane/monad/pkg/config/lambda"
	"github.com/bkeane/monad/pkg/config/vpc"
)

//
// Dependencies
//

type Basis interface {
	AwsConfig() aws.Config
	AccountId() string
	RegistryId() string
	RegistryRegion() string
	Region() string
	Name() string
	Path() string
	Image() string
	Tags() map[string]string
	PolicyTemplate() string
	RoleTemplate() string
	EnvTemplate() string
	Render(string) (string, error)
	Validate() error
}

//
// Config
//

type Config struct {
	// Private fields for lazy loading
	apigateway  *apigateway.Config
	cloudwatch  *cloudwatch.Config
	eventbridge *eventbridge.Config
	iam         *iam.Config
	lambda      *lambda.Config
	ecr         *ecr.Config
	vpc         *vpc.Config

	// Basis for lazy initialization
	basis Basis
}

//
// Derive
//

func Derive(ctx context.Context, basis Basis) (*Config, error) {
	var err error
	var cfg Config

	cfg.basis = basis

	if err = cfg.basis.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

//
// Accessors
//

func (c *Config) Lambda(ctx context.Context) (*lambda.Config, error) {
	var err error

	if c.lambda == nil {
		c.lambda, err = lambda.Derive(ctx, c.basis)
		if err != nil {
			return nil, err
		}

		err = c.lambda.Validate()
		if err != nil {
			return nil, err
		}
	}

	return c.lambda, nil
}

func (c *Config) ApiGateway(ctx context.Context) (*apigateway.Config, error) {
	var err error

	if c.apigateway == nil {
		c.apigateway, err = apigateway.Derive(ctx, c.basis)
		if err != nil {
			return nil, err
		}

		err = c.apigateway.Validate()
		if err != nil {
			return nil, err
		}
	}

	return c.apigateway, nil
}

func (c *Config) EventBridge(ctx context.Context) (*eventbridge.Config, error) {
	var err error

	if c.eventbridge == nil {
		c.eventbridge, err = eventbridge.Derive(ctx, c.basis)
		if err != nil {
			return nil, err
		}

		err = c.eventbridge.Validate()
		if err != nil {
			return nil, err
		}
	}

	return c.eventbridge, nil
}

func (c *Config) CloudWatch(ctx context.Context) (*cloudwatch.Config, error) {
	var err error

	if c.cloudwatch == nil {
		c.cloudwatch, err = cloudwatch.Derive(ctx, c.basis)
		if err != nil {
			return nil, err
		}

		err = c.cloudwatch.Validate()
		if err != nil {
			return nil, err
		}
	}

	return c.cloudwatch, nil
}

func (c *Config) Ecr(ctx context.Context) (*ecr.Config, error) {
	var err error

	if c.ecr == nil {
		c.ecr, err = ecr.Derive(ctx, c.basis)
		if err != nil {
			return nil, err
		}

		err = c.ecr.Validate()
		if err != nil {
			return nil, err
		}
	}

	return c.ecr, nil
}

func (c *Config) Iam(ctx context.Context) (*iam.Config, error) {
	var err error

	if c.iam == nil {
		c.iam, err = iam.Derive(ctx, c.basis)
		if err != nil {
			return nil, err
		}

		err = c.iam.Validate()
		if err != nil {
			return nil, err
		}
	}

	return c.iam, nil
}

func (c *Config) Vpc(ctx context.Context) (*vpc.Config, error) {
	var err error

	if c.vpc == nil {
		c.vpc, err = vpc.Derive(ctx, c.basis)
		if err != nil {
			return nil, err
		}

		err = c.vpc.Validate()
		if err != nil {
			return nil, err
		}
	}

	return c.vpc, nil
}
