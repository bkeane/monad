package caller

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	v "github.com/go-ozzo/ozzo-validation/v4"
)

// STSClient interface for dependency injection and testing
type STSClient interface {
	GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}

// Basis

type Basis struct {
	CallerConfig  *aws.Config
	CallerAccount *string
	CallerArn     *string
}

//
// Derive
//

// Derive creates a new caller Basis with AWS credential information
// Optionally accepts a custom STS client for testing
func Derive(ctx context.Context, clients ...STSClient) (*Basis, error) {
	var err error
	var basis Basis
	var client STSClient

	// Use provided client or create default
	if len(clients) > 0 {
		client = clients[0]
		// For testing, we still need a minimal config
		basis.CallerConfig = &aws.Config{}
	} else {
		awsconfig, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, err
		}
		basis.CallerConfig = &awsconfig
		client = sts.NewFromConfig(awsconfig)
	}

	caller, err := client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, err
	}

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
