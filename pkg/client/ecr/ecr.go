package ecr

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

type EcrConvention interface {
	Clients() (*ecr.Client, *registryv2.Client)
}

//
// Client
//

type Client struct {
	ecr        *ecr.Client
	registryv2 *registryv2.Client
}

//
// Init
//

func Init(ecr EcrConvention) *Client {
	var client Client
	client.ecr, client.registryv2 = ecr.Clients()
	return &client
}

// Login authenticates Docker with ECR registry
func (c *Client) Login(ctx context.Context) error {
	input := &ecr.GetAuthorizationTokenInput{
		RegistryIds: []string{},
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

// Untag removes a tag from a repository image
func (c *Client) Untag(ctx context.Context, repo, tag string) error {
	log.Info().
		Str("action", "untag").
		Str("repo", repo).
		Str("tag", tag).
		Msg("ecr")

	return c.registryv2.Untag(ctx, repo, tag)
}

// GetImage retrieves image information from registry
func (c *Client) GetImage(ctx context.Context, repo, tag string) (registryv2.ImagePointer, error) {
	log.Info().
		Str("action", "get").
		Str("repo", repo).
		Str("tag", tag).
		Msg("ecr")

	return c.registryv2.GetImage(ctx, repo, tag)
}

// CreateRepository creates a new ECR repository
func (c *Client) CreateRepository(ctx context.Context, repo string) error {
	var apiErr smithy.APIError

	log.Info().
		Str("action", "put").
		Str("repo", repo).
		Msg("ecr")

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

// DeleteRepository removes an ECR repository
func (c *Client) DeleteRepository(ctx context.Context, repo string) error {
	var apiErr smithy.APIError

	log.Info().
		Str("action", "delete").
		Str("repo", repo).
		Msg("ecr")

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

//
// Helpers
//

// parseToken decodes base64 ECR authorization token
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
