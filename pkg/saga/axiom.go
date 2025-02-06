package saga

import (
	"context"

	"github.com/bkeane/monad/pkg/config/release"
	"github.com/rs/zerolog"
)

type Step interface {
	Do(ctx context.Context) error
	Undo(ctx context.Context) error
}

type Axiom struct {
	iam         Step
	lambda      Step
	apigateway  Step
	eventbridge Step
	cloudwatch  Step
	log         *zerolog.Logger
}

func Init(ctx context.Context, r *release.Config) *Axiom {
	return &Axiom{
		iam:         IAM{}.Init(ctx, *r),
		lambda:      Lambda{}.Init(ctx, *r),
		apigateway:  ApiGatewayV2{}.Init(ctx, *r),
		eventbridge: EventBridge{}.Init(ctx, *r),
		cloudwatch:  Cloudwatch{}.Init(ctx, *r),
		log:         zerolog.Ctx(ctx),
	}
}

func (a *Axiom) Do(ctx context.Context) error {
	if err := a.iam.Do(ctx); err != nil {
		a.log.Error().Err(err).Msg("iam step failed")
		return err
	}

	if err := a.cloudwatch.Do(ctx); err != nil {
		a.log.Error().Err(err).Msg("cloudwatch step failed")
		return err
	}

	if err := a.lambda.Do(ctx); err != nil {
		a.log.Error().Err(err).Msg("lambda step failed")
		return err
	}

	if err := a.apigateway.Do(ctx); err != nil {
		a.log.Error().Err(err).Msg("apigateway step failed")
		return err
	}

	if err := a.eventbridge.Do(ctx); err != nil {
		a.log.Error().Err(err).Msg("eventbridge step failed")
		return err
	}

	return nil
}

func (a *Axiom) Undo(ctx context.Context) error {
	if err := a.eventbridge.Undo(ctx); err != nil {
		a.log.Error().Err(err).Msg("eventbridge step failed")
		return err
	}

	if err := a.apigateway.Undo(ctx); err != nil {
		a.log.Error().Err(err).Msg("apigateway step failed")
		return err
	}

	if err := a.cloudwatch.Undo(ctx); err != nil {
		a.log.Error().Err(err).Msg("cloudwatch step failed")
		return err
	}

	if err := a.lambda.Undo(ctx); err != nil {
		a.log.Error().Err(err).Msg("lambda step failed")
		return err
	}

	if err := a.iam.Undo(ctx); err != nil {
		a.log.Error().Err(err).Msg("iam step failed")
		return err
	}

	return nil
}
