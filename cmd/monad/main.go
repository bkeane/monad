package main

import (
	"context"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/bkeane/monad/cmd/monad/router"
	"github.com/bkeane/monad/cmd/monad/router/route"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
	ctx := context.Background()
	var root route.Root

	parser, err := arg.NewParser(
		arg.Config{
			IgnoreDefault:     true,
			StrictSubcommands: true,
		},
		&root,
	)

	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	parser.MustParse(os.Args[1:])

	if err := router.Route(ctx, root); err != nil {
		log.Fatal().Msg(err.Error())
	}
}
