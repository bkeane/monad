package registry

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/bkeane/monad/internal/registryv2"
	"github.com/bkeane/monad/pkg/basis/git"
	"github.com/bkeane/monad/pkg/basis/service"
	"github.com/dustin/go-humanize"
	"github.com/rs/zerolog/log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/smithy-go"
	"github.com/charmbracelet/lipgloss/table"
)

type EcrConfig interface {
	Clients() (*ecr.Client, *registryv2.Client)
	ImagePath() string
	ImageTag() string
	RegistryId() string
	Git() *git.Basis
	Service() *service.Basis
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

//
// ArtifactMetadata
//

type ArtifactMetadata struct {
	Repository string
	Tag        string
	Digest     string
	Size       int64
	PushedAt   *time.Time
	Owner      string
	Repo       string
	Service    string
}

//
// List
//

func (c *Client) List(ctx context.Context) ([]*ArtifactMetadata, error) {
	var artifacts []*ArtifactMetadata
	var nextToken *string

	// Paginate through all repositories
	for {
		input := &ecr.DescribeRepositoriesInput{
			MaxResults: aws.Int32(100),
			NextToken:  nextToken,
		}

		repos, err := c.ecr.DescribeRepositories(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to describe repositories: %w", err)
		}


		// For each repository, list its images
		for _, repo := range repos.Repositories {
			repoName := aws.ToString(repo.RepositoryName)

			// Parse repository name to extract owner/repo/service
			owner, repoField, serviceName := c.parseRepositoryName(repoName)

			// Apply filtering
			metadata := &ArtifactMetadata{
				Repository: repoName,
				Owner:      owner,
				Repo:       repoField,
				Service:    serviceName,
			}

			if !c.matchesFilter(metadata) {
				continue
			}

			// Get detailed image information for this repository with pagination
			var imageNextToken *string
			totalTaggedCount := 0
			totalUntaggedCount := 0

			for {
				detailInput := &ecr.DescribeImagesInput{
					RepositoryName: aws.String(repoName),
					MaxResults:     aws.Int32(100),
					NextToken:      imageNextToken,
				}

				details, err := c.ecr.DescribeImages(ctx, detailInput)
				if err != nil {
					log.Warn().
						Str("repository", repoName).
						Err(err).
						Msg("failed to describe images for repository")
					break
				}

				// Create artifact metadata for each image detail in this batch
				taggedCount := 0
				untaggedCount := 0
				for _, detail := range details.ImageDetails {
					// Skip images without tags
					if detail.ImageTags == nil || len(detail.ImageTags) == 0 {
						untaggedCount++
						continue
					}

					taggedCount++

					// Create an artifact for each tag (some images have multiple tags)
					for _, tag := range detail.ImageTags {
						artifact := &ArtifactMetadata{
							Repository: repoName,
							Owner:      owner,
							Repo:       repoField,
							Service:    serviceName,
							Tag:        tag,
						}

						if detail.ImageDigest != nil {
							artifact.Digest = aws.ToString(detail.ImageDigest)
						}

						if detail.ImageSizeInBytes != nil {
							artifact.Size = aws.ToInt64(detail.ImageSizeInBytes)
						}

						if detail.ImagePushedAt != nil {
							artifact.PushedAt = detail.ImagePushedAt
						}

						artifacts = append(artifacts, artifact)
					}
				}

				totalTaggedCount += taggedCount
				totalUntaggedCount += untaggedCount

				// Check if there are more images to fetch
				imageNextToken = details.NextToken
				if imageNextToken == nil {
					break
				}
			}
		}

		// Check if there are more repositories to fetch
		nextToken = repos.NextToken
		if nextToken == nil {
			break
		}
	}

	return artifacts, nil
}

func (c *Client) Table(ctx context.Context) (string, error) {
	artifacts, err := c.List(ctx)
	if err != nil {
		return "", err
	}

	// Sort artifacts by repository name, then by push date (newest first)
	sort.Slice(artifacts, func(i, j int) bool {
		if artifacts[i].Repository != artifacts[j].Repository {
			return artifacts[i].Repository < artifacts[j].Repository
		}

		// Within same repository, sort by push date (newest first)
		if artifacts[i].PushedAt != nil && artifacts[j].PushedAt != nil {
			return artifacts[i].PushedAt.After(*artifacts[j].PushedAt)
		}
		if artifacts[i].PushedAt != nil && artifacts[j].PushedAt == nil {
			return true
		}
		if artifacts[i].PushedAt == nil && artifacts[j].PushedAt != nil {
			return false
		}

		// Fall back to tag comparison
		return artifacts[i].Tag < artifacts[j].Tag
	})

	tbl := table.New()
	tbl.Headers("Repository", "Tag", "Digest", "Size", "Released")

	for _, artifact := range artifacts {
		digestDisplay := truncateDigest(artifact.Digest)
		sizeDisplay := formatSize(artifact.Size)
		pushedDisplay := formatTime(artifact.PushedAt)

		tbl.Row(artifact.Repository, artifact.Tag, digestDisplay, sizeDisplay, pushedDisplay)
	}

	return tbl.Render(), nil
}

//
// Helpers
//

// parseRepositoryName extracts owner, repo, and service from repository name
// Assumes format: owner/repo/service or owner/repo
func (c *Client) parseRepositoryName(repoName string) (owner, repo, service string) {
	parts := strings.Split(repoName, "/")

	if len(parts) >= 2 {
		owner = parts[0]
		repo = parts[1]
	}

	if len(parts) >= 3 {
		service = parts[2]
	}

	return owner, repo, service
}

// matchesFilter checks if artifact metadata matches the basis filter values
// * means match all for that field
func (c *Client) matchesFilter(metadata *ArtifactMetadata) bool {
	gitBasis := c.config.Git()
	serviceBasis := c.config.Service()

	// Check owner filter
	if gitBasis.Owner() != "*" && gitBasis.Owner() != metadata.Owner {
		return false
	}

	// Check repo filter
	if gitBasis.Repo() != "*" && gitBasis.Repo() != metadata.Repo {
		return false
	}

	// Check service filter - for ECR list, default to "*" to show all services
	serviceFilter := "*"  // Default to showing all services for ECR list

	// Override if service filter was explicitly provided
	if serviceBasis.Name() != "" {
		// Check if the service name was explicitly set via --service flag
		// vs defaulted from directory name by checking environment variable
		if os.Getenv("MONAD_SERVICE") != "" {
			serviceFilter = serviceBasis.Name()
		}
	}

	if serviceFilter != "*" && serviceFilter != metadata.Service {
		return false
	}

	return true
}

// truncateDigest shortens digest to 12 characters for display
func truncateDigest(digest string) string {
	if len(digest) <= 12 {
		return digest
	}
	// Skip sha256: prefix if present
	if strings.HasPrefix(digest, "sha256:") {
		digest = digest[7:]
	}
	if len(digest) > 12 {
		return digest[:12]
	}
	return digest
}

// formatSize converts bytes to human-readable format
func formatSize(size int64) string {
	if size == 0 {
		return "-"
	}
	return humanize.Bytes(uint64(size))
}

// formatTime formats time for display using human-friendly format
func formatTime(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return humanize.Time(*t)
}