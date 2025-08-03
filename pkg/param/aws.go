package param

import (
	"context"
	"fmt"
	"strings"

	"github.com/bkeane/monad/internal/registry"
	"github.com/bkeane/monad/internal/tmpl"
	"github.com/rs/zerolog/log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	dotenvlib "github.com/joho/godotenv"
)

type StateMetadata struct {
	Service string `json:",omitempty"`
	Repo    string `json:",omitempty"`
	Sha     string `json:",omitempty"`
	Branch  string `json:",omitempty"`
	Owner   string `json:",omitempty"`
	Image   string `json:",omitempty"`
	Api     string `json:",omitempty"`
	Bus     string `json:",omitempty"`
}

type Aws struct {
	GitConfig         `arg:"-" json:"-"`
	ServiceConfig     `arg:"-" json:"-"`
	tmpl.TemplateData `arg:"-" json:"-"`
	CallerConfig
	RegistryConfig
	LambdaConfig
	IamConfig
	VpcConfig
	CloudWatchConfig
	ApiGatewayConfig
	EventBridgeConfig
}

func (c *Aws) Validate(ctx context.Context, awsconfig aws.Config, git GitConfig, service ServiceConfig) error {
	// Ensure that provided Git and Service configurations have already been processed.
	if err := git.Validate(); err != nil {
		return fmt.Errorf("git config not properly processed: %w", err)
	}
	if err := service.Validate(); err != nil {
		return fmt.Errorf("service config not properly processed: %w", err)
	}

	c.GitConfig = git
	c.ServiceConfig = service

	if err := c.CallerConfig.Process(ctx, awsconfig); err != nil {
		return err
	}

	if err := c.RegistryConfig.Process(ctx, awsconfig); err != nil {
		return err
	}

	if err := c.LambdaConfig.Process(ctx, awsconfig); err != nil {
		return err
	}

	if err := c.IamConfig.Process(ctx, awsconfig); err != nil {
		return err
	}

	if err := c.VpcConfig.Process(ctx, awsconfig); err != nil {
		return err
	}

	if err := c.CloudWatchConfig.Process(ctx, awsconfig); err != nil {
		return err
	}

	if err := c.ApiGatewayConfig.Process(ctx, awsconfig); err != nil {
		return err
	}

	if err := c.EventBridgeConfig.Process(ctx, awsconfig); err != nil {
		return err
	}

	// Map data into templating struct
	c.TemplateData.Monad.Service = c.Service().Name()
	c.TemplateData.Git.Sha = c.Git().Sha()
	c.TemplateData.Git.Branch = c.Git().Branch()
	c.TemplateData.Git.Owner = c.Git().Owner()
	c.TemplateData.Git.Repo = c.Git().Repository()
	c.TemplateData.Caller.AccountId = c.Caller().AccountId()
	c.TemplateData.Caller.Region = c.Caller().Region()
	c.TemplateData.Registry.Id = c.Registry().Id()
	c.TemplateData.Registry.Region = c.Registry().Region()
	c.TemplateData.Resource.Name.Prefix = c.Schema().NamePrefix()
	c.TemplateData.Resource.Name.Full = c.Schema().Name()
	c.TemplateData.Resource.Path.Prefix = c.Schema().PathPrefix()
	c.TemplateData.Resource.Path.Full = c.Schema().Path()
	c.TemplateData.Lambda.Region = c.Lambda().Region()
	c.TemplateData.Lambda.FunctionArn = c.Lambda().FunctionArn()
	c.TemplateData.Lambda.PolicyArn = c.IAM().PolicyArn()
	c.TemplateData.Lambda.RoleArn = c.IAM().RoleArn()
	c.TemplateData.Cloudwatch.Region = c.CloudWatch().Region()
	c.TemplateData.Cloudwatch.LogGroupArn = c.CloudWatch().LogGroupArn()
	c.TemplateData.ApiGateway.Region = c.ApiGateway().Region()
	c.TemplateData.ApiGateway.Id = c.ApiGateway().ID()
	c.TemplateData.EventBridge.Region = c.EventBridge().Region()
	c.TemplateData.EventBridge.RuleName = c.Schema().Name()
	c.TemplateData.EventBridge.BusName = c.EventBridge().BusName()

	log.Info().
		Str("region", awsconfig.Region).
		Str("account", c.Caller().AccountId()).
		Msgf("aws")

	return nil
}

