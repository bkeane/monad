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

	v "github.com/go-ozzo/ozzo-validation/v4"
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
	ApiGatewayConfig  *apigateway.Config
	CloudwatchConfig  *cloudwatch.Config
	EventbridgeConfig *eventbridge.Config
	IamConfig         *iam.Config
	LambdaConfig      *lambda.Config
	EcrConfig         *ecr.Config
	VpcConfig         *vpc.Config
}

//
// Derive
//

func Derive(ctx context.Context, basis Basis) (*Config, error) {
	var err error
	var cfg Config

	if err = basis.Validate(); err != nil {
		return nil, err
	}

	cfg.ApiGatewayConfig, err = apigateway.Derive(ctx, basis)
	if err != nil {
		return nil, err
	}

	cfg.CloudwatchConfig, err = cloudwatch.Derive(ctx, basis)
	if err != nil {
		return nil, err
	}

	cfg.EventbridgeConfig, err = eventbridge.Derive(ctx, basis)
	if err != nil {
		return nil, err
	}

	cfg.IamConfig, err = iam.Derive(ctx, basis)
	if err != nil {
		return nil, err
	}

	cfg.LambdaConfig, err = lambda.Derive(ctx, basis)
	if err != nil {
		return nil, err
	}

	cfg.EcrConfig, err = ecr.Derive(ctx, basis)
	if err != nil {
		return nil, err
	}

	cfg.VpcConfig, err = vpc.Derive(ctx, basis)
	if err != nil {
		return nil, err
	}

	err = cfg.Validate()
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

//
// Validate
//

func (d *Config) Validate() error {
	return v.ValidateStruct(d,
		v.Field(&d.ApiGatewayConfig),
		v.Field(&d.CloudwatchConfig),
		v.Field(&d.EcrConfig),
		v.Field(&d.EventbridgeConfig),
		v.Field(&d.IamConfig),
		v.Field(&d.LambdaConfig),
		v.Field(&d.VpcConfig),
	)
}

//
// Accessors
//

func (c *Config) Lambda() *lambda.Config {
	return c.LambdaConfig
}

func (c *Config) ApiGateway() *apigateway.Config {
	return c.ApiGatewayConfig
}

func (c *Config) EventBridge() *eventbridge.Config {
	return c.EventbridgeConfig
}

func (c *Config) CloudWatch() *cloudwatch.Config {
	return c.CloudwatchConfig
}

func (c *Config) Ecr() *ecr.Config {
	return c.EcrConfig
}

func (c *Config) Iam() *iam.Config {
	return c.IamConfig
}

func (c *Config) Vpc() *vpc.Config {
	return c.VpcConfig
}
