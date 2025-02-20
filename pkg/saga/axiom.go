package saga

import (
	"context"

	"github.com/bkeane/monad/pkg/param"
	"github.com/bkeane/monad/pkg/registry"

	"github.com/rs/zerolog/log"
)

type Axiom struct {
	iam         *IAM
	lambda      *Lambda
	apigateway  *ApiGatewayV2
	eventbridge *EventBridge
	cloudwatch  *Cloudwatch
}

func Init(ctx context.Context, c param.Aws) *Axiom {
	return &Axiom{
		iam:         IAM{}.Init(ctx, c),
		lambda:      Lambda{}.Init(ctx, c),
		apigateway:  ApiGatewayV2{}.Init(ctx, c),
		eventbridge: EventBridge{}.Init(ctx, c),
		cloudwatch:  Cloudwatch{}.Init(ctx, c),
	}
}

func (a *Axiom) Do(ctx context.Context, image registry.ImagePointer) error {
	if err := a.iam.Do(ctx); err != nil {
		log.Error().Err(err).Msg("iam step failed")
		return err
	}

	if err := a.cloudwatch.Do(ctx); err != nil {
		log.Error().Err(err).Msg("cloudwatch step failed")
		return err
	}

	if err := a.lambda.Do(ctx, image); err != nil {
		log.Error().Err(err).Msg("lambda step failed")
		return err
	}

	if err := a.apigateway.Do(ctx); err != nil {
		log.Error().Err(err).Msg("apigateway step failed")
		return err
	}

	if err := a.eventbridge.Do(ctx); err != nil {
		log.Error().Err(err).Msg("eventbridge step failed")
		return err
	}

	return nil
}

func (a *Axiom) Undo(ctx context.Context) error {
	if err := a.eventbridge.Undo(ctx); err != nil {
		log.Error().Err(err).Msg("eventbridge step failed")
		return err
	}

	if err := a.apigateway.Undo(ctx); err != nil {
		log.Error().Err(err).Msg("apigateway step failed")
		return err
	}

	if err := a.cloudwatch.Undo(ctx); err != nil {
		log.Error().Err(err).Msg("cloudwatch step failed")
		return err
	}

	if err := a.lambda.Undo(ctx); err != nil {
		log.Error().Err(err).Msg("lambda step failed")
		return err
	}

	if err := a.iam.Undo(ctx); err != nil {
		log.Error().Err(err).Msg("iam step failed")
		return err
	}

	return nil
}
