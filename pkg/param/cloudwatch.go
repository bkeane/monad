package param

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	v "github.com/go-ozzo/ozzo-validation/v4"
)

type CloudWatchConfig struct {
	Client    *cloudwatchlogs.Client `arg:"-" json:"-"`
	Region    string                 `arg:"--log-region,env:MONAD_LOG_REGION" placeholder:"name" help:"cloudwatch log region"  default:"caller-region"`
	Retention int32                  `arg:"--log-retention,env:MONAD_LOG_RETENTION" placeholder:"days" help:"1, 3, 5, 7, 14, 30..." default:"3"`
}

func (c *CloudWatchConfig) Validate(ctx context.Context, awsconfig aws.Config) error {
	c.Client = cloudwatchlogs.NewFromConfig(awsconfig)

	if c.Region == "" {
		c.Region = awsconfig.Region
	}

	if c.Retention == 0 {
		c.Retention = int32(3)
	}

	return v.ValidateStruct(c,
		v.Field(&c.Retention, v.In(int32(1), int32(3), int32(5), int32(7), int32(14), int32(30), int32(60), int32(90), int32(120), int32(150), int32(180), int32(365), int32(400), int32(545), int32(731), int32(1827), int32(3653))),
	)
}
