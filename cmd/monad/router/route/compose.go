package route

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/bkeane/monad/pkg/param"
	ctypes "github.com/compose-spec/compose-go/types"
	"gopkg.in/yaml.v2"
)

type Compose struct {
	param.Registry
	Dockerfile   string `arg:"-f,--file" help:"path to dockerfile"`
	BuildContext string `arg:"--build-context" help:"path to build context"`
	Login        bool   `arg:"--login" help:"login to registry"`
}

func (c *Compose) Route(ctx context.Context, r Root) error {
	var err error

	if err := c.Registry.Validate(ctx, r.AwsConfig); err != nil {
		return err
	}

	if c.Dockerfile == "" {
		c.Dockerfile = "./Dockerfile"
	}

	c.Dockerfile, err = filepath.Abs(c.Dockerfile)
	if err != nil {
		return err
	}

	c.BuildContext, err = filepath.Abs(c.BuildContext)
	if err != nil {
		return err
	}

	name := r.Git.Service
	image := fmt.Sprintf("%s/%s/%s/%s:%s", c.Registry.Client.Url, r.Git.Owner, r.Git.Repository, r.Git.Service, r.Git.Branch)

	compose := &ctypes.Config{}

	service := ctypes.ServiceConfig{
		Name:  name,
		Image: image,
		Build: &ctypes.BuildConfig{
			Context:    c.BuildContext,
			Dockerfile: c.Dockerfile,
			Platforms: []string{
				"linux/amd64",
				"linux/arm64",
			},
		},
	}

	compose.Name = "build"
	compose.Services = []ctypes.ServiceConfig{
		service,
	}

	yamlBytes, err := yaml.Marshal(compose)
	if err != nil {
		return err
	}

	composeData := string(yamlBytes)
	fmt.Println(composeData)

	return nil
}
