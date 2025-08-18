package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bkeane/monad/pkg/basis"
	"github.com/bkeane/monad/pkg/flag"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	if value, ok := os.LookupEnv("LOG_LEVEL"); ok {
		if level, err := zerolog.ParseLevel(value); err == nil {
			zerolog.SetGlobalLevel(level)
		}
	}
}

func main() {
	cmd := &cli.Command{
		Name:  "monad",
		Usage: "management plane",
		Flags: flag.Parse(basis.Basis{}),
		Commands: []*cli.Command{
			{
				Name:  "init",
				Usage: "scaffold a monad",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println("initialized: ", cmd.Args().First())
					return nil
				},
			},
			{
				Name:  "deploy",
				Usage: "deploy a monad",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println("deployed: ", cmd.Args().First())
					return nil
				},
			},
			{
				Name:  "destroy",
				Usage: "destroy a monad",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println("destroyed: ", cmd.Args().First())
					return nil
				},
			},
			{
				Name:  "list",
				Usage: "list monads",
				Action: func(ctx context.Context, c *cli.Command) error {
					fmt.Println("listing monads...")
					return nil
				},
			},
			{
				Name:  "ecr",
				Usage: "monad artifacts",
				Commands: []*cli.Command{
					{
						Name:  "login",
						Usage: "login to ecr",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							fmt.Println("logged in!")
							return nil
						},
					},
					{
						Name:  "init",
						Usage: "initialize monad repository",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							fmt.Println("initialized: ", cmd.Args().First())
							return nil
						},
					},
					{
						Name:  "destroy",
						Usage: "destroy a monad repository",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							fmt.Println("destroyed: ", cmd.Args().First())
							return nil
						},
					},
					{
						Name:  "tag",
						Usage: "tag a monad image",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							fmt.Println("tagged: ", cmd.Args().First())
							return nil
						},
					},
					{
						Name:  "untag",
						Usage: "untag a monad image",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							fmt.Println("untagged: ", cmd.Args().First())
							return nil
						},
					},
				},
			},
			{
				Name:  "data",
				Usage: "contextual templating",
				Commands: []*cli.Command{
					{
						Name:  "list",
						Usage: "list available key/values",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							fmt.Println("printing data table...")
							return nil
						},
					},
					{
						Name:  "render",
						Usage: "render file to stdout",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							fmt.Println("rendering file: ", cmd.Args().First())
							return nil
						},
					},
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal().Err(err)
	}
}
