package registry

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/bkeane/monad/internal/registryv2"
	"github.com/rs/zerolog/log"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/smithy-go"
)

type EcrConfig interface {
	Clients() (*ecr.Client, *registryv2.Client)
	ImagePath() string
	ImageTag() string
	RegistryId() string
}

type ImageRegistry interface {
	GetImage(ctx context.Context) (registryv2.ImagePointer, error)
	ImagePath() string
	ImageTag() string
}

type Client struct {
	config     EcrConfig
	ecr        *ecr.Client
	registryv2 *registryv2.Client
}

func Derive(config EcrConfig) *Client {
	var client Client
	client.config = config
	client.ecr, client.registryv2 = config.Clients()
	return &client
}

func (c *Client) GetImage(ctx context.Context) (registryv2.ImagePointer, error) {
	repo := c.config.ImagePath()
	tag := c.config.ImageTag()

	log.Info().
		Str("action", "get").
		Str("repo", repo).
		Str("tag", tag).
		Msg("registry")

	return c.registryv2.GetImage(ctx, repo, tag)
}

func (c *Client) ImagePath() string {
	return c.config.ImagePath()
}

func (c *Client) ImageTag() string {
	return c.config.ImageTag()
}

func (c *Client) Login(ctx context.Context) error {
	input := &ecr.GetAuthorizationTokenInput{
		RegistryIds: []string{c.config.RegistryId()},
	}

	if c.registryv2.Url == "" {
		return fmt.Errorf("missing registry URL, likely due to not validating ECR config")
	}

	output, err := c.ecr.GetAuthorizationToken(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to get authorization token: %w", err)
	}

	if output == nil || output.AuthorizationData == nil {
		return fmt.Errorf("missing AuthorizationData in ECR response")
	}

	if len(output.AuthorizationData) != 1 || output.AuthorizationData[0].AuthorizationToken == nil {
		return fmt.Errorf("missing AuthorizationToken in ECR response")
	}

	username, password, err := parseToken(output.AuthorizationData[0].AuthorizationToken)
	if err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}

	cmd := exec.Command("docker", "login", "--username", username, "--password-stdin", c.registryv2.Url)
	cmd.Stdin = strings.NewReader(password)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error logging in to Docker: %v\n", err)
		os.Exit(1)
	}

	return nil
}

func (c *Client) Untag(ctx context.Context) error {
	repo := c.config.ImagePath()
	tag := c.config.ImageTag()

	log.Info().
		Str("action", "untag").
		Str("repo", repo).
		Str("tag", tag).
		Msg("registry")

	return c.registryv2.Untag(ctx, repo, tag)
}

func (c *Client) CreateRepository(ctx context.Context) error {
	var apiErr smithy.APIError
	repo := c.config.ImagePath()

	log.Info().
		Str("action", "put").
		Str("repo", repo).
		Msg("registry")

	err := c.registryv2.CreateRepository(ctx, repo)
	if err != nil {
		switch errors.As(err, &apiErr) {
		case apiErr.ErrorCode() == "RepositoryAlreadyExistsException":
			log.Warn().
				Str("repo", repo).
				Msg("repository already exists")
			return nil
		default:
			return err
		}
	}

	return nil
}

func (c *Client) DeleteRepository(ctx context.Context) error {
	var apiErr smithy.APIError
	repo := c.config.ImagePath()

	log.Info().
		Str("action", "delete").
		Str("repo", repo).
		Msg("registry")

	err := c.registryv2.DeleteRepository(ctx, repo)
	if err != nil {
		switch errors.As(err, &apiErr) {
		case apiErr.ErrorCode() == "RepositoryNotFoundException":
			log.Warn().
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