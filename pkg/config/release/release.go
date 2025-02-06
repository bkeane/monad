package release

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bkeane/monad/internal/git"
	"github.com/bkeane/substrate/pkg/env"
	"github.com/bkeane/substrate/pkg/substrate"

	"github.com/bkeane/monad/pkg/config/tmpl"
	"github.com/bkeane/monad/pkg/schema"
	"github.com/rs/zerolog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

type Resources struct {
	EphemeralStorage int32             `json:"ephemeralStorage"`
	MemorySize       int32             `json:"memorySize"`
	Timeout          int32             `json:"timeout"`
	Env              map[string]string `json:"env"`
}

type Image struct {
	Architecture []types.Architecture
	Digest       string
	schema       schema.Spec
}

type Config struct {
	AwsConfig aws.Config `json:"-"`
	Substrate *substrate.Substrate
	Image     Image
	log       *zerolog.Logger
}

func (r Config) Parse(ctx context.Context, awsconfig aws.Config, imageUri string, substrate *substrate.Substrate) (*Config, error) {
	var err error
	c := &Config{}
	c.log = zerolog.Ctx(ctx)
	c.Substrate = substrate

	image, err := c.Substrate.ECR.FetchByUri(ctx, imageUri)
	if err != nil {
		return nil, err
	}

	switch image.Architecture {
	case "amd64":
		c.Image.Architecture = []types.Architecture{types.ArchitectureX8664}
	case "arm64":
		c.Image.Architecture = []types.Architecture{types.ArchitectureArm64}
	default:
		return nil, fmt.Errorf("unsupported architecture: %s", image.Architecture)
	}

	// Everything is latest for now, see comment in schema.go. No blockers, just not mvp.
	c.Image.schema = schema.Version["latest"]
	if err = c.Image.schema.Decode(image.Config.Labels); err != nil {
		return nil, err
	}

	c.Image.Digest = image.Digest

	c.log.Info().
		Str("image_uri", imageUri).
		Str("branch", c.Image.schema.Git().Branch).
		Str("sha", c.Image.schema.Git().Sha).
		Str("origin", c.Image.schema.Git().Origin).
		Msg("release config")

	return c, nil
}

// Namespacing
func (o *Config) ResourceNamePrefix() string {
	return o.Image.schema.ResourceNamePrefix(o.Substrate.Options.OrgPrefixedNames)
}

func (o *Config) ResourceName() string {
	return o.Image.schema.ResourceName(o.Substrate.Options.OrgPrefixedNames)
}

func (o *Config) PossibleResourceNames() []string {
	return []string{
		o.Image.schema.ResourceName(true),
		o.Image.schema.ResourceName(false),
	}
}

func (o *Config) ResourcePathPrefix() string {
	return o.Image.schema.ResourcePathPrefix(o.Substrate.Options.OrgPrefixedPaths)
}

func (o *Config) ResourcePath() string {
	return o.Image.schema.ResourcePath(o.Substrate.Options.OrgPrefixedPaths)
}

func (o *Config) PossibleResourcePaths() []string {
	return []string{
		o.Image.schema.ResourcePath(true),
		o.Image.schema.ResourcePath(false),
	}
}

func (o *Config) ResourceTags() map[string]string {
	return o.Image.schema.ResourceTags()
}

// Ecr
func (o *Config) ImagePath() string {
	return o.Image.schema.ImagePath()
}

func (o *Config) ImageUri() string {
	return filepath.Join(o.Substrate.ECR.RegistryUrl(), o.Image.schema.ImagePath()) + "@" + o.Image.Digest
}

// IAM

// Policy
func (o *Config) PolicyName() string {
	return o.ResourceName()
}

func (o *Config) PolicyArn() string {
	return fmt.Sprintf("arn:aws:iam::%s:policy/%s", *o.Substrate.Account.Id, o.PolicyName())
}

func (o *Config) PolicyDocument() (string, error) {
	return o.Image.schema.PolicyDocument(tmpl.Init(o))
}

// Role
func (o *Config) RoleName() string {
	return o.ResourceName()
}

func (o *Config) RoleArn() string {
	return fmt.Sprintf("arn:aws:iam::%s:role/%s", *o.Substrate.Account.Id, o.RoleName())
}

func (o *Config) RoleDocument() (string, error) {
	return o.Image.schema.RoleDocument(tmpl.Init(o))
}

