package main

import (
	"context"
	"fmt"
	"os"

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
		Usage:  "management plane",
		Flags:  flag.Flags[basis.Basis](),
		Before: flag.Before[basis.Basis](),
		Commands: []*cli.Command{
			{
				Name:      "init",
				Usage:     "scaffold a monad",
				UsageText: "monad init <LANGUAGE> [LOCATION]",
				Flags:     flag.Flags[scaffold.Scaffold](),
				Before:    flag.Before[scaffold.Scaffold](),
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
				Usage:  "deploy a monad",
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
				Usage: "destroy a monad",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					saga, err := pkg.Saga(ctx)
					if err != nil {
						return err
					}

					return saga.Undo(ctx)
				},
			},
			{
				Name:  "list",
				Usage: "list monads",
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
				Name:  "ecr",
				Usage: "monad artifacts",
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
						Usage: "initialize monad repository",
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
						Usage: "destroy a monad repository",
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
						Usage: "untag a monad image",
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
				Name:  "data",
				Usage: "contextual templating",
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
						Name:  "render",
						Usage: "render file to stdout",
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
				},
			},
			{
				Name:   "logs",
				Usage:  "stream logs from monad",
				Flags:  flag.Flags[monadlog.Log](),
				Before: flag.Before[monadlog.Log](),
				Action: func(ctx context.Context, cmd *cli.Command) error {
					logs, err := pkg.Log(ctx)
					if err != nil {
						return err
					}

					return logs.Fetch(ctx)
				},
			},
		},
	}

	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		log.Fatal().Err(err).Msg("command failed")
	}
}
