package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bkeane/monad/cmd/monad/deploy"
	"github.com/bkeane/monad/cmd/monad/destroy"
	"github.com/bkeane/monad/cmd/monad/encode"
	"github.com/bkeane/monad/cmd/monad/listen"
	"github.com/bkeane/monad/cmd/monad/scaffold"
	"github.com/bkeane/monad/internal/logging"

	"github.com/alexflint/go-arg"
	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.WarnLevel)
	if value, ok := os.LookupEnv("LOG_LEVEL"); ok {
		if level, err := zerolog.ParseLevel(value); err == nil {
			zerolog.SetGlobalLevel(level)
		}
	}
}

type Root struct {
	Init    *scaffold.Root `arg:"subcommand:init" help:"create a service"`
	Compose *encode.Root   `arg:"subcommand:encode" help:"encode a service"`
	Deploy  *deploy.Root   `arg:"subcommand:deploy" help:"deploy a service"`
	Destroy *destroy.Root  `arg:"subcommand:destroy" help:"destroy a service"`
	Listen  *listen.Root   `arg:"subcommand:listen" help:"listen for events"`
}

func (r *Root) Route(ctx context.Context, awsconfig aws.Config) (*string, error) {
	switch {
	case r.Init != nil:
		return r.Init.Route(ctx)

	case r.Compose != nil:
		return r.Compose.Route(ctx, awsconfig)

	case r.Deploy != nil:
		return r.Deploy.Route(ctx, awsconfig)

	case r.Destroy != nil:
		return r.Destroy.Route(ctx, awsconfig)

	case r.Listen != nil:
		return r.Listen.Route(ctx, awsconfig)

	}

	return nil, nil
}

func main() {
	ctx := context.Background()

	var root Root
	var output *string
	var err error

	arg.MustParse(&root)

	awsconfig, err := awscfg.LoadDefaultConfig(ctx,
		awscfg.WithLogger(logging.AwsConfig(ctx)),
		awscfg.WithClientLogMode(aws.LogRetries))

	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	output, err = root.Route(ctx, awsconfig)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	if output != nil {
		fmt.Println(*output)
	}
}
