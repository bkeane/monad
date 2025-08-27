package config

import (
	"context"

	"github.com/bkeane/monad/pkg/basis/caller"
	"github.com/bkeane/monad/pkg/basis/defaults"
	"github.com/bkeane/monad/pkg/basis/git"
	"github.com/bkeane/monad/pkg/basis/registry"
	"github.com/bkeane/monad/pkg/basis/resource"
	"github.com/bkeane/monad/pkg/basis/service"
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
	Git() (*git.Basis, error)
	Caller() (*caller.Basis, error)
	Service() (*service.Basis, error)
	Resource() (*resource.Basis, error)
	Registry() (*registry.Basis, error)
	Defaults() (*defaults.Basis, error)
	Render(string) (string, error)
}

//
// Config
//

type Config struct {
	// Basis for lazy initialization
	basis Basis

	// Fields for cached lazy loading and flag definition
	ApiGatewayConfig  *apigateway.Config
	CloudWatchConfig  *cloudwatch.Config
	EventBridgeConfig *eventbridge.Config
	IamConfig         *iam.Config
	LambdaConfig      *lambda.Config
	EcrConfig         *ecr.Config
	VpcConfig         *vpc.Config
}

//
// Derive
//

func Derive(ctx context.Context, basis Basis) (*Config, error) {
	var cfg Config

	cfg.basis = basis

	return &cfg, nil
}

//
// Accessors
//

func (c *Config) Lambda(ctx context.Context) (*lambda.Config, error) {
	var err error

	if c.LambdaConfig == nil {
		c.LambdaConfig, err = lambda.Derive(ctx, c.basis)
		if err != nil {
			return nil, err
		}

		err = c.LambdaConfig.Validate()
		if err != nil {
			return nil, err
		}
	}

	return c.LambdaConfig, nil
}

func (c *Config) ApiGateway(ctx context.Context) (*apigateway.Config, error) {
	var err error

	if c.ApiGatewayConfig == nil {
		c.ApiGatewayConfig, err = apigateway.Derive(ctx, c.basis)
		if err != nil {
			return nil, err
		}

		err = c.ApiGatewayConfig.Validate()
		if err != nil {
			return nil, err
		}
	}

	return c.ApiGatewayConfig, nil
}

func (c *Config) EventBridge(ctx context.Context) (*eventbridge.Config, error) {
	var err error

	if c.EventBridgeConfig == nil {
		c.EventBridgeConfig, err = eventbridge.Derive(ctx, c.basis)
		if err != nil {
			return nil, err
		}

		err = c.EventBridgeConfig.Validate()
		if err != nil {
			return nil, err
		}
	}

	return c.EventBridgeConfig, nil
}

func (c *Config) CloudWatch(ctx context.Context) (*cloudwatch.Config, error) {
	var err error

	if c.CloudWatchConfig == nil {
		c.CloudWatchConfig, err = cloudwatch.Derive(ctx, c.basis)
		if err != nil {
			return nil, err
		}

		err = c.CloudWatchConfig.Validate()
		if err != nil {
			return nil, err
		}
	}

	return c.CloudWatchConfig, nil
}

func (c *Config) Ecr(ctx context.Context) (*ecr.Config, error) {
	var err error

	if c.EcrConfig == nil {
		c.EcrConfig, err = ecr.Derive(ctx, c.basis)
		if err != nil {
			return nil, err
		}

		err = c.EcrConfig.Validate()
		if err != nil {
			return nil, err
		}
	}

	return c.EcrConfig, nil
}

func (c *Config) Iam(ctx context.Context) (*iam.Config, error) {
	var err error

	if c.IamConfig == nil {
		c.IamConfig, err = iam.Derive(ctx, c.basis)
		if err != nil {
			return nil, err
		}

		err = c.IamConfig.Validate()
		if err != nil {
			return nil, err
		}
	}

	return c.IamConfig, nil
}

func (c *Config) Vpc(ctx context.Context) (*vpc.Config, error) {
	var err error

	if c.VpcConfig == nil {
		c.VpcConfig, err = vpc.Derive(ctx, c.basis)
		if err != nil {
			return nil, err
		}

		err = c.VpcConfig.Validate()
		if err != nil {
			return nil, err
		}
	}

	return c.VpcConfig, nil
}
