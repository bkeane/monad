package caller

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	v "github.com/go-ozzo/ozzo-validation/v4"
)

// Data

type Data struct {
	accountId string
	region    string
	arn       string
	userId    string
}

//
// Derive
//

func Derive(ctx context.Context, awsconfig aws.Config) (*Data, error) {
	var err error
	var data Data

	client := sts.NewFromConfig(awsconfig)

	caller, err := client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, err
	}

	data.accountId = *caller.Account
	data.arn = *caller.Arn
	data.userId = *caller.UserId
	data.region = awsconfig.Region

	if err := data.Validate(); err != nil {
		return nil, err
	}

	return &data, nil
}

//
// Validations
//

func (c *Data) Validate() error {
	return v.ValidateStruct(c,
		v.Field(&c.accountId, v.Required),
		v.Field(&c.arn, v.Required),
		v.Field(&c.userId, v.Required),
		v.Field(&c.region, v.Required),
	)
}

//
// Accessors
//

func (c *Data) AccountId() string { return c.accountId } // AWS account ID
func (c *Data) Region() string    { return c.region }    // AWS region
func (c *Data) Arn() string       { return c.arn }       // Caller ARN
func (c *Data) UserId() string    { return c.userId }    // Caller user ID
