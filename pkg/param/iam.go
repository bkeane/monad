package param

import (
	"context"
	"fmt"

	"github.com/bkeane/monad/internal/uriopt"
	v "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

type Iam struct {
	Client         *iam.Client `arg:"-" json:"-"`
	PolicyTemplate string      `arg:"--policy" placeholder:"template" help:"{} | file://policy.tmpl" default:"minimal-policy"`
	RoleTemplate   string      `arg:"--role" placeholder:"template" help:"{} | file://role.tmpl" default:"minimal-role"`
}

func (l *Iam) Validate(ctx context.Context, awsconfig aws.Config) error {
	var err error

	l.Client = iam.NewFromConfig(awsconfig)

	if l.PolicyTemplate == "" {
		l.PolicyTemplate, err = ReadDefault("defaults/policy.json.tmpl")
		if err != nil {
			return fmt.Errorf("failed to read default policy template")
		}

	} else {
		l.PolicyTemplate, err = uriopt.Json(l.PolicyTemplate)
		if err != nil {
			return fmt.Errorf("failed to read provided policy template")
		}

	}

	if l.RoleTemplate == "" {
		l.RoleTemplate, err = ReadDefault("defaults/role.json.tmpl")
		if err != nil {
			return fmt.Errorf("failed to read default role template")
		}

	} else {
		l.RoleTemplate, err = uriopt.Json(l.RoleTemplate)
		if err != nil {
			return fmt.Errorf("failed to read provided role template")
		}

	}

	return v.ValidateStruct(l,
		v.Field(&l.PolicyTemplate, v.NilOrNotEmpty),
		v.Field(&l.RoleTemplate, v.NilOrNotEmpty),
	)
}