// Receiver Pattern Namespaces

type GitResources struct{ *Aws }
type ServiceResources struct{ *Aws }
type IAMResources struct{ *Aws }
type LambdaResources struct{ *Aws }
type ApiGatewayResources struct{ *Aws }
type EventBridgeResources struct{ *Aws }
type CloudWatchResources struct{ *Aws }
type VpcResources struct{ *Aws }
type Schema struct{ *Aws }
type CallerResources struct{ *Aws }
type RegistryResources struct{ *Aws }

// Resource namespace factory methods

// Git returns the Git resource methods and configuration accessors
func (c *Aws) Git() GitResources {
	if err := c.GitConfig.Validate(); err != nil {
		log.Warn().Err(err).Msg("Git configuration validation failed")
	}
	return GitResources{c}
}

// Service returns the Service resource methods and configuration accessors
func (c *Aws) Service() ServiceResources {
	if err := c.ServiceConfig.Validate(); err != nil {
		log.Warn().Err(err).Msg("Service configuration validation failed")
	}
	return ServiceResources{c}
}

// IAM returns the IAM resource methods and configuration accessors
func (c *Aws) IAM() IAMResources {
	if err := c.IamConfig.Validate(); err != nil {
		log.Warn().Err(err).Msg("IAM configuration validation failed")
	}
	return IAMResources{c}
}

// Lambda returns the Lambda resource methods and configuration accessors
func (c *Aws) Lambda() LambdaResources {
	if err := c.LambdaConfig.Validate(); err != nil {
		log.Warn().Err(err).Msg("Lambda configuration validation failed")
	}
	return LambdaResources{c}
}

// ApiGateway returns the API Gateway resource methods and configuration accessors
func (c *Aws) ApiGateway() ApiGatewayResources {
	if err := c.ApiGatewayConfig.Validate(); err != nil {
		log.Warn().Err(err).Msg("API Gateway configuration validation failed")
	}
	return ApiGatewayResources{c}
}

// EventBridge returns the EventBridge resource methods and configuration accessors
func (c *Aws) EventBridge() EventBridgeResources {
	if err := c.EventBridgeConfig.Validate(); err != nil {
		log.Warn().Err(err).Msg("EventBridge configuration validation failed")
	}
	return EventBridgeResources{c}
}

// CloudWatch returns the CloudWatch resource methods and configuration accessors
func (c *Aws) CloudWatch() CloudWatchResources {
	if err := c.CloudWatchConfig.Validate(); err != nil {
		log.Warn().Err(err).Msg("CloudWatch configuration validation failed")
	}
	return CloudWatchResources{c}
}

// Vpc returns the VPC resource methods and configuration accessors
func (c *Aws) Vpc() VpcResources {
	if err := c.VpcConfig.Validate(); err != nil {
		log.Warn().Err(err).Msg("VPC configuration validation failed")
	}
	return VpcResources{c}
}

// Schema returns the cross-service resource organization and naming methods
func (c *Aws) Schema() Schema { return Schema{c} }

// Caller returns the AWS caller identity information
func (c *Aws) Caller() CallerResources {
	if err := c.CallerConfig.Validate(); err != nil {
		log.Warn().Err(err).Msg("Caller configuration validation failed")
	}
	return CallerResources{c}
}

// Registry returns the ECR registry configuration
func (c *Aws) Registry() RegistryResources {
	if err := c.RegistryConfig.Validate(); err != nil {
		log.Warn().Err(err).Msg("Registry configuration validation failed")
	}
	return RegistryResources{c}
}

// IAMResources methods

// RoleDocument returns the processed IAM role trust policy document from template
func (i IAMResources) RoleDocument() (string, error) {
	return tmpl.Template("role document", i.IamConfig.RoleTemplate, i.TemplateData)
}

// RoleArn returns the complete ARN for the IAM role
func (i IAMResources) RoleArn() string {
	return fmt.Sprintf("arn:aws:iam::%s:role/%s", i.Caller().AccountId(), i.Schema().Name())
}

