package route

import (
	"context"

	"github.com/bkeane/monad/internal/logging"
	"github.com/bkeane/monad/pkg/param"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

type Root struct {
	param.Git
	param.Service
	AwsConfig aws.Config `arg:"-" json:"-"`
	Deploy    *Deploy    `arg:"subcommand:deploy" help:"deploy a service"`
	Destroy   *Destroy   `arg:"subcommand:destroy" help:"destroy a service"`
	List      *List      `arg:"subcommand:list" help:"list services"`
	Ecr       *Ecr       `arg:"subcommand:ecr" help:"ecr commands"`
	Init      *Scaffold  `arg:"subcommand:init" help:"initialize a service"`
	Data      *Data      `arg:"subcommand:data" help:"contextual template data"`
}

func (r *Root) Validate(ctx context.Context) error {
	var err error

	if err = r.Git.Validate(); err != nil {
		return err
	}

	if err = r.Service.Validate(r.Git); err != nil {
		return err
	}

	r.AwsConfig, err = config.LoadDefaultConfig(ctx,
		config.WithLogger(logging.AwsConfig(ctx)),
		config.WithClientLogMode(aws.LogRetries))

	if err != nil {
		return err
	}

	return nil
}
