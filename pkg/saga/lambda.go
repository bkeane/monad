package saga

import (
	"context"
	"errors"
	"fmt"

	"github.com/bkeane/monad/pkg/param"
	"github.com/bkeane/monad/pkg/registry"
	"github.com/rs/zerolog/log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/smithy-go"
)

type Lambda struct {
	config param.Aws
}

func (s Lambda) Init(ctx context.Context, c param.Aws) *Lambda {
	return &Lambda{
		config: c,
	}
}

func (s *Lambda) Do(ctx context.Context, image registry.ImagePointer) error {
	log.Info().
		Str("action", "put").
		Str("name", s.config.ResourceName()).
		Msg("lambda")

	if _, err := s.PutFunction(ctx, image); err != nil {
		return err
	}

	return nil
}

func (s *Lambda) Undo(ctx context.Context) error {
	log.Info().
		Str("action", "delete").
		Str("name", s.config.ResourceName()).
		Msg("lambda")

	if _, err := s.DeleteFunction(ctx); err != nil {
		return err
	}

	return nil
}

// PUT Operations
func (s *Lambda) PutFunction(ctx context.Context, image registry.ImagePointer) (*lambda.GetFunctionOutput, error) {
	var apiErr smithy.APIError
	var architecture []types.Architecture

	switch image.Architecture {
	case "amd64":
		architecture = []types.Architecture{types.ArchitectureX8664}
	case "arm64":
		architecture = []types.Architecture{types.ArchitectureArm64}
	default:
		return nil, fmt.Errorf("unsupported architecture: %s", image.Architecture)
	}

	env, err := s.config.Env()
	if err != nil {
		return nil, err
	}

	read := &lambda.GetFunctionInput{
		FunctionName: aws.String(s.config.ResourceName()),
	}

	create := &lambda.CreateFunctionInput{
		FunctionName:  aws.String(s.config.ResourceName()),
		Role:          aws.String(s.config.EniRoleArn()),
		Tags:          s.config.Tags(),
		Architectures: architecture,
		PackageType:   types.PackageTypeImage,
		Timeout:       aws.Int32(s.config.Lambda.Timeout),
		MemorySize:    aws.Int32(s.config.Lambda.MemorySize),
		EphemeralStorage: &types.EphemeralStorage{
			Size: aws.Int32(s.config.Lambda.EphemeralStorage),
		},
		Code: &types.FunctionCode{
			ImageUri: aws.String(image.Uri),
		},
		VpcConfig: &types.VpcConfig{
			SecurityGroupIds: s.config.Vpc.SecurityGroupIds,
			SubnetIds:        s.config.Vpc.SubnetIds,
		},
		Environment: &types.Environment{
			Variables: env,
		},
		TracingConfig: &types.TracingConfig{
			Mode: types.TracingModePassThrough,
		},
		LoggingConfig: &types.LoggingConfig{
			LogGroup: aws.String(s.config.CloudwatchLogGroup()),
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
			Role:         aws.String(s.config.RoleArn()),
		},
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
		Resource: aws.String(s.config.FunctionArn()),
		Tags:     s.config.Tags(),
	}

	_, err = s.config.Lambda.Client.CreateFunction(ctx, create, RetryCreate)
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "ResourceConflictException":
			break
		default:
			return nil, err
		}
	}

	_, err = s.config.Lambda.Client.UpdateFunctionConfiguration(ctx, update.Config, RetryUpdate)
	if err != nil {
		return nil, err
	}

	_, err = s.config.Lambda.Client.UpdateFunctionCode(ctx, update.Code, RetryUpdate)
	if err != nil {
		return nil, err
	}

	_, err = s.config.Lambda.Client.UpdateFunctionConfiguration(ctx, update.Role, RetryUpdate)
	if err != nil {
		return nil, err
	}

	_, err = s.config.Lambda.Client.TagResource(ctx, tags)
	if err != nil {
		return nil, err
	}

	return s.config.Lambda.Client.GetFunction(ctx, read)
}

func (s *Lambda) DeleteFunction(ctx context.Context) (*lambda.DeleteFunctionOutput, error) {
	var apiErr smithy.APIError

	delete := &lambda.DeleteFunctionInput{
		FunctionName: aws.String(s.config.ResourceName()),
	}

	_, err := s.config.Lambda.Client.DeleteFunction(ctx, delete)
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
