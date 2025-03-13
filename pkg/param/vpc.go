package param

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	v "github.com/go-ozzo/ozzo-validation/v4"
)

type Vpc struct {
	Client           *ec2.Client `arg:"-" json:"-"`
	SecurityGroups   []string    `arg:"--vpc-sg,env:MONAD_SECURITY_GROUPS" placeholder:"id|name" help:"vpc-default,sg-456... [default: []]"`
	Subnets          []string    `arg:"--vpc-sn,env:MONAD_SUBNETS" placeholder:"id|name" help:"private,subnet-456... [default: []]"`
	SecurityGroupIds []string    `arg:"-"`
	SubnetIds        []string    `arg:"-"`
}

func (c *Vpc) Validate(ctx context.Context, awsconfig aws.Config) error {
	c.Client = ec2.NewFromConfig(awsconfig)

	if err := c.ResolveSecurityGroups(ctx); err != nil {
		return err
	}

	if err := c.ResolveSubnets(ctx); err != nil {
		return err
	}

	return v.ValidateStruct(c,
		// When security groups are provided, subnets must be provided
		v.Field(&c.SecurityGroupIds, v.When(len(c.SubnetIds) != 0, v.Required)),
		// When subnets are provided, security groups must be provided
		v.Field(&c.SubnetIds, v.When(len(c.SecurityGroups) != 0, v.Required)),
		// When security groups are provided, we must resolve them
		v.Field(&c.SecurityGroupIds, v.When(len(c.SecurityGroups) != 0, v.Required)),
		// When subnets are provided, we must resolve them
		v.Field(&c.SubnetIds, v.When(len(c.Subnets) != 0, v.Required)),
	)
}

func (c *Vpc) ResolveSecurityGroups(ctx context.Context) error {
	for _, nameOrId := range c.SecurityGroups {
		if strings.HasPrefix(nameOrId, "sg-") {
			c.SecurityGroupIds = append(c.SecurityGroupIds, nameOrId)
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

		result, err := c.Client.DescribeSecurityGroups(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to lookup security group %s: %w", nameOrId, err)
		}

		for _, sg := range result.SecurityGroups {
			if *sg.GroupName == nameOrId {
				c.SecurityGroupIds = append(c.SecurityGroupIds, *sg.GroupId)
			}
		}
	}

	return nil
}

func (c *Vpc) ResolveSubnets(ctx context.Context) error {
	for _, nameOrId := range c.Subnets {
		if strings.HasPrefix(nameOrId, "subnet-") {
			c.SubnetIds = append(c.SubnetIds, nameOrId)
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

		result, err := c.Client.DescribeSubnets(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to lookup subnet %s: %w", nameOrId, err)
		}

		for _, subnet := range result.Subnets {
			for _, tag := range subnet.Tags {
				if *tag.Key == "Name" && *tag.Value == nameOrId {
					c.SubnetIds = append(c.SubnetIds, *subnet.SubnetId)
				}
			}
		}
	}

	return nil
}
