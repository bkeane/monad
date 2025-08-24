package lambda

import (
	"context"
	"errors"
	"fmt"

	"github.com/bkeane/monad/internal/registryv2"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/smithy-go"
	"github.com/rs/zerolog/log"
)

type LambdaConfig interface {
	Client() *lambda.Client
	FunctionName() string
	Timeout() int32
	MemorySize() int32
	EphemeralStorage() int32
	FunctionArn() string
	Retries() int32
	Env() map[string]string
	Tags() map[string]string
}

type IamConfig interface {
	EniRoleArn() string
	RoleArn() string
}

type VpcConfig interface {
	SecurityGroupIds() []string
	SubnetIds() []string
}

type CloudWatchConfig interface {
	Name() string
}

type EcrConfig interface {
	Clients() (*ecr.Client, *registryv2.Client)
	ImagePath() string
	ImageTag() string
}

type Function struct {
	Name    string
	Memory  int32
	Disk    int32
	Timeout int32
}

type Summary struct {
	FunctionsCreated []Function
	FunctionsDeleted []Function
}

type Client struct {
	lambda     LambdaConfig
	ecr        EcrConfig
	iam        IamConfig
	vpc        VpcConfig
	cloudwatch CloudWatchConfig
}

func Derive(lambda LambdaConfig, ecr EcrConfig, iam IamConfig, vpc VpcConfig, cloudwatch CloudWatchConfig) *Client {
	return &Client{
		lambda:     lambda,
		ecr:        ecr,
		iam:        iam,
		vpc:        vpc,
		cloudwatch: cloudwatch,
	}
}

func (c *Client) Mount(ctx context.Context) error {
	summary, err := c.mount(ctx)
	if err != nil {
		return err
	}

	// Log only action=put for consistency with other services
	for _, function := range summary.FunctionsCreated {
		log.Info().
			Str("action", "put").
			Str("name", function.Name).
			Int32("memory", function.Memory).
			Int32("disk", function.Disk).
			Int32("timeout", function.Timeout).
			Msg("lambda")
	}

	return nil
}

func (c *Client) Unmount(ctx context.Context) error {
	summary, err := c.unmount(ctx)
	if err != nil {
		return err
	}

	// Log action=delete
	for _, function := range summary.FunctionsDeleted {
		log.Info().
			Str("action", "delete").
			Str("name", function.Name).
			Msg("lambda")
	}

	return nil
}

// Internal methods that return summaries of work done
func (c *Client) mount(ctx context.Context) (Summary, error) {
	var summary Summary

	if _, err := c.PutFunction(ctx); err != nil {
		return summary, err
	}

	// Record what was created
	function := Function{
		Name:    c.lambda.FunctionName(),
		Memory:  c.lambda.MemorySize(),
		Disk:    c.lambda.EphemeralStorage(),
		Timeout: c.lambda.Timeout(),
	}
	summary.FunctionsCreated = append(summary.FunctionsCreated, function)

	return summary, nil
}

func (c *Client) unmount(ctx context.Context) (Summary, error) {
	var summary Summary

	if _, err := c.DeleteFunction(ctx); err != nil {
		return summary, err
	}

	// Record what was deleted
	function := Function{
		Name:    c.lambda.FunctionName(),
		Memory:  c.lambda.MemorySize(),
		Disk:    c.lambda.EphemeralStorage(),
		Timeout: c.lambda.Timeout(),
	}
	summary.FunctionsDeleted = append(summary.FunctionsDeleted, function)

	return summary, nil
}

// GET Operations
func (c *Client) GetImage(ctx context.Context) (registryv2.ImagePointer, error) {
	_, registry := c.ecr.Clients()
	return registry.GetImage(ctx, c.ecr.ImagePath(), c.ecr.ImageTag())
}

