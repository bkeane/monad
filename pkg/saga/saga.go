package saga

import (
	"context"

	"github.com/bkeane/monad/pkg/basis"
	"github.com/bkeane/monad/pkg/client"
	"github.com/bkeane/monad/pkg/config"

	"github.com/rs/zerolog/log"
)

type Step interface {
	Mount(ctx context.Context) error
	Unmount(ctx context.Context) error
}

type Axiom struct {
	iam         Step
	eventbridge Step
	apigateway  Step
	cloudwatch  Step
	lambda      Step
}

func Init(ctx context.Context, basis *basis.Basis) (*Axiom, error) {
	// Derive Configuration
	config, err := config.Derive(ctx, basis)
	if err != nil {
		return nil, err
	}

	// Initialize unified client
	client, err := client.Init(config)
	if err != nil {
		return nil, err
	}

	return &Axiom{
		iam:         client.IAM(),
		cloudwatch:  client.CloudWatch(),
		lambda:      client.Lambda(),
		apigateway:  client.ApiGateway(),
		eventbridge: client.EventBridge(),
	}, nil
}

func (a *Axiom) Do(ctx context.Context) error {
	if err := a.iam.Mount(ctx); err != nil {
		log.Error().Err(err).Msg("iam mount failed")
		return err
	}

	if err := a.cloudwatch.Mount(ctx); err != nil {
		log.Error().Err(err).Msg("cloudwatch mount failed")
		return err
	}

	if err := a.lambda.Mount(ctx); err != nil {
		log.Error().Err(err).Msg("lambda mount failed")
		return err
	}

	if err := a.apigateway.Mount(ctx); err != nil {
		log.Error().Err(err).Msg("apigateway mount failed")
		return err
	}

	if err := a.eventbridge.Mount(ctx); err != nil {
		log.Error().Err(err).Msg("eventbridge mount failed")
		return err
	}

	return nil
}

func (a *Axiom) Undo(ctx context.Context) error {
	if err := a.eventbridge.Unmount(ctx); err != nil {
		log.Error().Err(err).Msg("eventbridge unmount failed")
		return err
	}

	if err := a.apigateway.Unmount(ctx); err != nil {
		log.Error().Err(err).Msg("apigateway unmount failed")
		return err
	}

	if err := a.cloudwatch.Unmount(ctx); err != nil {
		log.Error().Err(err).Msg("cloudwatch unmount failed")
		return err
	}

	if err := a.lambda.Unmount(ctx); err != nil {
		log.Error().Err(err).Msg("lambda unmount failed")
		return err
	}

	if err := a.iam.Unmount(ctx); err != nil {
		log.Error().Err(err).Msg("iam unmount failed")
		return err
	}

	return nil
}