// EniRole
func (o *Config) EniRoleName() string {
	return "AWSLambdaVPCAccessExecutionRole"
}

func (o *Config) EniRoleArn() string {
	return fmt.Sprintf("arn:aws:iam::%s:role/%s", *o.Substrate.Account.Id, o.EniRoleName())
}

func (o *Config) EniRolePolicyArn() string {
	return "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
}

// Resources
func (o *Config) Resources() (Resources, error) {
	given, err := o.Image.schema.ResourceDocument(tmpl.Init(o))
	if err != nil {
		return Resources{}, err
	}

	var resources Resources
	resources.Timeout = 3
	resources.MemorySize = 128
	resources.EphemeralStorage = 512
	resources.Env = map[string]string{
		"GIT_BRANCH": o.Image.schema.Git().Branch,
		"GIT_SHA":    o.Image.schema.Git().Sha,
		"GIT_ORIGIN": o.Image.schema.Git().Origin,
	}

	if err = json.Unmarshal([]byte(given), &resources); err != nil {
		return Resources{}, err
	}

	return resources, nil
}

// Lambda
func (o *Config) FunctionName() string {
	return o.ResourceName()
}

func (o *Config) FunctionArn() string {
	return fmt.Sprintf("arn:aws:lambda:%s:%s:function:%s", *o.Substrate.Lambda.Region, *o.Substrate.Account.Id, o.FunctionName())
}

func (o *Config) PossibleFunctionArns() []string {
	var arns []string
	for _, name := range o.PossibleResourceNames() {
		arns = append(arns, fmt.Sprintf("arn:aws:lambda:%s:%s:function:%s", *o.Substrate.Lambda.Region, *o.Substrate.Account.Id, name))
	}
	return arns
}

// ApiGatewayV2
func (o *Config) RouteKey() string {
	return fmt.Sprintf("ANY /%s/{proxy+}", o.ResourcePath())
}

func (o *Config) PossibleRouteKeys() []string {
	return []string{
		fmt.Sprintf("ANY /%s/{proxy+}", o.ResourcePath()),
		fmt.Sprintf("ANY /%s/{proxy+}", o.ResourcePath()),
	}
}

func (o *Config) ForwardedForPrefix() string {
	forwardedForPrefix := strings.Split(o.RouteKey(), " ")[1]
	forwardedForPrefix = strings.Replace(forwardedForPrefix, "/{proxy+}", "", 1)
	return forwardedForPrefix
}

func (o *Config) ApiGwPermissionStatementId(apiId string) string {
	return strings.Join([]string{"apigatewayv2", o.ResourceName(), apiId}, "-")
}

func (o *Config) ApiGwPermissionSourceArn() string {
	routeWithoutVerb := strings.Split(o.RouteKey(), " ")[1]
	permissionSourceArn := fmt.Sprintf("arn:aws:execute-api:%s:%s:%s/*/*%s",
		*o.Substrate.ApiGateway.Region, *o.Substrate.Account.Id, *o.Substrate.ApiGateway.Id, routeWithoutVerb)

	return permissionSourceArn
}

// EventBridge
type EventBridgeRule struct {
	BusName  string
	RuleName string
	Document string
}

func (o *Config) EventBridgeRules(busName string) (map[string]map[string]EventBridgeRule, error) {
	decoded, err := o.Image.schema.EventBridgeDocuments(tmpl.Init(o))
	if err != nil {
		return nil, err
	}

	rules := make(map[string]map[string]EventBridgeRule)
	rules[busName] = make(map[string]EventBridgeRule)

	for ruleName, ruleContent := range decoded {
		ruleName := o.eventBridgeRuleName(ruleName)
		rules[busName][ruleName] = EventBridgeRule{
			BusName:  busName,
			RuleName: ruleName,
			Document: ruleContent,
		}
	}

	return rules, nil
}

func (o *Config) eventBridgeRuleName(ruleName string) string {
	return strings.Join([]string{o.ResourceName(), ruleName}, "-")
}

func (o *Config) EventBridgeStatementId(busName string, ruleName string) string {
	return strings.Join([]string{"eventbridge", busName, ruleName}, "-")
}

func (o *Config) TemplateData() (env.Account, env.ECR, env.EventBridge, env.Lambda, env.ApiGateway, git.Git) {
	return o.Substrate.Account, o.Substrate.ECR, o.Substrate.EventBridge, o.Substrate.Lambda, o.Substrate.ApiGateway, o.Image.schema.Git()
}
