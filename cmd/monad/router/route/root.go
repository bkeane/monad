package route

import (
	"context"

	"github.com/bkeane/monad/internal/logging"
	"github.com/bkeane/monad/pkg/model/axiom"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

type Root struct {
	axiom.Axiom
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

	// Load AWS config first
	r.AwsConfig, err = config.LoadDefaultConfig(ctx,
		config.WithLogger(logging.AwsConfig(ctx)),
		config.WithClientLogMode(aws.LogRetries))
	if err != nil {
		return err
	}

	// Process axiom (handles git, service, caller initialization)
	if err = r.Axiom.Process(ctx, r.AwsConfig); err != nil {
		return err
	}

	return nil
}
