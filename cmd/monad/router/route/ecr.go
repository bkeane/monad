package route

import (
	"context"
	"fmt"

	"github.com/bkeane/monad/pkg/param"
)

type SubCommand struct{}

type Ecr struct {
	param.RegistryConfig
	Init   *SubCommand `arg:"subcommand:init" help:"initialize a repository"`
	Delete *SubCommand `arg:"subcommand:delete" help:"delete a repository"`
	Tag    *SubCommand `arg:"subcommand:tag" help:"tag an image"`
	Untag  *SubCommand `arg:"subcommand:untag" help:"untag an image"`
	Login  *SubCommand `arg:"subcommand:login" help:"login to a registry"`
}

func (e *Ecr) Route(ctx context.Context, r Root) error {
	if err := e.RegistryConfig.Process(ctx, r.AwsConfig); err != nil {
		return err
	}

	switch {
	case r.Ecr.Login != nil:
		if err := e.RegistryConfig.Login(ctx); err != nil {
			return err
		}
	case r.Ecr.Untag != nil:
		if err := e.RegistryConfig.Untag(ctx, r.ServiceConfig.ImagePath, r.ServiceConfig.ImageTag); err != nil {
			return err
		}
	case r.Ecr.Tag != nil:
		fmt.Printf("%s/%s", e.RegistryConfig.Client.Url, r.ServiceConfig.Image)
	case r.Ecr.Init != nil:
		if err := e.RegistryConfig.CreateRepository(ctx, r.ServiceConfig.ImagePath); err != nil {
			return err
		}
	case r.Ecr.Delete != nil:
		if err := e.RegistryConfig.DeleteRepository(ctx, r.ServiceConfig.ImagePath); err != nil {
			return err
		}
	}

	return nil
}
