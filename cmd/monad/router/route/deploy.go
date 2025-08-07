package route

import (
	"context"
	"fmt"

	"github.com/bkeane/monad/pkg/model/apigateway"
	"github.com/bkeane/monad/pkg/model/cloudwatch"
	"github.com/bkeane/monad/pkg/model/ecr"
	"github.com/bkeane/monad/pkg/model/eventbridge"
	"github.com/bkeane/monad/pkg/model/iam"
	"github.com/bkeane/monad/pkg/model/lambda"
	"github.com/bkeane/monad/pkg/model/vpc"
)

type Deploy struct {
	apigateway.ApiGateway
	cloudwatch.CloudWatch
	ecr.Ecr
	eventbridge.EventBridge
	iam.IAM
	lambda.Lambda
	vpc.VPC
}

func (d *Deploy) Route(ctx context.Context, r Root) error {
	// Process the embedded model with CLI args
	fmt.Println("not implemented yet")
	// if err := d.Model.Process(ctx, r.AwsConfig); err != nil {
	// 	return err
	// }

	// saga, err := saga.Init(ctx, r.AwsConfig)
	// if err != nil {
	// 	return err
	// }

	// return saga.Do(ctx)
	return nil
}
