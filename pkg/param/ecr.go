package param

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/bkeane/monad/pkg/registry"
	"github.com/rs/zerolog/log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/smithy-go"
	v "github.com/go-ozzo/ozzo-validation/v4"
)

type RegistryConfig struct {
	ecrc   *ecr.Client      `arg:"-" json:"-"`
	Client *registry.Client `arg:"-" json:"-"`
	Id     string           `arg:"--ecr-id,env:MONAD_REGISTRY_ID" placeholder:"id" help:"ecr registry id" default:"caller-account-id"`
	Region string           `arg:"--ecr-region,env:MONAD_REGISTRY_REGION" placeholder:"name" help:"ecr registry region" default:"caller-region"`
}

func (r *RegistryConfig) Validate(ctx context.Context, awsconfig aws.Config) error {
	var err error

	r.ecrc = ecr.NewFromConfig(awsconfig)

	if r.Id == "" {
		input := &ecr.DescribeRegistryInput{}
		output, err := r.ecrc.DescribeRegistry(ctx, input)
		if err != nil {
			return err
		}

		r.Id = *output.RegistryId
	}

	if r.Region == "" {
		r.Region = awsconfig.Region
	}

	r.Client, err = registry.InitEcr(ctx, awsconfig, r.Id, r.Region)
	if err != nil {
		return err
	}

	return v.ValidateStruct(r,
		v.Field(&r.Id, v.Required),
		v.Field(&r.Region, v.Required),
	)
}

func (r *RegistryConfig) Login(ctx context.Context) error {
	input := &ecr.GetAuthorizationTokenInput{
		RegistryIds: []string{r.Id},
	}

	// Make validation of paramters structs the initializer for said structs so that this can never happen.
	if r.Client.Url == "" {
		return fmt.Errorf("missing registry url, likely due to not validating param.Registry")
	}

	output, err := r.ecrc.GetAuthorizationToken(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to get authorization token: %w", err)
	}

	if output == nil || output.AuthorizationData == nil {
		return fmt.Errorf("missing AuthorizationData in ECR Public response")
	}

	if len(output.AuthorizationData) != 1 || output.AuthorizationData[0].AuthorizationToken == nil {
		return fmt.Errorf("missing AuthorizationToken in ECR Public response")
	}

	username, password, err := parseToken(output.AuthorizationData[0].AuthorizationToken)
	if err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}

	cmd := exec.Command("docker", "login", "--username", username, "--password-stdin", r.Client.Url)
	cmd.Stdin = strings.NewReader(password)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error logging in to Docker: %v\n", err)
		os.Exit(1)
	}

	return nil
}

func (r *RegistryConfig) Untag(ctx context.Context, repo, tag string) error {
	log.Info().
		Str("action", "untag").
		Str("registry", r.Id).
		Str("region", r.Region).
		Str("repo", repo).
		Str("tag", tag).
		Msg("ecr")

	return r.Client.Untag(ctx, repo, tag)
}

func (r *RegistryConfig) GetImage(ctx context.Context, repo, tag string) (registry.ImagePointer, error) {
	log.Info().
		Str("action", "get").
		Str("registry", r.Id).
		Str("region", r.Region).
		Str("repo", repo).
		Str("tag", tag).
		Msg("ecr")

	return r.Client.FromPath(ctx, repo, tag)
}

func (r *RegistryConfig) CreateRepository(ctx context.Context, repo string) error {
	var apiErr smithy.APIError

	log.Info().
		Str("action", "put").
		Str("registry", r.Id).
		Str("region", r.Region).
		Str("repo", repo).
		Msg("ecr")

	err := r.Client.CreateRepository(ctx, repo)
	if err != nil {
		switch errors.As(err, &apiErr) {
		case apiErr.ErrorCode() == "RepositoryAlreadyExistsException":
			log.Warn().
				Str("registry", r.Id).
				Str("region", r.Region).
				Str("repo", repo).
				Msg("repository already exists")
			return nil

		default:
			return err

		}
	}

	return nil
}

func (r *RegistryConfig) DeleteRepository(ctx context.Context, repo string) error {
	var apiErr smithy.APIError

	log.Info().
		Str("action", "delete").
		Str("registry", r.Id).
		Str("region", r.Region).
		Str("repo", repo).
		Msg("ecr")

	err := r.Client.DeleteRepository(ctx, repo)
	if err != nil {
		switch errors.As(err, &apiErr) {
		case apiErr.ErrorCode() == "RepositoryNotFoundException":
			log.Warn().
				Str("registry", r.Id).
				Str("region", r.Region).
				Str("repo", repo).
				Msg("repository not found")
			return nil
		default:
			return err
		}
	}

	return nil
}

func parseToken(token *string) (username string, password string, err error) {
	decodedToken, err := base64.StdEncoding.DecodeString(*token)
	if err != nil {
		return "", "", fmt.Errorf("invalid token: %w", err)
	}

	parts := strings.SplitN(string(decodedToken), ":", 2)
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid token: expected two parts, got %d", len(parts))
	}

	return parts[0], parts[1], nil
}
