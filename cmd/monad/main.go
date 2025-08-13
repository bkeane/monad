package main

import (
	"context"
	"os"

	"github.com/bkeane/monad/pkg/basis"

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

	_, err := basis.Derive(ctx, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("'sploded")
	}
}
