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
	CallerConfig  *aws.Config
	CallerAccount *string
	CallerArn     *string
}

//
// Derive
//

func Derive(ctx context.Context) (*Basis, error) {
	var err error
	var basis Basis

	awsconfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	client := sts.NewFromConfig(awsconfig)
	caller, err := client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, err
	}

	basis.CallerConfig = &awsconfig
	basis.CallerAccount = caller.Account
	basis.CallerArn = caller.Arn

	err = basis.Validate()
	if err != nil {
		return nil, err
	}

	return &basis, err
}

//
// Validations
//

func (c *Basis) Validate() error {
	return v.ValidateStruct(c,
		v.Field(&c.CallerConfig, v.Required),
		v.Field(&c.CallerAccount, v.Required),
		v.Field(&c.CallerArn, v.Required),
	)
}

// Accessors

func (c *Basis) AwsConfig() aws.Config {
	return *c.CallerConfig
}

func (c *Basis) AccountId() string {
	return *c.CallerAccount
}

func (c *Basis) Arn() string {
	return *c.CallerArn
}