// PolicyDocument returns the processed IAM policy document from template
func (i IAMResources) PolicyDocument() (string, error) {
	return tmpl.Template("policy document", i.IamConfig.PolicyTemplate, i.TemplateData)
}

// PolicyArn returns the complete ARN for the IAM policy
func (i IAMResources) PolicyArn() string {
	return fmt.Sprintf("arn:aws:iam::%s:policy/%s", i.Caller().AccountId(), i.Schema().Name())
}

// BoundaryPolicy returns the boundary policy name (extracted from ARN if provided)
func (i IAMResources) BoundaryPolicy() string {
	if strings.HasPrefix(i.IamConfig.BoundaryPolicy, "arn:aws:iam::") {
		return strings.Split(i.IamConfig.BoundaryPolicy, ":policy/")[1]
	}
	return i.IamConfig.BoundaryPolicy
}

// BoundaryPolicyArn returns the complete ARN for the boundary policy
func (i IAMResources) BoundaryPolicyArn() string {
	if i.IamConfig.BoundaryPolicy == "" {
		return i.IamConfig.BoundaryPolicy
	}
	if strings.HasPrefix(i.IamConfig.BoundaryPolicy, "arn:aws:iam::") {
		return i.IamConfig.BoundaryPolicy
	}
	return fmt.Sprintf("arn:aws:iam::%s:policy/%s", i.Caller().AccountId(), i.IamConfig.BoundaryPolicy)
}

// EniRoleName returns the standard AWS Lambda VPC execution role name
func (i IAMResources) EniRoleName() string {
	return "AWSLambdaVPCAccessExecutionRole"
}

// EniRoleArn returns the complete ARN for the Lambda VPC execution role
func (i IAMResources) EniRoleArn() string {
	return fmt.Sprintf("arn:aws:iam::%s:role/%s", i.Caller().AccountId(), i.EniRoleName())
}

// EniRolePolicyArn returns the AWS managed policy ARN for Lambda VPC access
func (i IAMResources) EniRolePolicyArn() string {
	return "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
}

// IAM configuration accessors

// Client returns the AWS IAM service client
func (i IAMResources) Client() *iam.Client { return i.IamConfig.Client }

// LambdaResources methods

// FunctionArn returns the complete ARN for the Lambda function
func (l LambdaResources) FunctionArn() string {
	return fmt.Sprintf("arn:aws:lambda:%s:%s:function:%s",
		l.Lambda().Region(), l.Caller().AccountId(), l.Schema().Name())
}

// Environment returns processed environment variables for the Lambda function
func (l LambdaResources) Environment() (map[string]string, error) {
	return l.Schema().Environment()
}

// Lambda configuration accessors

// Client returns the AWS Lambda service client
func (l LambdaResources) Client() *lambda.Client { return l.LambdaConfig.Client }

// Region returns the AWS region for Lambda deployment
func (l LambdaResources) Region() string { return l.LambdaConfig.Region }

// Timeout returns the function timeout in seconds
func (l LambdaResources) Timeout() int32 { return l.LambdaConfig.Timeout }

// MemorySize returns the allocated memory in MB
func (l LambdaResources) MemorySize() int32 { return l.LambdaConfig.MemorySize }

// EphemeralStorage returns the ephemeral storage size in MB
func (l LambdaResources) EphemeralStorage() int32 { return l.LambdaConfig.EphemeralStorage }

// EnvTemplate returns the environment template string or file path
func (l LambdaResources) EnvTemplate() string { return l.LambdaConfig.EnvTemplate }

// Retries returns the number of async invoke retries
func (l LambdaResources) Retries() int32 { return l.LambdaConfig.Retries }

// ApiGatewayResources methods

// RouteKeys returns all processed API Gateway route patterns
func (a ApiGatewayResources) RouteKeys() ([]string, error) {
	var routeKeys []string
	for _, route := range a.ApiGatewayConfig.Route {
		routeKey, err := tmpl.Template("route", route, a.TemplateData)
		if err != nil {
			return nil, err
		}
		routeKeys = append(routeKeys, routeKey)
	}
	return routeKeys, nil
}