// PUT Operations
func (c *Client) PutFunction(ctx context.Context) (*lambda.GetFunctionOutput, error) {
	var apiErr smithy.APIError
	var architecture []types.Architecture

	image, err := c.GetImage(ctx)
	if err != nil {
		return nil, err
	}

	switch image.Architecture {
	case "amd64":
		architecture = []types.Architecture{types.ArchitectureX8664}
	case "arm64":
		architecture = []types.Architecture{types.ArchitectureArm64}
	default:
		return nil, fmt.Errorf("unsupported architecture: %s", image.Architecture)
	}

	read := &lambda.GetFunctionInput{
		FunctionName: aws.String(c.lambda.FunctionName()),
	}

	create := &lambda.CreateFunctionInput{
		FunctionName:  aws.String(c.lambda.FunctionName()),
		Role:          aws.String(c.iam.EniRoleArn()),
		Tags:          c.lambda.Tags(),
		Architectures: architecture,
		PackageType:   types.PackageTypeImage,
		Timeout:       aws.Int32(c.lambda.Timeout()),
		MemorySize:    aws.Int32(c.lambda.MemorySize()),
		EphemeralStorage: &types.EphemeralStorage{
			Size: aws.Int32(c.lambda.EphemeralStorage()),
		},
		Code: &types.FunctionCode{
			ImageUri: aws.String(image.Uri),
		},
		VpcConfig: &types.VpcConfig{
			SecurityGroupIds: c.vpc.SecurityGroupIds(),
			SubnetIds:        c.vpc.SubnetIds(),
		},
		Environment: &types.Environment{
			Variables: c.lambda.Env(),
		},
		TracingConfig: &types.TracingConfig{
			Mode: types.TracingModePassThrough,
		},
		LoggingConfig: &types.LoggingConfig{
			LogGroup: aws.String(c.cloudwatch.Name()),
		},
	}

	update := struct {
		Role   *lambda.UpdateFunctionConfigurationInput
		Config *lambda.UpdateFunctionConfigurationInput
		Code   *lambda.UpdateFunctionCodeInput
	}{
		Config: &lambda.UpdateFunctionConfigurationInput{
			FunctionName: create.FunctionName,
			Role:         create.Role,
			MemorySize:   create.MemorySize,
			Timeout:      create.Timeout,
			EphemeralStorage: &types.EphemeralStorage{
				Size: create.EphemeralStorage.Size,
			},
			VpcConfig:     create.VpcConfig,
			Environment:   create.Environment,
			TracingConfig: create.TracingConfig,
			LoggingConfig: &types.LoggingConfig{
				LogGroup: create.LoggingConfig.LogGroup,
			},
		},
		Code: &lambda.UpdateFunctionCodeInput{
			FunctionName:  create.FunctionName,
			ImageUri:      create.Code.ImageUri,
			Architectures: create.Architectures,
		},
		Role: &lambda.UpdateFunctionConfigurationInput{
			FunctionName: create.FunctionName,
			Role:         aws.String(c.iam.RoleArn()),
		},
	}

	updateRetryBehavior := &lambda.PutFunctionEventInvokeConfigInput{
		FunctionName:         create.FunctionName,
		MaximumRetryAttempts: aws.Int32(c.lambda.Retries()),
	}

	if len(create.VpcConfig.SecurityGroupIds) == 0 && len(create.VpcConfig.SubnetIds) == 0 {
		// Due to peculiarities in the lambda api, we need to pass empty slices instead of nil slices to the update config.
		// Unfortunately, the create function struct has different type behavior than the update function struct.
		// We've chosed to handle this in the update struct as opposed to the create struct.
		update.Config.VpcConfig = &types.VpcConfig{
			SecurityGroupIds: []string{},
			SubnetIds:        []string{},
		}
	}

	tags := &lambda.TagResourceInput{
		Resource: aws.String(c.lambda.FunctionArn()),
		Tags:     c.lambda.Tags(),
	}

	_, err = c.lambda.Client().CreateFunction(ctx, create, RetryCreate)
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "ResourceConflictException":
			break
		default:
			return nil, err
		}
	}

	_, err = c.lambda.Client().UpdateFunctionConfiguration(ctx, update.Config, RetryUpdate)
	if err != nil {
		return nil, err
	}

	_, err = c.lambda.Client().UpdateFunctionCode(ctx, update.Code, RetryUpdate)
	if err != nil {
		return nil, err
	}

	_, err = c.lambda.Client().UpdateFunctionConfiguration(ctx, update.Role, RetryUpdate)
	if err != nil {
		return nil, err
	}

	_, err = c.lambda.Client().PutFunctionEventInvokeConfig(ctx, updateRetryBehavior)
	if err != nil {
		return nil, err
	}

	_, err = c.lambda.Client().TagResource(ctx, tags)
	if err != nil {
		return nil, err
	}

	return c.lambda.Client().GetFunction(ctx, read)
}

// DELETE Operations
func (c *Client) DeleteFunction(ctx context.Context) (*lambda.DeleteFunctionOutput, error) {
	var apiErr smithy.APIError

	log.Info().
		Str("action", "delete").
		Str("name", c.lambda.FunctionName()).
		Msg("lambda")

	delete := &lambda.DeleteFunctionInput{
		FunctionName: aws.String(c.lambda.FunctionName()),
	}

	_, err := c.lambda.Client().DeleteFunction(ctx, delete)
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "ResourceNotFoundException":
			return nil, nil
		default:
			return nil, err
		}
	}

	return nil, err
}

// Util

func RetryCreate(options *lambda.Options) {
	options.Retryer = retry.AddWithErrorCodes(options.Retryer,
		(*types.InvalidParameterValueException)(nil).ErrorCode(),
	)
	options.Retryer = retry.AddWithMaxAttempts(options.Retryer, 15)
}

func RetryUpdate(options *lambda.Options) {
	options.Retryer = retry.AddWithErrorCodes(options.Retryer,
		(*types.ResourceConflictException)(nil).ErrorCode(),
	)
	options.Retryer = retry.AddWithMaxAttempts(options.Retryer, 15)
}
