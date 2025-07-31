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
		Str("action", "put").
		Str("group", s.config.CloudWatch().LogGroup()).
		Int32("retention", s.config.CloudWatch().LogRetention()).
		Msg("cloudwatch")

	if err := s.PutLogGroup(ctx); err != nil {
		return err
	}

	if err := s.PutLogGroupRetentionPolicy(ctx, s.config.CloudWatch().LogRetention()); err != nil {
		return err
	}

	return nil
}

func (s *Cloudwatch) Undo(ctx context.Context) error {
	log.Info().
		Str("action", "delete").
		Str("group", s.config.CloudWatch().LogGroup()).
		Msg("cloudwatch")

	if err := s.DeleteLogGroup(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Cloudwatch) PutLogGroup(ctx context.Context) error {
	var apiErr smithy.APIError
	_, err := s.config.CloudWatch().Client().CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(s.config.CloudWatch().LogGroup()),
		Tags:         s.config.Schema().Tags(),
	})

	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "ResourceAlreadyExistsException":
			_, err := s.config.CloudWatch().Client().TagResource(ctx, &cloudwatchlogs.TagResourceInput{
				ResourceArn: aws.String(s.config.CloudWatch().LogGroupArn()),
				Tags:        s.config.Schema().Tags(),
			})
			return err
		default:
			return err
		}
	}

	return err
}

func (c *Cloudwatch) PutLogGroupRetentionPolicy(ctx context.Context, days int32) error {
	_, err := c.config.CloudWatch().Client().PutRetentionPolicy(ctx, &cloudwatchlogs.PutRetentionPolicyInput{
		LogGroupName:    aws.String(c.config.CloudWatch().LogGroup()),
		RetentionInDays: aws.Int32(days),
	})

	return err
}

func (c *Cloudwatch) DeleteLogGroup(ctx context.Context) error {
	var apiErr smithy.APIError
	_, err := c.config.CloudWatch().Client().DeleteLogGroup(ctx, &cloudwatchlogs.DeleteLogGroupInput{
		LogGroupName: aws.String(c.config.CloudWatch().LogGroup()),
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
