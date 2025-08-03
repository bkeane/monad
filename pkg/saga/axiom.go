package saga

import (
	"context"

	"github.com/bkeane/monad/internal/registry"
	"github.com/rs/zerolog/log"

	// Resources
	"github.com/bkeane/monad/pkg/param"

	// Clients
	gwc "github.com/bkeane/monad/pkg/client/apigateway"
	cwc "github.com/bkeane/monad/pkg/client/cloudwatch"
	ebc "github.com/bkeane/monad/pkg/client/eventbridge"
	iamc "github.com/bkeane/monad/pkg/client/iam"
	lmbc "github.com/bkeane/monad/pkg/client/lambda"

	// Steps
	gws "github.com/bkeane/monad/pkg/saga/steps/apigateway"
	cws "github.com/bkeane/monad/pkg/saga/steps/cloudwatch"
	ebs "github.com/bkeane/monad/pkg/saga/steps/eventbridge"
	iams "github.com/bkeane/monad/pkg/saga/steps/iam"
	lmbs "github.com/bkeane/monad/pkg/saga/steps/lambda"
)

type Axiom struct {
	iam         *iams.Step
	eventbridge *ebs.Step
	apigateway  *gws.Step
	cloudwatch  *cws.Step
	lambda      *lmbs.Step
}

func Init(ctx context.Context, c *param.Aws) *Axiom {
	iamc := iamc.Init(c.IAM(), c.Schema())
	lmbc := lmbc.Init(c.Lambda(), c.IAM(), c.Vpc(), c.CloudWatch(), c.Schema())
	gwc := gwc.Init(c.ApiGateway(), c.Lambda())
	cwc := cwc.Init(c.CloudWatch(), c.Schema())
	ebc := ebc.Init(c.EventBridge(), c.Lambda(), c.Schema())

	return &Axiom{
		iam:         iams.Init(iamc),
		lambda:      lmbs.Init(lmbc),
		apigateway:  gws.Init(gwc),
		eventbridge: ebs.Init(ebc),
		cloudwatch:  cws.Init(cwc),
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
