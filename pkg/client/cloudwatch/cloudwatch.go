package cloudwatch

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/smithy-go"
	"github.com/rs/zerolog/log"
)

type CloudWatchConvention interface {
	Client() *cloudwatchlogs.Client
	Name() string
	Arn() string
	Retention() int32
	Tags() map[string]string
}

//
// Client
//

type Client struct {
	cloudwatch CloudWatchConvention
}

//
// Init
//

func Init(cloudwatch CloudWatchConvention) *Client {
	return &Client{
		cloudwatch: cloudwatch,
	}
}

func (s *Client) Mount(ctx context.Context) error {
	if err := s.PutLogGroup(ctx); err != nil {
		return err
	}

	if err := s.PutRetentionPolicy(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Client) Unmount(ctx context.Context) error {
	return s.DeleteLogGroup(ctx)
}

func (s *Client) PutLogGroup(ctx context.Context) error {
	var apiErr smithy.APIError

	log.Info().
		Str("action", "put").
		Str("group", s.cloudwatch.Name()).
		Int32("retention", s.cloudwatch.Retention()).
		Msg("cloudwatch")

	_, err := s.cloudwatch.Client().CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(s.cloudwatch.Name()),
		Tags:         s.cloudwatch.Tags(),
	})

	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "ResourceAlreadyExistsException":
			_, err := s.cloudwatch.Client().TagResource(ctx, &cloudwatchlogs.TagResourceInput{
				ResourceArn: aws.String(s.cloudwatch.Arn()),
				Tags:        s.cloudwatch.Tags(),
			})
			return err
		default:
			return err
		}
	}

	return err
}

func (c *Client) PutRetentionPolicy(ctx context.Context) error {
	_, err := c.cloudwatch.Client().PutRetentionPolicy(ctx, &cloudwatchlogs.PutRetentionPolicyInput{
		LogGroupName:    aws.String(c.cloudwatch.Name()),
		RetentionInDays: aws.Int32(c.cloudwatch.Retention()),
	})

	return err
}

func (c *Client) DeleteLogGroup(ctx context.Context) error {
	var apiErr smithy.APIError

	log.Info().
		Str("action", "delete").
		Str("group", c.cloudwatch.Name()).
		Msg("cloudwatch")

	_, err := c.cloudwatch.Client().DeleteLogGroup(ctx, &cloudwatchlogs.DeleteLogGroupInput{
		LogGroupName: aws.String(c.cloudwatch.Name()),
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
