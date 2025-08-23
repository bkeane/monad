package saga

import (
	"context"

	"github.com/bkeane/monad/pkg/client"
	"github.com/bkeane/monad/pkg/client/apigateway"
	"github.com/bkeane/monad/pkg/client/cloudwatch"
	"github.com/bkeane/monad/pkg/client/eventbridge"
	"github.com/bkeane/monad/pkg/client/iam"
	"github.com/bkeane/monad/pkg/client/lambda"

	"github.com/rs/zerolog/log"
)

type Client interface {
	IAM() *iam.Client
	CloudWatch() *cloudwatch.Client
	Lambda() *lambda.Client
	ApiGateway() *apigateway.Client
	EventBridge() *eventbridge.Client
}

type Step interface {
	Mount(ctx context.Context) error
	Unmount(ctx context.Context) error
}

type Saga struct {
	iam         Step
	eventbridge Step
	apigateway  Step
	cloudwatch  Step
	lambda      Step
}

func Derive(ctx context.Context, client *client.Client) *Saga {
	return &Saga{
		iam:         client.IAM(),
		cloudwatch:  client.CloudWatch(),
		lambda:      client.Lambda(),
		apigateway:  client.ApiGateway(),
		eventbridge: client.EventBridge(),
	}
}

func (a *Saga) Do(ctx context.Context) error {
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

func (a *Saga) Undo(ctx context.Context) error {
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
