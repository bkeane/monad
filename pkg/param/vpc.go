package param

import (
	v "github.com/go-ozzo/ozzo-validation/v4"
)

type Vpc struct {
	SecurityGroupIds []string `arg:"--vpc-sg-ids" placeholder:"list" help:"sg-123,sg-456... [default: []]"`
	SubnetIds        []string `arg:"--vpc-sn-ids" placeholder:"list" help:"subnet-123,subnet-456... [default: []]"`
}

func (c *Vpc) Validate() error {
	return v.ValidateStruct(c,
		v.Field(&c.SecurityGroupIds, v.When(len(c.SecurityGroupIds) != 0, v.Required)),
		v.Field(&c.SubnetIds, v.When(len(c.SubnetIds) != 0, v.Required)),
	)
}
