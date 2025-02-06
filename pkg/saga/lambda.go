package saga

import (
	"context"
	"errors"

	"github.com/bkeane/monad/pkg/config/release"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/smithy-go"
	"github.com/rs/zerolog"
)

type Lambda struct {
	release release.Config
	lambda  *lambda.Client
	log     *zerolog.Logger
}

func (s Lambda) Init(ctx context.Context, r release.Config) *Lambda {
	return &Lambda{
		release: r,
		lambda:  lambda.NewFromConfig(r.AwsConfig),
		log:     zerolog.Ctx(ctx),
	}
}

func (s *Lambda) Do(ctx context.Context) error {
	s.log.Info().Msg("ensuring lambda function exists")
	if _, err := s.PutFunction(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Lambda) Undo(ctx context.Context) error {
	s.log.Info().Msg("deleting lambda function")
	if _, err := s.DeleteFunction(ctx); err != nil {
		return err
	}

	return nil
}

// PUT Operations

func (s *Lambda) PutFunction(ctx context.Context) (*lambda.GetFunctionOutput, error) {
	var apiErr smithy.APIError

	resources, err := s.release.Resources()
	if err != nil {
		return nil, err
	}

	read := &lambda.GetFunctionInput{
		FunctionName: aws.String(s.release.FunctionName()),
	}

	create := &lambda.CreateFunctionInput{
		FunctionName:  aws.String(s.release.FunctionName()),
		Role:          aws.String(s.release.EniRoleArn()),
		Tags:          s.release.ResourceTags(),
		Architectures: s.release.Image.Architecture,
		PackageType:   types.PackageTypeImage,
		Timeout:       aws.Int32(resources.Timeout),
		MemorySize:    aws.Int32(resources.MemorySize),
		EphemeralStorage: &types.EphemeralStorage{
			Size: aws.Int32(resources.EphemeralStorage),
		},
		Code: &types.FunctionCode{
			ImageUri: aws.String(s.release.ImageUri()),
		},
		VpcConfig: &types.VpcConfig{
			SecurityGroupIds: s.release.Substrate.VPC.SecurityGroupIds,
			SubnetIds:        s.release.Substrate.VPC.SubnetIds,
		},
		Environment: &types.Environment{
			Variables: resources.Env,
		},
		TracingConfig: &types.TracingConfig{
			Mode: types.TracingModePassThrough,
		},
	}

	update := struct {
		Role   *lambda.UpdateFunctionConfigurationInput
		Config *lambda.UpdateFunctionConfigurationInput
		Code   *lambda.UpdateFunctionCodeInput
	}{
		Config: &lambda.UpdateFunctionConfigurationInput{
			FunctionName:  create.FunctionName,
			Role:          create.Role,
			MemorySize:    create.MemorySize,
			Timeout:       create.Timeout,
			VpcConfig:     create.VpcConfig,
			Environment:   create.Environment,
			TracingConfig: create.TracingConfig,
		},
		Code: &lambda.UpdateFunctionCodeInput{
			FunctionName:  create.FunctionName,
			ImageUri:      create.Code.ImageUri,
			Architectures: create.Architectures,
		},
		Role: &lambda.UpdateFunctionConfigurationInput{
			FunctionName: create.FunctionName,
			Role:         aws.String(s.release.RoleArn()),
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
		Resource: aws.String(s.release.FunctionArn()),
		Tags:     s.release.ResourceTags(),
	}

	_, err = s.lambda.CreateFunction(ctx, create, RetryCreate)
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "ResourceConflictException":
			break
		default:
			return nil, err
		}
	}

	_, err = s.lambda.UpdateFunctionConfiguration(ctx, update.Config, RetryUpdate)
	if err != nil {
		return nil, err
	}

	_, err = s.lambda.UpdateFunctionCode(ctx, update.Code, RetryUpdate)
	if err != nil {
		return nil, err
	}

	_, err = s.lambda.UpdateFunctionConfiguration(ctx, update.Role, RetryUpdate)
	if err != nil {
		return nil, err
	}

	_, err = s.lambda.TagResource(ctx, tags)
	if err != nil {
		return nil, err
	}

	return s.lambda.GetFunction(ctx, read)
}

func (s *Lambda) DeleteFunction(ctx context.Context) (*lambda.DeleteFunctionOutput, error) {
	var apiErr smithy.APIError

	delete := &lambda.DeleteFunctionInput{
		FunctionName: aws.String(s.release.FunctionName()),
	}

	_, err := s.lambda.DeleteFunction(ctx, delete)
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
