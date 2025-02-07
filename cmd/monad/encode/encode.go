package encode

import (
	"context"
	"path/filepath"

	"github.com/bkeane/monad/internal/git"
	"github.com/bkeane/monad/pkg/config/source"
	"github.com/bkeane/monad/pkg/config/tmpl"
	"github.com/bkeane/substrate/pkg/substrate"

	"github.com/aws/aws-sdk-go-v2/aws"
	ctypes "github.com/compose-spec/compose-go/types"
	"gopkg.in/yaml.v2"
)

type Dockerfile struct {
	Path string `arg:"positional" help:"path to dockerfile" default:"./Dockerfile"`
}

type Options struct {
	Substrate string `arg:"-s,--substrate" help:"substrate hub name" default:"agar"`
	Context   string `arg:"-c,--context" help:"set build context" default:"."`
	List      bool   `arg:"-l,--list" help:"list available template interpolations."`
}

type Overrides struct {
	Branch *string `arg:"-b,--branch" help:"override branch name in build"`
	Sha    *string `arg:"-s,--sha" help:"override sha in build"`
}

type Root struct {
	Options
	Overrides
	Dockerfile
	source    *source.Config
	substrate *substrate.Substrate
}

func (r *Root) Route(ctx context.Context, awsconfig aws.Config) (*string, error) {
	var err error

	path := filepath.Join(r.Options.Context, r.Dockerfile.Path)

	git, err := git.Git{}.Parse(path)
	if err != nil {
		return nil, err
	}

	r.substrate, err = substrate.Parse(ctx, awsconfig, r.Options.Substrate)
	if err != nil {
		return nil, err
	}

	if r.Overrides.Branch != nil {
		git.Branch = *r.Overrides.Branch
	}

	if r.Overrides.Sha != nil {
		git.Sha = *r.Overrides.Sha
	}

	r.source, err = source.Config{}.Parse(ctx, awsconfig, git, r.substrate)
	if err != nil {
		return nil, err
	}

	if r.Options.List {
		return tmpl.Init(r.source).Table()
	}

	return r.compose()
}

func (r *Root) compose() (*string, error) {
	c := &ctypes.Config{}
	name := r.source.ResourceName()
	image := r.source.ImageBranchTag(r.substrate.ECR.RegistryUrl())
	tags := r.source.ImageTags(r.substrate.ECR.RegistryUrl())

	service := ctypes.ServiceConfig{
		Name:  name,
		Image: image,
		Build: &ctypes.BuildConfig{
			Context:    r.Options.Context,
			Dockerfile: r.Dockerfile.Path,
			Labels:     r.source.Labels(),
			Tags:       tags,
			Platforms: []string{
				"linux/amd64",
				"linux/arm64",
			},
		},
	}

	c.Name = r.source.Origin()
	c.Services = []ctypes.ServiceConfig{
		service,
	}

	yamlBytes, err := yaml.Marshal(c)
	if err != nil {
		return nil, err
	}

	composeData := string(yamlBytes)
	return &composeData, nil
}
