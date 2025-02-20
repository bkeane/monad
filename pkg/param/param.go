package param

import (
	"context"
	"embed"

	"github.com/aws/aws-sdk-go-v2/aws"
)

//go:embed defaults/*
var defaults embed.FS

func ReadDefault(name string) (string, error) {
	content, err := defaults.ReadFile(name)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

type Param interface {
	Validate() error
}

type AwsParam interface {
	Validate(ctx context.Context, awsconfig aws.Config) error
}
