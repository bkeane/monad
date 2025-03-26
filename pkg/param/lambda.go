package param

import (
	"context"
	"fmt"

	"github.com/bkeane/monad/internal/uriopt"
	v "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

type Lambda struct {
	Client           *lambda.Client `arg:"-" json:"-"`
	Region           string         `arg:"--lambda-region,env:MONAD_LAMBDA_REGION" placeholder:"name" help:"lambda region" default:"caller-region"`
	EnvTemplate      string         `arg:"--env,env:MONAD_ENV" placeholder:"template" help:"string | file://env.tmpl" default:"minimal-env"`
	EphemeralStorage int32          `arg:"--disk,env:MONAD_DISK" placeholder:"mb" help:"ephemeral storage size" default:"512"`
	MemorySize       int32          `arg:"--memory,env:MONAD_MEMORY" placeholder:"mb" help:"memory size" default:"128"`
	Timeout          int32          `arg:"--timeout,env:MONAD_TIMEOUT" placeholder:"seconds" help:"function timeout" default:"3"`
}

func (l *Lambda) Validate(ctx context.Context, awsconfig aws.Config) error {
	var err error

	l.Client = lambda.NewFromConfig(awsconfig)

	if l.EnvTemplate == "" {
		l.EnvTemplate, err = ReadDefault("defaults/.env.tmpl")
		if err != nil {
			return fmt.Errorf("failed to read default env template: %w", err)
		}

	} else {
		l.EnvTemplate, err = uriopt.String(l.EnvTemplate)
		if err != nil {
			return fmt.Errorf("failed to read provided env template: %w", err)
		}

	}

	if l.Region == "" {
		l.Region = awsconfig.Region
	}

	if l.EphemeralStorage == 0 {
		l.EphemeralStorage = int32(512)
	}

	if l.MemorySize == 0 {
		l.MemorySize = int32(128)
	}

	if l.Timeout == 0 {
		l.Timeout = int32(3)
	}

	return v.ValidateStruct(l,
		v.Field(&l.EnvTemplate, v.NilOrNotEmpty),
		v.Field(&l.Region, v.Required),
		v.Field(&l.EphemeralStorage, v.Required),
		v.Field(&l.MemorySize, v.Required),
		v.Field(&l.Timeout, v.Required),
	)
}
