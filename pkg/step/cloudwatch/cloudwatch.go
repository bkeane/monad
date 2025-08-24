package cloudwatch

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/smithy-go"
	"github.com/rs/zerolog/log"
)

type CloudWatchConfig interface {
	Client() *cloudwatchlogs.Client
	Name() string
	Arn() string
	Retention() int32
	Tags() map[string]string
}

type LogGroup struct {
	Name      string
	Retention int32
}

type Summary struct {
	LogGroupsCreated []LogGroup
	LogGroupsDeleted []LogGroup
}

//
// Client
//

type Step struct {
	cloudwatch CloudWatchConfig
}

//
// Derive
//

func Derive(cloudwatch CloudWatchConfig) *Step {
	return &Step{
		cloudwatch: cloudwatch,
	}
}

func (s *Step) Mount(ctx context.Context) error {
	summary, err := s.mount(ctx)
	if err != nil {
		return err
	}

	// Log only action=put for consistency with other services
	for _, logGroup := range summary.LogGroupsCreated {
		log.Info().
			Str("action", "put").
			Str("group", logGroup.Name).
			Int32("retention", logGroup.Retention).
			Msg("cloudwatch")
	}

	return nil
}

func (s *Step) Unmount(ctx context.Context) error {
	summary, err := s.unmount(ctx)
	if err != nil {
		return err
	}

	// Log action=delete
	for _, logGroup := range summary.LogGroupsDeleted {
		log.Info().
			Str("action", "delete").
			Str("group", logGroup.Name).
			Msg("cloudwatch")
	}

	return nil
}

// Internal methods that return summaries of work done
func (s *Step) mount(ctx context.Context) (Summary, error) {
	var summary Summary

	if err := s.PutLogGroup(ctx); err != nil {
		return summary, err
	}

	if err := s.PutRetentionPolicy(ctx); err != nil {
		return summary, err
	}

	// Record what was created
	logGroup := LogGroup{
		Name:      s.cloudwatch.Name(),
		Retention: s.cloudwatch.Retention(),
	}
	summary.LogGroupsCreated = append(summary.LogGroupsCreated, logGroup)

	return summary, nil
}

func (s *Step) unmount(ctx context.Context) (Summary, error) {
	var summary Summary

	if err := s.DeleteLogGroup(ctx); err != nil {
		return summary, err
	}

	// Record what was deleted
	logGroup := LogGroup{
		Name:      s.cloudwatch.Name(),
		Retention: s.cloudwatch.Retention(),
	}
	summary.LogGroupsDeleted = append(summary.LogGroupsDeleted, logGroup)

	return summary, nil
}

func (s *Step) PutLogGroup(ctx context.Context) error {
	var apiErr smithy.APIError

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

func (c *Step) PutRetentionPolicy(ctx context.Context) error {
	_, err := c.cloudwatch.Client().PutRetentionPolicy(ctx, &cloudwatchlogs.PutRetentionPolicyInput{
		LogGroupName:    aws.String(c.cloudwatch.Name()),
		RetentionInDays: aws.Int32(c.cloudwatch.Retention()),
	})

	return err
}

func (c *Step) DeleteLogGroup(ctx context.Context) error {
	var apiErr smithy.APIError

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
