package param

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
)

func ReadDefault(name string) (string, error) {
	content, err := Defaults.ReadFile(name)
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