// ForwardedPrefixes returns the path prefixes extracted from all route patterns
// Since all routes are required to end with {proxy+}, we can simply strip that suffix
func (a ApiGatewayResources) ForwardedPrefixes() ([]string, error) {
	routeKeys, err := a.RouteKeys()
	if err != nil {
		return nil, err
	}

	var prefixes []string
	for _, routeKey := range routeKeys {
		path := strings.Split(routeKey, " ")[1] // Extract path from "METHOD /path"
		prefix := strings.TrimSuffix(path, "/{proxy+}")
		prefixes = append(prefixes, prefix)
	}
	return prefixes, nil
}

// PermissionStatementId returns the Lambda permission statement ID for API Gateway
func (a ApiGatewayResources) PermissionStatementId(apiId string) string {
	return strings.Join([]string{"apigatewayv2", a.Schema().Name(), apiId}, "-")
}

// PermissionSourceArns returns the API Gateway execution ARNs for Lambda permissions for all routes
func (a ApiGatewayResources) PermissionSourceArns() ([]string, error) {
	routeKeys, err := a.RouteKeys()
	if err != nil {
		return nil, err
	}

	var sourceArns []string
	for _, routeKey := range routeKeys {
		routeWithoutVerb := strings.Split(routeKey, " ")[1]
		permissionSourceArn := fmt.Sprintf("arn:aws:execute-api:%s:%s:%s/*/*%s",
			a.ApiGateway().Region(), a.Caller().AccountId(), a.ApiGateway().ID(), routeWithoutVerb)
		sourceArns = append(sourceArns, permissionSourceArn)
	}
	return sourceArns, nil
}

// Legacy methods for backward compatibility - use first route only

// RouteKey returns the processed API Gateway route pattern for the first route (legacy method)
func (a ApiGatewayResources) RouteKey() (string, error) {
	routeKeys, err := a.RouteKeys()
	if err != nil {
		return "", err
	}
	if len(routeKeys) == 0 {
		return "", fmt.Errorf("no routes configured")
	}
	return routeKeys[0], nil
}

// ForwardedForPrefix returns the path prefix for the first route (legacy method)
func (a ApiGatewayResources) ForwardedForPrefix() (string, error) {
	prefixes, err := a.ForwardedPrefixes()
	if err != nil {
		return "", err
	}
	if len(prefixes) == 0 {
		return "", fmt.Errorf("no routes configured")
	}
	return prefixes[0], nil
}

// PermissionSourceArn returns the API Gateway execution ARN for the first route (legacy method)
func (a ApiGatewayResources) PermissionSourceArn() (string, error) {
	sourceArns, err := a.PermissionSourceArns()
	if err != nil {
		return "", err
	}
	if len(sourceArns) == 0 {
		return "", fmt.Errorf("no routes configured")
	}
	return sourceArns[0], nil
}

// ApiGateway configuration accessors

// Client returns the AWS API Gateway service client
func (a ApiGatewayResources) Client() *apigatewayv2.Client { return a.ApiGatewayConfig.Client }

// Region returns the AWS region for API Gateway deployment
func (a ApiGatewayResources) Region() string { return a.ApiGatewayConfig.Region }

// Api returns the API Gateway ID or name
func (a ApiGatewayResources) Api() string { return a.ApiGatewayConfig.Api }

// Route returns the list of route patterns
func (a ApiGatewayResources) Route() []string { return a.ApiGatewayConfig.Route }

// Auth returns the list of authorization configurations
func (a ApiGatewayResources) Auth() []string { return a.ApiGatewayConfig.Auth }

// ID returns the resolved API Gateway ID
func (a ApiGatewayResources) ID() string { return a.ApiGatewayConfig.ApiId }

// AuthType returns the resolved authorization types
func (a ApiGatewayResources) AuthType() []string { return a.ApiGatewayConfig.AuthType }

// AuthorizerId returns the resolved authorizer IDs
func (a ApiGatewayResources) AuthorizerId() []string { return a.ApiGatewayConfig.AuthorizerId }

// EventBridgeResources methods

