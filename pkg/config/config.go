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
	// Fields for cached lazy loading and flag definition
	ApiGatewayCfg  *apigateway.Config
	CloudWatchCfg  *cloudwatch.Config
	EventBridgeCfg *eventbridge.Config
	IamCfg         *iam.Config
	LambdaCfg      *lambda.Config
	EcrCfg         *ecr.Config
	VpcCfg         *vpc.Config

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

	if c.LambdaCfg == nil {
		c.LambdaCfg, err = lambda.Derive(ctx, c.basis)
		if err != nil {
			return nil, err
		}

		err = c.LambdaCfg.Validate()
		if err != nil {
			return nil, err
		}
	}

	return c.LambdaCfg, nil
}

func (c *Config) ApiGateway(ctx context.Context) (*apigateway.Config, error) {
	var err error

	if c.ApiGatewayCfg == nil {
		c.ApiGatewayCfg, err = apigateway.Derive(ctx, c.basis)
		if err != nil {
			return nil, err
		}

		err = c.ApiGatewayCfg.Validate()
		if err != nil {
			return nil, err
		}
	}

	return c.ApiGatewayCfg, nil
}

func (c *Config) EventBridge(ctx context.Context) (*eventbridge.Config, error) {
	var err error

	if c.EventBridgeCfg == nil {
		c.EventBridgeCfg, err = eventbridge.Derive(ctx, c.basis)
		if err != nil {
			return nil, err
		}

		err = c.EventBridgeCfg.Validate()
		if err != nil {
			return nil, err
		}
	}

	return c.EventBridgeCfg, nil
}

func (c *Config) CloudWatch(ctx context.Context) (*cloudwatch.Config, error) {
	var err error

	if c.CloudWatchCfg == nil {
		c.CloudWatchCfg, err = cloudwatch.Derive(ctx, c.basis)
		if err != nil {
			return nil, err
		}

		err = c.CloudWatchCfg.Validate()
		if err != nil {
			return nil, err
		}
	}

	return c.CloudWatchCfg, nil
}

func (c *Config) Ecr(ctx context.Context) (*ecr.Config, error) {
	var err error

	if c.EcrCfg == nil {
		c.EcrCfg, err = ecr.Derive(ctx, c.basis)
		if err != nil {
			return nil, err
		}

		err = c.EcrCfg.Validate()
		if err != nil {
			return nil, err
		}
	}

	return c.EcrCfg, nil
}

func (c *Config) Iam(ctx context.Context) (*iam.Config, error) {
	var err error

	if c.IamCfg == nil {
		c.IamCfg, err = iam.Derive(ctx, c.basis)
		if err != nil {
			return nil, err
		}

		err = c.IamCfg.Validate()
		if err != nil {
			return nil, err
		}
	}

	return c.IamCfg, nil
}

func (c *Config) Vpc(ctx context.Context) (*vpc.Config, error) {
	var err error

	if c.VpcCfg == nil {
		c.VpcCfg, err = vpc.Derive(ctx, c.basis)
		if err != nil {
			return nil, err
		}

		err = c.VpcCfg.Validate()
		if err != nil {
			return nil, err
		}
	}

	return c.VpcCfg, nil
}
