package client

import (
	"github.com/bkeane/monad/pkg/client/apigateway"
	"github.com/bkeane/monad/pkg/client/cloudwatch"
	"github.com/bkeane/monad/pkg/client/ecr"
	"github.com/bkeane/monad/pkg/client/eventbridge"
	"github.com/bkeane/monad/pkg/client/iam"
	"github.com/bkeane/monad/pkg/client/lambda"
	"github.com/bkeane/monad/pkg/config"

	v "github.com/go-ozzo/ozzo-validation/v4"
)

//
// Client
//

type Client struct {
	iam         *iam.Client
	cloudwatch  *cloudwatch.Client
	lambda      *lambda.Client
	apigateway  *apigateway.Client
	eventbridge *eventbridge.Client
	ecr         *ecr.Client
}

//
// Derive
//

func Derive(config *config.Config) (*Client, error) {
	client := &Client{
		iam:         iam.Derive(config.Iam()),
		cloudwatch:  cloudwatch.Derive(config.CloudWatch()),
		lambda:      lambda.Derive(config.Lambda(), config.Ecr(), config.Iam(), config.Vpc(), config.CloudWatch()),
		apigateway:  apigateway.Derive(config.ApiGateway(), config.Lambda()),
		eventbridge: eventbridge.Derive(config.EventBridge(), config.Lambda()),
		ecr:         ecr.Derive(config.Ecr()),
	}

	if err := client.Validate(); err != nil {
		return nil, err
	}

	return client, nil
}

//
// Validate
//

func (c *Client) Validate() error {
	return v.ValidateStruct(c,
		v.Field(&c.iam),
		v.Field(&c.cloudwatch),
		v.Field(&c.lambda),
		v.Field(&c.apigateway),
		v.Field(&c.eventbridge),
		v.Field(&c.ecr),
	)
}

//
// Accessors
//

// IAM returns the IAM client instance
func (c *Client) IAM() *iam.Client {
	return c.iam
}

// CloudWatch returns the CloudWatch client instance
func (c *Client) CloudWatch() *cloudwatch.Client {
	return c.cloudwatch
}

// Lambda returns the Lambda client instance
func (c *Client) Lambda() *lambda.Client {
	return c.lambda
}

// ApiGateway returns the API Gateway client instance
func (c *Client) ApiGateway() *apigateway.Client {
	return c.apigateway
}

// EventBridge returns the EventBridge client instance
func (c *Client) EventBridge() *eventbridge.Client {
	return c.eventbridge
}

// Ecr returns the ECR client instance
func (c *Client) Ecr() *ecr.Client {
	return c.ecr
}