// RuleDocument returns the processed EventBridge rule document from template
func (e EventBridgeResources) RuleDocument() (string, error) {
	if e.EventBridgeConfig.RuleTemplate == "" {
		return "", nil
	}

	content, err := tmpl.Template("rule document", e.EventBridgeConfig.RuleTemplate, e.TemplateData)
	if err != nil {
		return "", err
	}

	return content, nil
}

// PermissionStatementId returns the Lambda permission statement ID for EventBridge
func (e EventBridgeResources) PermissionStatementId() string {
	return strings.Join([]string{"eventbridge", e.EventBridgeConfig.BusName, e.Schema().Name()}, "-")
}

// EventBridge configuration accessors

// Client returns the AWS EventBridge service client
func (e EventBridgeResources) Client() *eventbridge.Client { return e.EventBridgeConfig.Client }

// Region returns the AWS region for EventBridge deployment
func (e EventBridgeResources) Region() string { return e.EventBridgeConfig.Region }

// BusName returns the EventBridge custom bus name
func (e EventBridgeResources) BusName() string { return e.EventBridgeConfig.BusName }

// RuleTemplate returns the EventBridge rule template string or file path
func (e EventBridgeResources) RuleTemplate() string { return e.EventBridgeConfig.RuleTemplate }

// CloudWatchResources methods

// LogGroup returns the CloudWatch log group name for the Lambda function
func (c CloudWatchResources) LogGroup() string {
	return fmt.Sprintf("/aws/lambda/%s", c.Schema().Path())
}

// LogGroupArn returns the complete ARN for the CloudWatch log group
func (c CloudWatchResources) LogGroupArn() string {
	return fmt.Sprintf("arn:aws:logs:%s:%s:log-group:%s", c.CloudWatch().Region(), c.Caller().AccountId(), c.LogGroup())
}

// LogRetention returns the log retention period in days
func (c CloudWatchResources) LogRetention() int32 {
	return c.CloudWatchConfig.Retention
}

// CloudWatch configuration accessors

// Client returns the AWS CloudWatch service client
func (c CloudWatchResources) Client() *cloudwatchlogs.Client { return c.CloudWatchConfig.Client }

// Region returns the AWS region for CloudWatch deployment
func (c CloudWatchResources) Region() string { return c.CloudWatchConfig.Region }

// Retention returns the log retention period in days
func (c CloudWatchResources) Retention() int32 { return c.CloudWatchConfig.Retention }

// VpcResources methods

// VPC configuration accessors

// Client returns the AWS EC2 service client for VPC operations
func (v VpcResources) Client() *ec2.Client { return v.VpcConfig.Client }

// SecurityGroupIds returns the resolved security group IDs
func (v VpcResources) SecurityGroupIds() []string { return v.VpcConfig.SecurityGroupIds }

// SubnetIds returns the resolved subnet IDs
func (v VpcResources) SubnetIds() []string { return v.VpcConfig.SubnetIds }

// SecurityGroups returns the security group names/IDs from configuration
func (v VpcResources) SecurityGroups() []string { return v.VpcConfig.SecurityGroups }

// Subnets returns the subnet names/IDs from configuration
func (v VpcResources) Subnets() []string { return v.VpcConfig.Subnets }

// Schema methods - cross-service resource organization and naming

// NamePrefix returns the resource name prefix: {repo}-{branch}
func (s Schema) NamePrefix() string {
	return fmt.Sprintf("%s-%s", s.GitConfig.Repository, s.GitConfig.Branch)
}

// Name returns the full resource name: {repo}-{branch}-{service}
func (s Schema) Name() string {
	return fmt.Sprintf("%s-%s", s.NamePrefix(), s.ServiceConfig.Name)
}

// PathPrefix returns the resource path prefix: {repo}/{branch}
func (s Schema) PathPrefix() string {
	return fmt.Sprintf("%s/%s", s.GitConfig.Repository, s.GitConfig.Branch)
}

// Path returns the full resource path: {repo}/{branch}/{service}
func (s Schema) Path() string {
	return fmt.Sprintf("%s/%s", s.PathPrefix(), s.ServiceConfig.Name)
}

