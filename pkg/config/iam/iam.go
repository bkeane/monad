package iam

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/caarlos0/env/v11"
	v "github.com/go-ozzo/ozzo-validation/v4"
)

type Basis interface {
	AwsConfig() aws.Config
	AccountId() string
	Name() string
	RoleTemplate() string
	PolicyTemplate() string
	Render(string) (string, error)
	Tags() map[string]string
}

//
// Convention
//

type Config struct {
	basis          Basis
	client         *iam.Client
	Boundary       string `env:"MONAD_BOUNDARY_POLICY" flag:"--boundary" usage:"IAM boundary policy ARN or name"`
	PolicyPath     string `env:"MONAD_POLICY" flag:"--policy" usage:"IAM policy template file path"`
	policyTemplate string
	policyDocument string
	RolePath       string `env:"MONAD_ROLE" flag:"--role" usage:"IAM role template file path"`
	roleTemplate   string
	roleDocument   string
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

	cfg.basis = basis
	cfg.client = iam.NewFromConfig(basis.AwsConfig())

	// Policy derivation
	if cfg.PolicyPath == "" {
		cfg.policyTemplate = basis.PolicyTemplate()

	} else {
		bytes, err := os.ReadFile(cfg.PolicyPath)
		if err != nil {
			return nil, err
		}
		cfg.policyTemplate = string(bytes)
	}

	cfg.policyDocument, err = basis.Render(cfg.policyTemplate)
	if err != nil {
		return nil, err
	}

	// Role derivation
	if cfg.RolePath == "" {
		cfg.roleTemplate = basis.RoleTemplate()

	} else {
		bytes, err := os.ReadFile(cfg.RolePath)
		if err != nil {
			return nil, err
		}
		cfg.roleTemplate = string(bytes)
	}

	cfg.roleDocument, err = basis.Render(cfg.roleTemplate)
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
		v.Field(&c.basis, v.Required),
		v.Field(&c.client, v.Required),
		v.Field(&c.policyDocument, v.Required),
		v.Field(&c.roleDocument, v.Required),
	)
}

//
// Accessors
//

// Client returns the AWS IAM service client
func (c *Config) Client() *iam.Client { return c.client }

// RoleName returns the IAM role name using axiom resource naming
func (c *Config) RoleName() string {
	return c.basis.Name()
}

// RoleArn returns the complete ARN for the IAM role
func (c *Config) RoleArn() string {
	return fmt.Sprintf("arn:aws:iam::%s:role/%s", c.basis.AccountId(), c.RoleName())
}

// RoleDocument returns the IAM assume role document
func (c *Config) RoleDocument() string {
	return c.roleDocument
}

// PolicyName returns the IAM policy name using axiom resource naming
func (c *Config) PolicyName() string {
	return c.basis.Name()
}

// PolicyArn returns the complete ARN for the IAM policy
func (c *Config) PolicyArn() string {
	return fmt.Sprintf("arn:aws:iam::%s:policy/%s", c.basis.AccountId(), c.PolicyName())
}

// PolicyDocument returns the IAM policy document
func (c *Config) PolicyDocument() string {
	return c.policyDocument
}

// BoundaryPolicyName returns the boundary policy name (extracted from ARN if provided)
func (c *Config) BoundaryPolicyName() string {
	if strings.HasPrefix(c.Boundary, "arn:aws:iam::") {
		return strings.Split(c.Boundary, ":policy/")[1]
	}
	return c.Boundary
}

// BoundaryPolicyArn returns the complete ARN for the boundary policy
func (c *Config) BoundaryPolicyArn() string {
	if c.Boundary == "" {
		return c.Boundary
	}

	if strings.HasPrefix(c.Boundary, "arn:aws:iam::") {
		return c.Boundary
	}

	return fmt.Sprintf("arn:aws:iam::%s:policy/%s", c.basis.AccountId(), c.Boundary)
}

// EniRoleName returns the standard AWS Lambda VPC execution role name
func (c *Config) EniRoleName() string {
	return "AWSLambdaVPCAccessExecutionRole"
}

// EniRoleArn returns the complete ARN for the Lambda VPC execution role
func (c *Config) EniRoleArn() string {
	return fmt.Sprintf("arn:aws:iam::%s:role/%s", c.basis.AccountId(), c.EniRoleName())
}

// EniRolePolicyArn returns the AWS managed policy ARN for Lambda VPC access
func (c *Config) EniRolePolicyArn() string {
	return "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
}

// Tags returns standardized IAM resource tags
func (c *Config) Tags() []iamtypes.Tag {
	var tags []iamtypes.Tag
	for key, value := range c.basis.Tags() {
		tags = append(tags, iamtypes.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}
	return tags
}
