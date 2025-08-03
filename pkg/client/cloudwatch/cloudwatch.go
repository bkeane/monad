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
	LogGroup() string
	LogGroupArn() string
	LogRetention() int32
	Client() *cloudwatchlogs.Client
}

type SchemaResources interface {
	Tags() map[string]string
}

type Client struct {
	cloudwatch CloudWatchResources
	schema     SchemaResources
}

func Init(cloudwatch CloudWatchResources, schema SchemaResources) *Client {
	return &Client{
		cloudwatch: cloudwatch,
		schema:     schema,
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
		Str("group", s.cloudwatch.LogGroup()).
		Int32("retention", s.cloudwatch.LogRetention()).
		Msg("cloudwatch")

	_, err := s.cloudwatch.Client().CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(s.cloudwatch.LogGroup()),
		Tags:         s.schema.Tags(),
	})

	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "ResourceAlreadyExistsException":
			_, err := s.cloudwatch.Client().TagResource(ctx, &cloudwatchlogs.TagResourceInput{
				ResourceArn: aws.String(s.cloudwatch.LogGroupArn()),
				Tags:        s.schema.Tags(),
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
		LogGroupName:    aws.String(c.cloudwatch.LogGroup()),
		RetentionInDays: aws.Int32(c.cloudwatch.LogRetention()),
	})

	return err
}

func (c *Client) DeleteLogGroup(ctx context.Context) error {
	var apiErr smithy.APIError

	log.Info().
		Str("action", "delete").
		Str("group", c.cloudwatch.LogGroup()).
		Msg("cloudwatch")

	_, err := c.cloudwatch.Client().DeleteLogGroup(ctx, &cloudwatchlogs.DeleteLogGroupInput{
		LogGroupName: aws.String(c.cloudwatch.LogGroup()),
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