// ImagePath returns the full image path: {repo}/{branch}/{service}
func (s Schema) ImagePath() string {
	return fmt.Sprintf("%s/%s/%s", s.GitConfig.Repository, s.GitConfig.Branch, s.ServiceConfig.Name)
}

// Environment returns processed environment variables with cross-service metadata
func (s Schema) Environment() (map[string]string, error) {
	env, err := tmpl.Template("env", s.Lambda().EnvTemplate(), s.TemplateData)
	if err != nil {
		return nil, err
	}

	envMap, err := dotenvlib.Parse(strings.NewReader(env))
	if err != nil {
		return nil, err
	}

	// Inject default environment variables
	// Generally useful and used for efficient state lookup
	metadata := s.StateMetadata()
	envMap["MONAD_SERVICE"] = metadata.Service
	envMap["MONAD_OWNER"] = metadata.Owner
	envMap["MONAD_REPO"] = metadata.Repo
	envMap["MONAD_SHA"] = metadata.Sha
	envMap["MONAD_BRANCH"] = metadata.Branch
	envMap["MONAD_IMAGE"] = metadata.Image

	if s.ApiGateway().ID() != "" {
		envMap["MONAD_API"] = s.ApiGateway().Api()
	}

	if s.EventBridge().RuleTemplate() != "" {
		envMap["MONAD_BUS"] = s.EventBridge().BusName()
	}

	return envMap, nil
}

// Tags returns standardized AWS resource tags for cross-service consistency
func (s Schema) Tags() map[string]string {
	metadata := s.StateMetadata()

	return map[string]string{
		"Monad":   "true",
		"Service": metadata.Service,
		"Owner":   metadata.Owner,
		"Repo":    metadata.Repo,
		"Branch":  metadata.Branch,
		"Sha":     metadata.Sha,
	}
}

// StateMetadata returns service metadata for state management
func (s Schema) StateMetadata() *StateMetadata {
	return &StateMetadata{
		Service: s.ServiceConfig.Name,
		Owner:   s.GitConfig.Owner,
		Repo:    s.GitConfig.Repository,
		Branch:  s.GitConfig.Branch,
		Sha:     s.GitConfig.Sha,
		Image:   s.ServiceConfig.Image,
	}
}

// CallerResources methods - AWS caller identity accessors

// AccountId returns the AWS account ID
func (c CallerResources) AccountId() string { return c.CallerConfig.AccountId }

// Region returns the AWS region
func (c CallerResources) Region() string { return c.CallerConfig.Region }

// Arn returns the caller ARN
func (c CallerResources) Arn() string { return c.CallerConfig.Arn }

// UserId returns the caller user ID
func (c CallerResources) UserId() string { return c.CallerConfig.UserId }

// Client returns the AWS STS client
func (c CallerResources) Client() *sts.Client { return c.CallerConfig.Client }

// RegistryResources methods - ECR registry accessors

// Id returns the ECR registry ID
func (r RegistryResources) Id() string { return r.RegistryConfig.Id }

// Region returns the ECR registry region
func (r RegistryResources) Region() string { return r.RegistryConfig.Region }

// Client returns the registry client
func (r RegistryResources) Client() *registry.Client { return r.RegistryConfig.Client }

// GitResources methods - Git configuration accessors

// Owner returns the git repository owner
func (g GitResources) Owner() string { return g.GitConfig.Owner }

// Repository returns the git repository name
func (g GitResources) Repository() string { return g.GitConfig.Repository }

// Branch returns the git repository branch
func (g GitResources) Branch() string { return g.GitConfig.Branch }

// Sha returns the git repository sha
func (g GitResources) Sha() string { return g.GitConfig.Sha }

// Chdir returns the git working directory
func (g GitResources) Chdir() string { return g.GitConfig.Chdir }

// ServiceResources methods - Service configuration accessors

// Name returns the service name
func (s ServiceResources) Name() string { return s.ServiceConfig.Name }

// Image returns the service container image
func (s ServiceResources) ImagePath() string { return s.ServiceConfig.ImagePath }

// ImageTag returns the service container image tag
func (s ServiceResources) ImageTag() string { return s.ServiceConfig.ImageTag }

// Image returns the full service image path
func (s ServiceResources) Image() string { return s.ServiceConfig.Image }
