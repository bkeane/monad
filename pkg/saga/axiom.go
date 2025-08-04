package saga

import (
	"context"

	"github.com/rs/zerolog/log"

	// Resources
	"github.com/bkeane/monad/pkg/param"

	// Clients
	gw "github.com/bkeane/monad/pkg/client/apigateway"
	cw "github.com/bkeane/monad/pkg/client/cloudwatch"
	eb "github.com/bkeane/monad/pkg/client/eventbridge"
	iam "github.com/bkeane/monad/pkg/client/iam"
	lmb "github.com/bkeane/monad/pkg/client/lambda"
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

func Init(ctx context.Context, c *param.Aws) *Axiom {
	return &Axiom{
		lambda:      lmb.Init(c.Lambda(), c.Registry(), c.IAM(), c.Vpc(), c.CloudWatch()),
		apigateway:  gw.Init(c.ApiGateway(), c.Lambda()),
		eventbridge: eb.Init(c.EventBridge(), c.Lambda()),
		cloudwatch:  cw.Init(c.CloudWatch()),
		iam:         iam.Init(c.IAM()),
	}
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
