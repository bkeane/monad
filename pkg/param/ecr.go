package param

import (
	"context"

	"github.com/bkeane/monad/pkg/registry"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	v "github.com/go-ozzo/ozzo-validation/v4"
)

type Registry struct {
	Client *registry.Client `arg:"-" json:"-"`
	Id     string           `arg:"--ecr-id,env:MONAD_REGISTRY_ID" placeholder:"id" help:"ecr registry id" default:"caller-account-id"`
	Region string           `arg:"--ecr-region,env:MONAD_REGISTRY_REGION" placeholder:"name" help:"ecr registry region" default:"caller-region"`
}

func (r *Registry) Validate(ctx context.Context, awsconfig aws.Config) error {
	var err error

	if r.Id == "" {
		ecrc := ecr.NewFromConfig(awsconfig)
		input := &ecr.DescribeRegistryInput{}
		output, err := ecrc.DescribeRegistry(ctx, input)
		if err != nil {
			return err
		}

		r.Id = *output.RegistryId
	}

	if r.Region == "" {
		r.Region = awsconfig.Region
	}

	r.Client, err = registry.InitEcr(ctx, awsconfig, r.Id, r.Region)
	if err != nil {
		return err
	}

	return v.ValidateStruct(r,
		v.Field(&r.Id, v.Required),
		v.Field(&r.Region, v.Required),
	)
}
