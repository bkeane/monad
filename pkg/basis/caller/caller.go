package caller

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	v "github.com/go-ozzo/ozzo-validation/v4"
)

// Basis

type Basis struct {
	AwsConfig aws.Config
	AccountId string
	Arn       string
}

//
// Derive
//

func Derive(ctx context.Context) (*Basis, error) {
	var err error
	var basis Basis

	basis.AwsConfig, err = config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	client := sts.NewFromConfig(basis.AwsConfig)

	caller, err := client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, err
	}

	basis.AccountId = *caller.Account
	basis.Arn = *caller.Arn

	return &basis, nil
}

//
// Validations
//

func (c *Basis) Validate() error {
	return v.ValidateStruct(c,
		v.Field(&c.AwsConfig, v.Required),
		v.Field(&c.AccountId, v.Required),
		v.Field(&c.Arn, v.Required),
	)
}
