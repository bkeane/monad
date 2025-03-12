package param

import (
	"context"
	"fmt"
	"strings"

	"github.com/bkeane/monad/internal/tmpl"

	"github.com/aws/aws-sdk-go-v2/aws"
	dotenvlib "github.com/joho/godotenv"
)

type Aws struct {
	Git Git `arg:"-" json:"-"`
	Caller
	Registry
	Lambda
	Iam
	Vpc
	CloudWatch
	ApiGateway
	EventBridge
}

func (c *Aws) Validate(ctx context.Context, awsconfig aws.Config, git Git) error {
	c.Git = git

	if err := c.Caller.Validate(ctx, awsconfig); err != nil {
		return err
	}

	if err := c.Registry.Validate(ctx, awsconfig); err != nil {
		return err
	}

	if err := c.Lambda.Validate(ctx, awsconfig); err != nil {
		return err
	}

	if err := c.Iam.Validate(ctx, awsconfig); err != nil {
		return err
	}

	if err := c.Vpc.Validate(ctx, awsconfig); err != nil {
		return err
	}

	if err := c.CloudWatch.Validate(ctx, awsconfig); err != nil {
		return err
	}

	if err := c.ApiGateway.Validate(ctx, awsconfig); err != nil {
		return err
	}

	if err := c.EventBridge.Validate(ctx, awsconfig); err != nil {
		return err
	}

	return nil
}

// Computed Properties

func (c *Aws) ResourceName() string {
	branch := strings.ReplaceAll(c.Git.Branch, "/", "-")
	return fmt.Sprintf("%s-%s-%s", c.Git.Repository, branch, c.Git.Service)
}

func (c *Aws) ResourcePath() string {
	return fmt.Sprintf("%s/%s/%s", c.Git.Repository, c.Git.Branch, c.Git.Service)
}

func (c *Aws) CloudwatchLogGroup() string {
	return fmt.Sprintf("/aws/lambda/%s", c.ResourcePath())
}

func (c *Aws) CloudwatchLogRetention() int32 {
	return c.CloudWatch.LogRetention
}

func (c *Aws) Env() (map[string]string, error) {
	env, err := tmpl.Template("env", c.Lambda.EnvTemplate, c)
	if err != nil {
		return nil, err
	}

	return dotenvlib.Parse(strings.NewReader(env))
}

func (c *Aws) FunctionArn() string {
	return fmt.Sprintf("arn:aws:lambda:%s:%s:function:%s",
		c.Lambda.Region, c.Caller.AccountId, c.ResourceName())
}

func (c *Aws) RoleDocument() (string, error) {
	return tmpl.Template("role document", c.Iam.RoleTemplate, c)
}

func (c *Aws) RoleArn() string {
	return fmt.Sprintf("arn:aws:iam::%s:role/%s", c.Caller.AccountId, c.ResourceName())
}

func (c *Aws) PolicyDocument() (string, error) {
	return tmpl.Template("policy document", c.Iam.PolicyTemplate, c)
}

func (c *Aws) PolicyArn() string {
	return fmt.Sprintf("arn:aws:iam::%s:policy/%s", c.Caller.AccountId, c.ResourceName())
}

func (c *Aws) BoundaryPolicyArn() string {
	if c.Iam.BoundaryPolicy == "" {
		return c.Iam.BoundaryPolicy
	}

	if strings.HasPrefix(c.Iam.BoundaryPolicy, "arn:aws:iam::") {
		return c.Iam.BoundaryPolicy
	}

	return fmt.Sprintf("arn:aws:iam::%s:policy/%s", c.Caller.AccountId, c.Iam.BoundaryPolicy)
}

func (c *Aws) EniRoleName() string {
	return "AWSLambdaVPCAccessExecutionRole"
}

func (c *Aws) EniRoleArn() string {
	return fmt.Sprintf("arn:aws:iam::%s:role/%s", c.Caller.AccountId, c.EniRoleName())
}

func (c *Aws) EniRolePolicyArn() string {
	return "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
}

func (c *Aws) RouteKey() string {
	return fmt.Sprintf("ANY /%s/{proxy+}", c.ResourcePath())
}

func (c *Aws) ForwardedForPrefix() string {
	forwardedForPrefix := strings.Split(c.RouteKey(), " ")[1]
	forwardedForPrefix = strings.Replace(forwardedForPrefix, "/{proxy+}", "", 1)
	return forwardedForPrefix
}

func (c *Aws) ApiGwPermissionStatementId(apiId string) string {
	return strings.Join([]string{"apigatewayv2", c.ResourceName(), apiId}, "-")
}

func (c *Aws) ApiGwPermissionSourceArn() string {
	routeWithoutVerb := strings.Split(c.RouteKey(), " ")[1]
	permissionSourceArn := fmt.Sprintf("arn:aws:execute-api:%s:%s:%s/*/*%s",
		c.ApiGateway.Region, c.Caller.AccountId, c.ApiGateway.Id, routeWithoutVerb)

	return permissionSourceArn
}

func (c *Aws) RuleDocument() (string, error) {
	if c.EventBridge.RuleTemplate == "" {
		return "", nil
	}

	content, err := tmpl.Template("rule document", c.EventBridge.RuleTemplate, c)
	if err != nil {
		return "", err
	}

	return content, nil
}

func (c *Aws) EventBridgeStatementId() string {
	return strings.Join([]string{"eventbridge", c.EventBridge.BusName, c.ResourceName()}, "-")
}

func (c *Aws) Tags() map[string]string {
	return map[string]string{
		"Owner":      c.Git.Owner,
		"Repository": c.Git.Repository,
		"Service":    c.Git.Service,
		"Branch":     c.Git.Branch,
		"Sha":        c.Git.Sha,
	}
}
