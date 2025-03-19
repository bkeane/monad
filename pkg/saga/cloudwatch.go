package saga

import (
	"context"
	"errors"

	"github.com/bkeane/monad/pkg/param"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/smithy-go"
	"github.com/rs/zerolog/log"
)

type Cloudwatch struct {
	config param.Aws
}

func (s Cloudwatch) Init(ctx context.Context, c param.Aws) *Cloudwatch {
	return &Cloudwatch{
		config: c,
	}
}

func (s *Cloudwatch) Do(ctx context.Context) error {
	log.Info().
		Str("arn", s.config.CloudwatchLogGroup()).
		Int32("retention", s.config.CloudwatchLogRetention()).
		Msg("ensuring log group exists")

	if err := s.PutLogGroup(ctx); err != nil {
		return err
	}

	if err := s.PutLogGroupRetentionPolicy(ctx, s.config.CloudwatchLogRetention()); err != nil {
		return err
	}

	return nil
}

func (s *Cloudwatch) Undo(ctx context.Context) error {
	log.Info().
		Str("arn", s.config.CloudwatchLogGroup()).
		Msg("deleting log group")

	if err := s.DeleteLogGroup(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Cloudwatch) PutLogGroup(ctx context.Context) error {
	var apiErr smithy.APIError
	_, err := s.config.CloudWatch.Client.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(s.config.CloudwatchLogGroup()),
		Tags:         s.config.Tags(),
	})

	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "ResourceAlreadyExistsException":
			_, err := s.config.CloudWatch.Client.TagResource(ctx, &cloudwatchlogs.TagResourceInput{
				ResourceArn: aws.String(s.config.CloudwatchLogGroup()),
				Tags:        s.config.Tags(),
			})
			return err
		default:
			return err
		}
	}

	return err
}

func (c *Cloudwatch) PutLogGroupRetentionPolicy(ctx context.Context, days int32) error {
	_, err := c.config.CloudWatch.Client.PutRetentionPolicy(ctx, &cloudwatchlogs.PutRetentionPolicyInput{
		LogGroupName:    aws.String(c.config.CloudwatchLogGroup()),
		RetentionInDays: aws.Int32(days),
	})

	return err
}

func (c *Cloudwatch) DeleteLogGroup(ctx context.Context) error {
	var apiErr smithy.APIError
	_, err := c.config.CloudWatch.Client.DeleteLogGroup(ctx, &cloudwatchlogs.DeleteLogGroupInput{
		LogGroupName: aws.String(c.config.CloudwatchLogGroup()),
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
