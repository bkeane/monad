package vpc

import (
	"context"
	"fmt"
	"strings"

	"github.com/bkeane/monad/pkg/basis/caller"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/caarlos0/env/v11"
	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/rs/zerolog/log"
)

type Basis interface {
	Caller() (*caller.Basis, error)
}

//
// Convention
//

type Config struct {
	client              *ec2.Client
	VpcSecurityGroups   []string `env:"MONAD_SECURITY_GROUPS" flag:"--vpc-sg" usage:"VPC security group IDs or names" hint:"name|id"`
	VpcSecurityGroupIds []string
	VpcSubnets          []string `env:"MONAD_SUBNETS" flag:"--vpc-sn" usage:"VPC subnet IDs or names" hint:"name|id"`
	VpcSubnetIds        []string
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

	caller, err := basis.Caller()
	if err != nil {
		return nil, err
	}

	cfg.client = ec2.NewFromConfig(caller.AwsConfig())

	if err := cfg.resolveSecurityGroups(ctx); err != nil {
		return nil, err
	}

	if err := cfg.resolveSubnets(ctx); err != nil {
		return nil, err
	}

	if err = cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate ensures VPC configuration is complete
func (c *Config) Validate() error {
	return v.ValidateStruct(c,
		v.Field(&c.client, v.Required),
		// When security groups are provided, subnets must be provided
		v.Field(&c.VpcSecurityGroups, v.When(len(c.VpcSubnets) != 0, v.Required)),
		// When subnets are provided, security groups must be provided
		v.Field(&c.VpcSubnets, v.When(len(c.VpcSecurityGroups) != 0, v.Required)),
		// When security groups are provided, we must resolve them
		v.Field(&c.VpcSecurityGroupIds, v.When(len(c.VpcSecurityGroups) != 0, v.Required)),
		// When subnets are provided, we must resolve them
		v.Field(&c.VpcSubnets, v.When(len(c.VpcSubnets) != 0, v.Required)),
	)
}

//
// Accessors
//

// Client returns the AWS EC2 service client for VPC operations
func (c *Config) Client() *ec2.Client { return c.client }

// SecurityGroups returns the security group names/IDs from configuration
func (c *Config) SecurityGroups() []string { return c.VpcSecurityGroups }

// SecurityGroupIds returns the resolved security group IDs
func (c *Config) SecurityGroupIds() []string { return c.VpcSecurityGroupIds }

// Subnets returns the subnet names/IDs from configuration
func (c *Config) Subnets() []string { return c.VpcSubnets }

// SubnetIds returns the resolved subnet IDs
func (c *Config) SubnetIds() []string { return c.VpcSubnetIds }

//
// Helpers
//

// resolveSecurityGroups resolves security group names to IDs
func (c *Config) resolveSecurityGroups(ctx context.Context) error {
	for _, nameOrId := range c.VpcSecurityGroups {
		if strings.HasPrefix(nameOrId, "sg-") {
			c.VpcSecurityGroupIds = append(c.VpcSecurityGroupIds, nameOrId)
			continue
		}

		input := &ec2.DescribeSecurityGroupsInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("group-name"),
					Values: []string{nameOrId},
				},
			},
		}

		result, err := c.client.DescribeSecurityGroups(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to lookup security group %s: %w", nameOrId, err)
		}

		for _, sg := range result.SecurityGroups {
			if *sg.GroupName == nameOrId {
				c.VpcSecurityGroupIds = append(c.VpcSecurityGroupIds, *sg.GroupId)
			}
		}
	}

	return nil
}

// resolveSubnets resolves subnet names to IDs
func (c *Config) resolveSubnets(ctx context.Context) error {
	for _, nameOrId := range c.VpcSubnets {
		if strings.HasPrefix(nameOrId, "subnet-") {
			c.VpcSubnetIds = append(c.VpcSubnetIds, nameOrId)
			continue
		}

		input := &ec2.DescribeSubnetsInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("tag:Name"),
					Values: []string{nameOrId},
				},
			},
		}

		result, err := c.client.DescribeSubnets(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to lookup subnet %s: %w", nameOrId, err)
		}

		for _, subnet := range result.Subnets {
			for _, tag := range subnet.Tags {
				if *tag.Key == "Name" && *tag.Value == nameOrId {
					c.VpcSubnetIds = append(c.VpcSubnetIds, *subnet.SubnetId)
				}
			}
		}

		if len(c.VpcSubnetIds) == 0 {
			log.Warn().Str("given", nameOrId).Msg("no subnets found for given")
		}
	}

	return nil
}
