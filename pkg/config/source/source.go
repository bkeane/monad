package source

import (
	"context"

	"github.com/bkeane/monad/internal/git"
	"github.com/bkeane/monad/pkg/schema"
	"github.com/bkeane/substrate/pkg/env"
	"github.com/bkeane/substrate/pkg/substrate"
	"github.com/rs/zerolog"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type Config struct {
	AwsConfig aws.Config `json:"-"`
	Substrate *substrate.Substrate
	schema    schema.Spec
	log       *zerolog.Logger
}

func (s Config) Parse(ctx context.Context, awsconfig aws.Config, git *git.Git, substrate *substrate.Substrate) (*Config, error) {
	c := &Config{}
	c.log = zerolog.Ctx(ctx)
	c.AwsConfig = awsconfig
	c.Substrate = substrate

	c.schema = schema.Version["latest"]
	if err := c.schema.Encode(*git); err != nil {
		return nil, err
	}

	return c, nil
}

func (s *Config) ResourceNamePrefix() string {
	return s.schema.ResourceNamePrefix(s.Substrate.Options.OrgPrefixedNames)
}

func (s *Config) ResourceName() string {
	return s.schema.ResourceName(s.Substrate.Options.OrgPrefixedNames)
}

func (s *Config) ResourcePath() string {
	return s.schema.ResourcePath(s.Substrate.Options.OrgPrefixedNames)
}

func (s *Config) ResourcePathPrefix() string {
	return s.schema.ResourcePathPrefix(s.Substrate.Options.OrgPrefixedNames)
}

func (s *Config) ImageBranchTag(registry string) string {
	return s.schema.ImageBranchTag(registry)
}

func (s *Config) ImageTags(registry string) []string {
	return s.schema.ImageTags(registry)
}

func (s *Config) ImagePath() string {
	return s.schema.ImagePath()
}

func (s *Config) Origin() string {
	return s.schema.Git().Origin
}

func (s *Config) Labels() map[string]string {
	return s.schema.EncodedMap()
}

func (s *Config) TemplateData() (env.Account, env.ECR, env.EventBridge, env.Lambda, env.ApiGateway, git.Git) {
	return s.Substrate.Account, s.Substrate.ECR, s.Substrate.EventBridge, s.Substrate.Lambda, s.Substrate.ApiGateway, s.schema.Git()
}
