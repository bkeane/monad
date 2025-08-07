package config

import (
	"context"

	"github.com/bkeane/monad/pkg/config/apigateway"
	"github.com/bkeane/monad/pkg/config/cloudwatch"
	"github.com/bkeane/monad/pkg/config/ecr"
	"github.com/bkeane/monad/pkg/config/eventbridge"
	"github.com/bkeane/monad/pkg/config/iam"
	"github.com/bkeane/monad/pkg/config/lambda"
	"github.com/bkeane/monad/pkg/config/vpc"

	"github.com/aws/aws-sdk-go-v2/aws"
	v "github.com/go-ozzo/ozzo-validation/v4"
)

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
	PolicyDocument() (string, error)
	RoleDocument() (string, error)
	EnvDocument() (string, error)
	RuleDocument() (string, error)
	Validate() error
}

//
// Config
//

type Config struct {
	apigateway  *apigateway.Config
	cloudwatch  *cloudwatch.Config
	eventbridge *eventbridge.Config
	iam         *iam.Config
	lambda      *lambda.Config
	ecr         *ecr.Config
	vpc         *vpc.Config
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

	cfg.apigateway, err = apigateway.Derive(ctx, basis)
	if err != nil {
		return nil, err
	}

	cfg.cloudwatch, err = cloudwatch.Derive(ctx, basis)
	if err != nil {
		return nil, err
	}

	cfg.eventbridge, err = eventbridge.Derive(ctx, basis)
	if err != nil {
		return nil, err
	}

	cfg.iam, err = iam.Derive(ctx, basis)
	if err != nil {
		return nil, err
	}

	cfg.lambda, err = lambda.Derive(ctx, basis)
	if err != nil {
		return nil, err
	}

	cfg.ecr, err = ecr.Derive(ctx, basis)
	if err != nil {
		return nil, err
	}

	cfg.vpc, err = vpc.Derive(ctx, basis)
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
		v.Field(&d.apigateway),
		v.Field(&d.cloudwatch),
		v.Field(&d.ecr),
		v.Field(&d.eventbridge),
		v.Field(&d.iam),
		v.Field(&d.lambda),
		v.Field(&d.vpc),
	)
}

//
// Accessors
//

func (c *Config) Lambda() *lambda.Config {
	return c.lambda
}

func (c *Config) ApiGateway() *apigateway.Config {
	return c.apigateway
}

func (c *Config) EventBridge() *eventbridge.Config {
	return c.eventbridge
}

func (c *Config) CloudWatch() *cloudwatch.Config {
	return c.cloudwatch
}

func (c *Config) Ecr() *ecr.Config {
	return c.ecr
}

func (c *Config) Iam() *iam.Config {
	return c.iam
}

func (c *Config) Vpc() *vpc.Config {
	return c.vpc
}
