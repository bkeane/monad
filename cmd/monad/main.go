package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bkeane/monad/cmd/monad/desc"
	"github.com/bkeane/monad/cmd/monad/pkg"
	"github.com/bkeane/monad/pkg/basis"
	"github.com/bkeane/monad/pkg/config"
	"github.com/bkeane/monad/pkg/flag"
	monadlog "github.com/bkeane/monad/pkg/log"
	"github.com/bkeane/monad/pkg/scaffold"

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
	flag.DisableDefaults()

	cmd := &cli.Command{
		Name:   "monad",
		Usage:  "service management",
		Flags:  flag.Flags[basis.Basis](),
		Before: flag.Before[basis.Basis](),
		Commands: []*cli.Command{
			{
				Name:        "init",
				Usage:       "scaffold a service",
				UsageText:   "monad init <LANGUAGE> [LOCATION]",
				Description: desc.Init(),
				Flags:       flag.Flags[scaffold.Scaffold](),
				Before:      flag.Before[scaffold.Scaffold](),
				Action: func(ctx context.Context, cmd *cli.Command) error {
					language := cmd.Args().Get(0)
					targetDir := cmd.Args().Get(1)

					if language == "" {
						return fmt.Errorf("language is required")
					}

					scaffold, err := pkg.Scaffold(ctx)
					if err != nil {
						return err
					}

					return scaffold.Create(language, targetDir)
				},
			},
			{
				Name:   "deploy",
				Usage:  "deploy a service",
				Flags:  flag.Flags[config.Config](),
				Before: flag.Before[config.Config](),
				Action: func(ctx context.Context, cmd *cli.Command) error {
					saga, err := pkg.Saga(ctx)
					if err != nil {
						return err
					}

					return saga.Do(ctx)
				},
			},
			{
				Name:  "destroy",
				Usage: "destroy a service",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					saga, err := pkg.Saga(ctx)
					if err != nil {
						return err
					}

					return saga.Undo(ctx)
				},
			},
			{
				Name:        "list",
				Usage:       "list services",
				Description: desc.List(),
				Action: func(ctx context.Context, c *cli.Command) error {
					state, err := pkg.State(ctx)
					if err != nil {
						return err
					}

					table, err := state.Table(ctx)
					if err != nil {
						return err
					}

					fmt.Println(table)

					return nil
				},
			},
			{
				Name:   "ecr",
				Usage:  "service artifacts",
				Flags:  flag.Flags[basis.Basis](),
				Before: flag.Before[basis.Basis](),
				Commands: []*cli.Command{
					{
						Name:  "login",
						Usage: "login to ecr",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							registry, err := pkg.Registry(ctx)
							if err != nil {
								return err
							}

							return registry.Login(ctx)
						},
					},
					{
						Name:  "init",
						Usage: "initialize service repository",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							registry, err := pkg.Registry(ctx)
							if err != nil {
								return err
							}

							return registry.CreateRepository(ctx)
						},
					},
					{
						Name:  "destroy",
						Usage: "destroy a service repository",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							registry, err := pkg.Registry(ctx)
							if err != nil {
								return err
							}

							return registry.DeleteRepository(ctx)
						},
					},
					{
						Name:  "untag",
						Usage: "untag a service image",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							registry, err := pkg.Registry(ctx)
							if err != nil {
								return err
							}

							return registry.Untag(ctx)
						},
					},
				},
			},
			{
				Name:   "render",
				Usage:  "contextual templating",
				Flags:  flag.Flags[basis.Basis](),
				Before: flag.Before[basis.Basis](),
				Commands: []*cli.Command{
					{
						Name:  "list",
						Usage: "list available key/values",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							basis, err := pkg.Basis(ctx)
							if err != nil {
								return err
							}

							table, err := basis.Table()
							if err != nil {
								return err
							}

							fmt.Print(table)
							return nil
						},
					},
					{
						Name:      "file",
						Usage:     "render file to stdout",
						UsageText: "monad render file <PATH>",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							file := cmd.Args().First()
							if file == "" {
								return fmt.Errorf("file required")
							}

							basis, err := pkg.Basis(ctx)
							if err != nil {
								return err
							}

							content, err := os.ReadFile(file)
							if err != nil {
								return err
							}

							result, err := basis.Render(string(content))
							if err != nil {
								return err
							}

							fmt.Print(result)
							return nil
						},
					},
					{
						Name:      "string",
						Usage:     "render string to stdout",
						UsageText: "monad render string <STRING>",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							input := cmd.Args().First()
							if input == "" {
								return fmt.Errorf("input required")
							}

							basis, err := pkg.Basis(ctx)
							if err != nil {
								return err
							}

							result, err := basis.Render(input)
							if err != nil {
								return err
							}

							fmt.Print(result)
							return nil
						},
					},
				},
			},
			{
				Name:   "logs",
				Usage:  "fetch service logs",
				Flags:  flag.Flags[monadlog.LogGroup](),
				Before: flag.Before[monadlog.LogGroup](),
				Action: func(ctx context.Context, cmd *cli.Command) error {
					logs, err := pkg.Log(ctx)
					if err != nil {
						return err
					}

					if logs.LogGroupTail {
						return logs.Tail(ctx)
					}

					return logs.Dump(ctx)

				},
			},
		},
	}

	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		log.Fatal().Err(err).Msg("command failed")
	}
}
