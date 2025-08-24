package step

import (
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

func Derive(config *config.Config) (*Steps, error) {
	registryClient := registry.Derive(config.Ecr())
	
	steps := &Steps{
		iam:         iam.Derive(config.Iam()),
		cloudwatch:  cloudwatch.Derive(config.CloudWatch()),
		lambda:      lambda.Derive(config.Lambda(), registryClient, config.Iam(), config.Vpc(), config.CloudWatch()),
		apigateway:  apigateway.Derive(config.ApiGateway(), config.Lambda()),
		eventbridge: eventbridge.Derive(config.EventBridge(), config.Lambda()),
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

