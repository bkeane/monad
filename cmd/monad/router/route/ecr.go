package route

import (
	"context"
	"fmt"

	"github.com/bkeane/monad/pkg/model"
)

type SubCommand struct{}

type Ecr struct {
	Init   *SubCommand `arg:"subcommand:init" help:"initialize a repository"`
	Delete *SubCommand `arg:"subcommand:delete" help:"delete a repository"`
	Tag    *SubCommand `arg:"subcommand:tag" help:"tag an image"`
	Untag  *SubCommand `arg:"subcommand:untag" help:"untag an image"`
	Login  *SubCommand `arg:"subcommand:login" help:"login to a registry"`
}

func (e *Ecr) Route(ctx context.Context, r Root) error {
	// Initialize unified model
	model := &model.Model{}
	if err := model.Process(ctx, r.AwsConfig); err != nil {
		return err
	}

	switch {
	case r.Ecr.Login != nil:
		if err := model.ECR().Login(ctx); err != nil {
			return err
		}
	case r.Ecr.Untag != nil:
		if err := model.ECR().Untag(ctx, model.ECR().ImagePath(), model.ECR().ImageTag()); err != nil {
			return err
		}
	case r.Ecr.Tag != nil:
		fmt.Printf("%s/%s", model.ECR().Client().Url, model.Axiom().Service().Image())
	case r.Ecr.Init != nil:
		if err := model.ECR().CreateRepository(ctx, model.ECR().ImagePath()); err != nil {
			return err
		}
	case r.Ecr.Delete != nil:
		if err := model.ECR().DeleteRepository(ctx, model.ECR().ImagePath()); err != nil {
			return err
		}
	}

	return nil
}
