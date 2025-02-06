package saga

import (
	"context"
	"errors"

	"github.com/bkeane/monad/pkg/config/release"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/smithy-go"

	"github.com/rs/zerolog"
)

type Cloudwatch struct {
	Client  *cloudwatchlogs.Client
	release release.Config
	log     *zerolog.Logger
}

func (c Cloudwatch) Init(ctx context.Context, r release.Config) *Cloudwatch {
	return &Cloudwatch{
		Client:  cloudwatchlogs.NewFromConfig(r.AwsConfig),
		release: r,
		log:     zerolog.Ctx(ctx),
	}
}

func (c *Cloudwatch) Do(ctx context.Context) error {
	c.log.Info().Msg("ensuring log group exists")
	if err := c.PutLogGroup(ctx); err != nil {
		return err
	}

	c.log.Info().Msg("ensuring log group retention policy")
	if err := c.PutLogGroupRetentionPolicy(ctx, 7); err != nil {
		return err
	}

	return nil
}

func (c *Cloudwatch) Undo(ctx context.Context) error {
	c.log.Info().Msg("deleting log group")
	if err := c.DeleteLogGroup(ctx); err != nil {
		return err
	}

	return nil
}

func (c *Cloudwatch) PutLogGroup(ctx context.Context) error {
	var apiErr smithy.APIError
	_, err := c.Client.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String("/aws/lambda/" + c.release.ResourceName()),
		Tags:         c.release.ResourceTags(),
	})

	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "ResourceAlreadyExistsException":
			return nil
		default:
			return err
		}
	}

	return err
}

func (c *Cloudwatch) PutLogGroupRetentionPolicy(ctx context.Context, days int32) error {
	_, err := c.Client.PutRetentionPolicy(ctx, &cloudwatchlogs.PutRetentionPolicyInput{
		LogGroupName:    aws.String("/aws/lambda/" + c.release.ResourceName()),
		RetentionInDays: aws.Int32(days),
	})

	return err
}

func (c *Cloudwatch) DeleteLogGroup(ctx context.Context) error {
	var apiErr smithy.APIError
	_, err := c.Client.DeleteLogGroup(ctx, &cloudwatchlogs.DeleteLogGroupInput{
		LogGroupName: aws.String("/aws/lambda/" + c.release.ResourceName()),
	})

	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "ResourceNotFoundException":
			return nil
		default:
			return err
		}
	}

	return err
}
