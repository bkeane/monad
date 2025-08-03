package lambda

import (
	"context"
	"errors"
	"fmt"

	"github.com/bkeane/monad/internal/registry"
	"github.com/rs/zerolog/log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/smithy-go"
)

type LambdaResources interface {
	Environment() (map[string]string, error)
	Timeout() int32
	MemorySize() int32
	EphemeralStorage() int32
	FunctionArn() string
	Retries() int32
	Client() *lambda.Client
}

type IAMResources interface {
	EniRoleArn() string
	RoleArn() string
}

type VpcResources interface {
	SecurityGroupIds() []string
	SubnetIds() []string
}

type CloudWatchResources interface {
	LogGroup() string
}

type ServiceResources interface {
	ImagePath() string
	ImageTag() string
}

type EcrResources interface {
	Client() *registry.Client
}

type SchemaResources interface {
	Name() string
	Tags() map[string]string
}

type Client struct {
	lambda     LambdaResources
	service    ServiceResources
	ecr        EcrResources
	iam        IAMResources
	vpc        VpcResources
	cloudwatch CloudWatchResources
	schema     SchemaResources
}

func Init(lambda LambdaResources, service ServiceResources, ecr EcrResources, iam IAMResources, vpc VpcResources, cloudwatch CloudWatchResources, schema SchemaResources) *Client {
	return &Client{
		lambda:     lambda,
		service:    service,
		ecr:        ecr,
		iam:        iam,
		vpc:        vpc,
		cloudwatch: cloudwatch,
		schema:     schema,
	}
}

func (s *Client) Mount(ctx context.Context) error {
	if _, err := s.PutFunction(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Client) Unmount(ctx context.Context) error {
	if _, err := s.DeleteFunction(ctx); err != nil {
		return err
	}

	return nil
}

// GET Operations
func (s *Client) GetImage(ctx context.Context) (registry.ImagePointer, error) {
	return s.ecr.Client().GetImage(ctx, s.service.ImagePath(), s.service.ImageTag())
}

// PUT Operations
func (s *Client) PutFunction(ctx context.Context) (*lambda.GetFunctionOutput, error) {
	var apiErr smithy.APIError
	var architecture []types.Architecture

	log.Info().
		Str("action", "put").
		Str("name", s.schema.Name()).
		Msg("lambda")

	image, err := s.GetImage(ctx)
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

	env, err := s.lambda.Environment()
	if err != nil {
		return nil, err
	}

	read := &lambda.GetFunctionInput{
		FunctionName: aws.String(s.schema.Name()),
	}

	create := &lambda.CreateFunctionInput{
		FunctionName:  aws.String(s.schema.Name()),
		Role:          aws.String(s.iam.EniRoleArn()),
		Tags:          s.schema.Tags(),
		Architectures: architecture,
		PackageType:   types.PackageTypeImage,
		Timeout:       aws.Int32(s.lambda.Timeout()),
		MemorySize:    aws.Int32(s.lambda.MemorySize()),
		EphemeralStorage: &types.EphemeralStorage{
			Size: aws.Int32(s.lambda.EphemeralStorage()),
		},
		Code: &types.FunctionCode{
			ImageUri: aws.String(image.Uri),
		},
		VpcConfig: &types.VpcConfig{
			SecurityGroupIds: s.vpc.SecurityGroupIds(),
			SubnetIds:        s.vpc.SubnetIds(),
		},
		Environment: &types.Environment{
			Variables: env,
		},
		TracingConfig: &types.TracingConfig{
			Mode: types.TracingModePassThrough,
		},
		LoggingConfig: &types.LoggingConfig{
			LogGroup: aws.String(s.cloudwatch.LogGroup()),
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
			Role:         aws.String(s.iam.RoleArn()),
		},
	}

	updateRetryBehavior := &lambda.PutFunctionEventInvokeConfigInput{
		FunctionName:         create.FunctionName,
		MaximumRetryAttempts: aws.Int32(s.lambda.Retries()),
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
		Resource: aws.String(s.lambda.FunctionArn()),
		Tags:     s.schema.Tags(),
	}

	_, err = s.lambda.Client().CreateFunction(ctx, create, RetryCreate)
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "ResourceConflictException":
			break
		default:
			return nil, err
		}
	}

	_, err = s.lambda.Client().UpdateFunctionConfiguration(ctx, update.Config, RetryUpdate)
	if err != nil {
		return nil, err
	}

	_, err = s.lambda.Client().UpdateFunctionCode(ctx, update.Code, RetryUpdate)
	if err != nil {
		return nil, err
	}

	_, err = s.lambda.Client().UpdateFunctionConfiguration(ctx, update.Role, RetryUpdate)
	if err != nil {
		return nil, err
	}

	_, err = s.lambda.Client().PutFunctionEventInvokeConfig(ctx, updateRetryBehavior)
	if err != nil {
		return nil, err
	}

	_, err = s.lambda.Client().TagResource(ctx, tags)
	if err != nil {
		return nil, err
	}

	return s.lambda.Client().GetFunction(ctx, read)
}

// DELETE Operations
func (s *Client) DeleteFunction(ctx context.Context) (*lambda.DeleteFunctionOutput, error) {
	var apiErr smithy.APIError

	log.Info().
		Str("action", "delete").
		Str("name", s.schema.Name()).
		Msg("lambda")

	delete := &lambda.DeleteFunctionInput{
		FunctionName: aws.String(s.schema.Name()),
	}

	_, err := s.lambda.Client().DeleteFunction(ctx, delete)
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
