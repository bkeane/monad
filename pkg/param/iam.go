package param

import (
	"context"
	"fmt"

	"github.com/bkeane/monad/internal/uriopt"
	v "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

type IamConfig struct {
	Client         *iam.Client `arg:"-" json:"-"`
	PolicyTemplate string      `arg:"--policy,env:MONAD_POLICY" placeholder:"template" help:"string | file://policy.tmpl" default:"minimal-policy"`
	RoleTemplate   string      `arg:"--role,env:MONAD_ROLE" placeholder:"template" help:"string | file://role.tmpl" default:"minimal-role"`
	BoundaryPolicy string      `arg:"--boundary,env:MONAD_BOUNDARY_POLICY" placeholder:"arn|name" help:"boundary policy" default:"no-boundary"`
}

func (l *IamConfig) Validate(ctx context.Context, awsconfig aws.Config) error {
	var err error

	l.Client = iam.NewFromConfig(awsconfig)

	if l.PolicyTemplate == "" {
		l.PolicyTemplate, err = ReadDefault("defaults/policy.json.tmpl")
		if err != nil {
			return fmt.Errorf("failed to read default policy template: %w", err)
		}

	} else {
		l.PolicyTemplate, err = uriopt.Json(l.PolicyTemplate)
		if err != nil {
			return fmt.Errorf("failed to read provided policy template: %w", err)
		}

	}

	if l.RoleTemplate == "" {
		l.RoleTemplate, err = ReadDefault("defaults/role.json.tmpl")
		if err != nil {
			return fmt.Errorf("failed to read default role template: %w", err)
		}

	} else {
		l.RoleTemplate, err = uriopt.Json(l.RoleTemplate)
		if err != nil {
			return fmt.Errorf("failed to read provided role template: %w", err)
		}

	}

	return v.ValidateStruct(l,
		v.Field(&l.PolicyTemplate, v.Required),
		v.Field(&l.RoleTemplate, v.Required),
	)
}
