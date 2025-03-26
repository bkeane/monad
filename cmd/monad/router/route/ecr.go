package route

import (
	"context"
	"fmt"

	"github.com/bkeane/monad/pkg/param"
)

type Untag struct {
	param.Target
}

type Tag struct {
	param.Target
}

type Login struct{}

type Ecr struct {
	param.Registry
	Untag *Untag `arg:"subcommand:untag"`
	Login *Login `arg:"subcommand:login"`
	Tag   *Tag   `arg:"subcommand:tag"`
}

func (e *Ecr) Route(ctx context.Context, r Root) error {
	if err := e.Registry.Validate(ctx, r.AwsConfig); err != nil {
		return err
	}

	switch {
	case r.Ecr.Login != nil:
		if err := e.Registry.Login(ctx); err != nil {
			return err
		}
	case r.Ecr.Untag != nil:
		if err := e.Untag.Target.Validate(r.Git); err != nil {
			return err
		}

		if err := e.Registry.Untag(ctx, e.Untag.Target.ImagePath, e.Untag.Target.ImageTag); err != nil {
			return err
		}
	case r.Ecr.Tag != nil:
		if err := e.Tag.Target.Validate(r.Git); err != nil {
			return err
		}

		fmt.Printf("%s/%s", e.Registry.Client.Url, e.Tag.Target.Image)
	}

	return nil
}
