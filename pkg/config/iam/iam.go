package iam

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/bkeane/monad/pkg/basis/caller"
	"github.com/bkeane/monad/pkg/basis/defaults"
	"github.com/bkeane/monad/pkg/basis/resource"
	"github.com/caarlos0/env/v11"
	v "github.com/go-ozzo/ozzo-validation/v4"
)

type Basis interface {
	Caller() (*caller.Basis, error)
	Defaults() (*defaults.Basis, error)
	Resource() (*resource.Basis, error)
	Render(string) (string, error)
}

//
// Convention
//

type Config struct {
	client            *iam.Client
	IamBoundary       string `env:"MONAD_BOUNDARY_POLICY" flag:"--boundary" usage:"IAM boundary policy ARN or name" hint:"name|arn"`
	IamPolicyPath     string `env:"MONAD_POLICY" flag:"--policy" usage:"IAM policy template file path" hint:"name|arn"`
	IamPolicyTemplate string
	IamPolicyDocument string
	IamRolePath       string `env:"MONAD_ROLE" flag:"--role" usage:"IAM role template file path" hint:"name|arn"`
	IamRoleTemplate   string
	IamRoleDocument   string
	caller            *caller.Basis
	resource          *resource.Basis
	defaults          *defaults.Basis
}

//
// Derive
//

func Derive(ctx context.Context, basis Basis) (*Config, error) {
	var err error
	var cfg Config

	// Parse environment variables into struct fields
	if err = env.Parse(&cfg); err != nil {
		return nil, err
	}

	cfg.caller, err = basis.Caller()
	if err != nil {
		return nil, err
	}

	cfg.defaults, err = basis.Defaults()
	if err != nil {
		return nil, err
	}

	cfg.resource, err = basis.Resource()
	if err != nil {
		return nil, err
	}

	cfg.client = iam.NewFromConfig(cfg.caller.AwsConfig())

	// Policy derivation
	if cfg.IamPolicyPath == "" {
		cfg.IamPolicyTemplate = cfg.defaults.PolicyTemplate()

	} else {
		bytes, err := os.ReadFile(cfg.IamPolicyPath)
		if err != nil {
			return nil, err
		}
		cfg.IamPolicyTemplate = string(bytes)
	}

	cfg.IamPolicyDocument, err = basis.Render(cfg.IamPolicyTemplate)
	if err != nil {
		return nil, err
	}

	// Role derivation
	if cfg.IamRolePath == "" {
		cfg.IamRoleTemplate = cfg.defaults.RoleTemplate()

	} else {
		bytes, err := os.ReadFile(cfg.IamRolePath)
		if err != nil {
			return nil, err
		}
		cfg.IamRoleTemplate = string(bytes)
	}

	cfg.IamRoleDocument, err = basis.Render(cfg.IamRoleTemplate)
	if err != nil {
		return nil, err
	}

	if err = cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

//
// Validations
//

func (c *Config) Validate() error {
	return v.ValidateStruct(c,
		v.Field(&c.client, v.Required),
		v.Field(&c.IamPolicyDocument, v.Required),
		v.Field(&c.IamRoleDocument, v.Required),
		v.Field(&c.caller, v.Required),
		v.Field(&c.resource, v.Required),
		v.Field(&c.defaults, v.Required),
	)
}

//
// Accessors
//

// Client returns the AWS IAM service client
func (c *Config) Client() *iam.Client { return c.client }

// RoleName returns the IAM role name using axiom resource naming
func (c *Config) RoleName() string {
	return c.resource.Name()
}

// RoleArn returns the complete ARN for the IAM role
func (c *Config) RoleArn() string {
	return fmt.Sprintf("arn:aws:iam::%s:role/%s", c.caller.AccountId(), c.RoleName())
}

// RoleDocument returns the IAM assume role document
func (c *Config) RoleDocument() string {
	return c.IamRoleDocument
}

// PolicyName returns the IAM policy name using axiom resource naming
func (c *Config) PolicyName() string {
	return c.resource.Name()
}

// PolicyArn returns the complete ARN for the IAM policy
func (c *Config) PolicyArn() string {
	return fmt.Sprintf("arn:aws:iam::%s:policy/%s", c.caller.AccountId(), c.PolicyName())
}

// PolicyDocument returns the IAM policy document
func (c *Config) PolicyDocument() string {
	return c.IamPolicyDocument
}

// BoundaryPolicyName returns the boundary policy name (extracted from ARN if provided)
func (c *Config) BoundaryPolicyName() string {
	if strings.HasPrefix(c.IamBoundary, "arn:aws:iam::") {
		return strings.Split(c.IamBoundary, ":policy/")[1]
	}
	return c.IamBoundary
}

// BoundaryPolicyArn returns the complete ARN for the boundary policy
func (c *Config) BoundaryPolicyArn() string {
	if c.IamBoundary == "" {
		return c.IamBoundary
	}

	if strings.HasPrefix(c.IamBoundary, "arn:aws:iam::") {
		return c.IamBoundary
	}

	return fmt.Sprintf("arn:aws:iam::%s:policy/%s", c.caller.AccountId(), c.IamBoundary)
}

// EniRoleName returns the standard AWS Lambda VPC execution role name
func (c *Config) EniRoleName() string {
	return "AWSLambdaVPCAccessExecutionRole"
}

// EniRoleArn returns the complete ARN for the Lambda VPC execution role
func (c *Config) EniRoleArn() string {
	return fmt.Sprintf("arn:aws:iam::%s:role/%s", c.caller.AccountId(), c.EniRoleName())
}

// EniRolePolicyArn returns the AWS managed policy ARN for Lambda VPC access
func (c *Config) EniRolePolicyArn() string {
	return "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
}

// Tags returns standardized IAM resource tags
func (c *Config) Tags() []iamtypes.Tag {
	var tags []iamtypes.Tag
	for key, value := range c.resource.Tags() {
		tags = append(tags, iamtypes.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}
	return tags
}
