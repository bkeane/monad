package step

import (
	"context"

	"github.com/bkeane/monad/pkg/config"
	"github.com/bkeane/monad/pkg/registry"
	"github.com/bkeane/monad/pkg/step/apigateway"
	"github.com/bkeane/monad/pkg/step/cloudwatch"
	"github.com/bkeane/monad/pkg/step/eventbridge"
	"github.com/bkeane/monad/pkg/step/iam"
	"github.com/bkeane/monad/pkg/step/lambda"

	v "github.com/go-ozzo/ozzo-validation/v4"
)

//
// Steps
//

type Steps struct {
	iam         *iam.Step
	cloudwatch  *cloudwatch.Step
	lambda      *lambda.Step
	apigateway  *apigateway.Step
	eventbridge *eventbridge.Step
}

//
// Derive
//

func Derive(ctx context.Context, config *config.Config) (*Steps, error) {
	ecrConfig, err := config.Ecr(ctx)
	if err != nil {
		return nil, err
	}
	registryClient := registry.Derive(ecrConfig)

	iamConfig, err := config.Iam(ctx)
	if err != nil {
		return nil, err
	}

	cloudwatchConfig, err := config.CloudWatch(ctx)
	if err != nil {
		return nil, err
	}

	lambdaConfig, err := config.Lambda(ctx)
	if err != nil {
		return nil, err
	}

	vpcConfig, err := config.Vpc(ctx)
	if err != nil {
		return nil, err
	}

	apigatewayConfig, err := config.ApiGateway(ctx)
	if err != nil {
		return nil, err
	}

	eventbridgeConfig, err := config.EventBridge(ctx)
	if err != nil {
		return nil, err
	}
	
	steps := &Steps{
		iam:         iam.Derive(iamConfig),
		cloudwatch:  cloudwatch.Derive(cloudwatchConfig),
		lambda:      lambda.Derive(lambdaConfig, registryClient, iamConfig, vpcConfig, cloudwatchConfig),
		apigateway:  apigateway.Derive(apigatewayConfig, lambdaConfig),
		eventbridge: eventbridge.Derive(eventbridgeConfig, lambdaConfig),
	}

	if err := steps.Validate(); err != nil {
		return nil, err
	}

	return steps, nil
}

//
// Validate
//

func (s *Steps) Validate() error {
	return v.ValidateStruct(s,
		v.Field(&s.iam),
		v.Field(&s.cloudwatch),
		v.Field(&s.lambda),
		v.Field(&s.apigateway),
		v.Field(&s.eventbridge),
	)
}

//
// Accessors
//

// IAM returns the IAM step instance
func (s *Steps) IAM() *iam.Step {
	return s.iam
}

// CloudWatch returns the CloudWatch step instance
func (s *Steps) CloudWatch() *cloudwatch.Step {
	return s.cloudwatch
}

// Lambda returns the Lambda step instance
func (s *Steps) Lambda() *lambda.Step {
	return s.lambda
}

// ApiGateway returns the API Gateway step instance
func (s *Steps) ApiGateway() *apigateway.Step {
	return s.apigateway
}

// EventBridge returns the EventBridge step instance
func (s *Steps) EventBridge() *eventbridge.Step {
	return s.eventbridge
}

