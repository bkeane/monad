package cloudwatch

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/smithy-go"
	"github.com/rs/zerolog/log"
)

type CloudWatchResources interface {
	LogGroupName() string
	LogGroupArn() string
	LogGroupTags() map[string]string
	LogGroupRetention() int32
	Client() *cloudwatchlogs.Client
}

type Client struct {
	cloudwatch CloudWatchResources
}

func Init(cloudwatch CloudWatchResources) *Client {
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
		Str("group", s.cloudwatch.LogGroupName()).
		Int32("retention", s.cloudwatch.LogGroupRetention()).
		Msg("cloudwatch")

	_, err := s.cloudwatch.Client().CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(s.cloudwatch.LogGroupName()),
		Tags:         s.cloudwatch.LogGroupTags(),
	})

	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "ResourceAlreadyExistsException":
			_, err := s.cloudwatch.Client().TagResource(ctx, &cloudwatchlogs.TagResourceInput{
				ResourceArn: aws.String(s.cloudwatch.LogGroupArn()),
				Tags:        s.cloudwatch.LogGroupTags(),
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
		LogGroupName:    aws.String(c.cloudwatch.LogGroupName()),
		RetentionInDays: aws.Int32(c.cloudwatch.LogGroupRetention()),
	})

	return err
}

func (c *Client) DeleteLogGroup(ctx context.Context) error {
	var apiErr smithy.APIError

	log.Info().
		Str("action", "delete").
		Str("group", c.cloudwatch.LogGroupName()).
		Msg("cloudwatch")

	_, err := c.cloudwatch.Client().DeleteLogGroup(ctx, &cloudwatchlogs.DeleteLogGroupInput{
		LogGroupName: aws.String(c.cloudwatch.LogGroupName()),
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
