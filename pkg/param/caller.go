package param

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	v "github.com/go-ozzo/ozzo-validation/v4"
)

type Caller struct {
	Client    *sts.Client `arg:"-" json:"-"`
	AccountId string      `arg:"-" json:"-"`
	Arn       string      `arg:"-" json:"-"`
	UserId    string      `arg:"-" json:"-"`
}

func (c *Caller) Validate(ctx context.Context, awsconfig aws.Config) error {
	c.Client = sts.NewFromConfig(awsconfig)

	caller, err := c.Client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return err
	}

	c.AccountId = *caller.Account
	c.Arn = *caller.Arn
	c.UserId = *caller.UserId

	return v.ValidateStruct(c,
		v.Field(&c.AccountId, v.Required),
		v.Field(&c.Arn, v.Required),
		v.Field(&c.UserId, v.Required),
	)
}
