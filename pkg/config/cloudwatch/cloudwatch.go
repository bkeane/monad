package cloudwatch

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	v "github.com/go-ozzo/ozzo-validation/v4"
)

type Basis interface {
	AwsConfig() aws.Config
	AccountId() string
	Region() string
	Path() string
	Tags() map[string]string
}

// Convention

type Config struct {
	basis     Basis
	client    *cloudwatchlogs.Client
	region    string `env:"MONAD_LOG_REGION"`
	retention int32  `env:"MONAD_LOG_RETENTION"`
}

//
// Derive
//

func Derive(ctx context.Context, basis Basis) (*Config, error) {
	var err error
	var cfg Config

	cfg.basis = basis
	cfg.client = cloudwatchlogs.NewFromConfig(basis.AwsConfig())

	if cfg.region == "" {
		cfg.region = basis.Region()
	}

	if cfg.retention == 0 {
		cfg.retention = int32(14)
	}

	if err = cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, err
}

//
// Validate
//

func (c *Config) Validate() error {
	return v.ValidateStruct(c,
		v.Field(&c.client, v.Required),
		v.Field(&c.region, v.Required),
		v.Field(&c.retention, v.Required),
	)
}

// Accessors

// Client returns the AWS CloudWatch service client
func (c *Config) Client() *cloudwatchlogs.Client { return c.client }

// Path returns the CloudWatch log group name for the Lambda function
func (c *Config) Name() string {
	return fmt.Sprintf("/aws/lambda/%s", c.basis.Path())
}

// Arn returns the complete ARN for the CloudWatch log group
func (c *Config) Arn() string {
	return fmt.Sprintf("arn:aws:logs:%s:%s:log-group:%s", c.region, c.basis.AccountId(), c.Name())
}

// LogGroupRetention returns the log retention period in days
func (c *Config) Retention() int32 {
	return c.retention
}

// LogGroupTags returns the CloudWatch log group tags
func (c *Config) Tags() map[string]string {
	return c.basis.Tags()
}
