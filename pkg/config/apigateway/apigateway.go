package apigateway

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/bkeane/monad/pkg/basis/caller"
	"github.com/bkeane/monad/pkg/basis/resource"
	"github.com/caarlos0/env/v11"
	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/rs/zerolog/log"
)

type Basis interface {
	Caller() (*caller.Basis, error)
	Resource() (*resource.Basis, error)
	Render(string) (string, error)
}

//
// Convention
//

type Config struct {
	client                  *apigatewayv2.Client
	ApiGateway              string   `env:"MONAD_API" flag:"--api" usage:"API Gateway ID or name" hint:"name|id"`
	ApiGatewayRegion        string   `env:"MONAD_API_REGION"`
	ApiGatewayRoutePatterns []string `env:"MONAD_ROUTE" flag:"--route" usage:"API Gateway route patterns" hint:"pattern"`
	ApiGatewayAuthTypes     []string `env:"MONAD_AUTH" flag:"--auth" usage:"API Gateway authorization types" hint:"name|id"`
	ApiGatewayId            string
	ApiGatewayAuthType      []string
	ApiGatewayAuthorizerId  []string
	caller                  *caller.Basis
	resource                *resource.Basis
	basis                   Basis
}

//
// Derive
//

func Derive(ctx context.Context, basis Basis) (*Config, error) {
	var err error
	var cfg Config

	// Parse environment variables into struct fields
	if err = env.Parse(&cfg); err != nil {
		return nil, err
	}

	cfg.basis = basis

	cfg.resource, err = basis.Resource()
	if err != nil {
		return nil, err
	}

	cfg.caller, err = basis.Caller()
	if err != nil {
		return nil, err
	}

	cfg.client = apigatewayv2.NewFromConfig(cfg.caller.AwsConfig())

	if cfg.ApiGatewayRegion == "" {
		cfg.ApiGatewayRegion = cfg.caller.AwsConfig().Region
	}

	if len(cfg.ApiGatewayRoutePatterns) == 0 {
		standard := "ANY /{{.Git.Repo}}/{{.Git.Branch}}/{{.Service.Name}}/{proxy+}"
		cfg.ApiGatewayRoutePatterns = append(cfg.ApiGatewayRoutePatterns, standard)
	}

	// Render template variables in route patterns
	for i, pattern := range cfg.ApiGatewayRoutePatterns {
		rendered, err := cfg.basis.Render(pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to render route pattern %q: %w", pattern, err)
		}
		cfg.ApiGatewayRoutePatterns[i] = rendered
	}

	if len(cfg.ApiGatewayAuthTypes) == 0 {
		standard := "aws_iam"
		cfg.ApiGatewayAuthTypes = append(cfg.ApiGatewayAuthTypes, standard)
	}

	if cfg.ApiGateway != "" {
		if err = cfg.resolve(ctx); err != nil {
			return nil, err
		}
	}

	if err = cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

//
// Validations
//

func (c *Config) Validate() error {
	return v.ValidateStruct(c,
		v.Field(&c.basis, v.Required),
		v.Field(&c.client, v.Required),
		v.Field(&c.ApiGatewayRoutePatterns, v.Required, v.Each(v.By(onlyGreedyProxies)), v.Length(len(c.ApiGatewayAuthTypes), len(c.ApiGatewayAuthTypes))),
		v.Field(&c.ApiGatewayRegion, v.Required),
		v.Field(&c.ApiGatewayAuthTypes, v.Required),
		v.Field(&c.ApiGatewayAuthType, v.Each(v.In("NONE", "AWS_IAM", "CUSTOM", "JWT"))),
		v.Field(&c.ApiGatewayAuthType, v.Length(len(c.ApiGatewayAuthorizerId), len(c.ApiGatewayAuthorizerId))),
		v.Field(&c.ApiGatewayAuthorizerId, v.Length(len(c.ApiGatewayAuthType), len(c.ApiGatewayAuthType))),
	)
}

//
// Computed
//

// Client returns the AWS API Gateway service client
func (c *Config) Client() *apigatewayv2.Client {
	return c.client
}

// Api returns the resolved API Gateway ID
func (c *Config) ApiId() string { return c.ApiGateway }

// Route returns the list of route patterns
func (c *Config) Route() []string { return c.ApiGatewayRoutePatterns }

// Auth returns the list of authorization configurations
func (c *Config) Auth() []string { return c.ApiGatewayAuthTypes }

// AuthType returns the resolved authorization types
func (c *Config) AuthType() []string { return c.ApiGatewayAuthType }

// AuthorizerId returns the resolved authorizer IDs
func (c *Config) AuthorizerId() []string { return c.ApiGatewayAuthorizerId }

// ForwardedPrefixes returns path prefixes extracted from all route patterns
func (c *Config) ForwardedPrefixes() ([]string, error) {
	var prefixes []string
	for _, routeKey := range c.Route() {
		path := strings.Split(routeKey, " ")[1] // Extract path from "METHOD /path"
		prefix := strings.TrimSuffix(path, "/{proxy+}")
		prefixes = append(prefixes, prefix)
	}
	return prefixes, nil
}

// PermissionStatementId returns Lambda permission statement ID for API Gateway
func (c *Config) PermissionStatementId(apiId string) string {
	return strings.Join([]string{"apigatewayv2", c.resource.Name(), apiId}, "-")
}

// PermissionSourceArns returns API Gateway execution ARNs for Lambda permissions
func (c *Config) PermissionSourceArns() ([]string, error) {
	var sourceArns []string
	for _, routeKey := range c.Route() {
		routeWithoutVerb := strings.Split(routeKey, " ")[1]
		permissionSourceArn := fmt.Sprintf("arn:aws:execute-api:%s:%s:%s/*/*%s",
			c.ApiGatewayRegion, c.caller.AccountId(), c.ApiId(), routeWithoutVerb)
		sourceArns = append(sourceArns, permissionSourceArn)
	}
	return sourceArns, nil
}

//
// Helpers
//

// resolve handles API and authorizer name/ID resolution
func (c *Config) resolve(ctx context.Context) error {
	if c.client == nil {
		return fmt.Errorf("API Gateway client not initialized")
	}

	if err := c.resolveApi(ctx); err != nil {
		return err
	}

	if err := c.resolveAuth(ctx); err != nil {
		return err
	}

	return nil
}

// resolveApi finds API ID from name or validates existing ID
func (c *Config) resolveApi(ctx context.Context) error {
	var apiIds []string
	var apiNames []string

	apis, err := c.client.GetApis(ctx, &apigatewayv2.GetApisInput{})
	if err != nil {
		return err
	}

	for _, api := range apis.Items {
		apiIds = append(apiIds, *api.ApiId)
		apiNames = append(apiNames, *api.Name)

		if *api.Name == c.ApiGateway || *api.ApiId == c.ApiGateway {
			c.ApiGatewayId = *api.ApiId
			return nil
		}
	}

	log.Error().
		Str("given", c.ApiGateway).
		Strs("valid_ids", apiIds).
		Strs("valid_names", apiNames).
		Msg("API not found")

	return fmt.Errorf("API %s not found", c.ApiGateway)
}

// resolveAuth resolves authorizer names to IDs and types
func (c *Config) resolveAuth(ctx context.Context) error {
	foundNames := []string{"none", "aws_iam"}
	foundIds := []string{"", ""}
	foundTypes := []string{"NONE", "AWS_IAM"}

	if c.ApiGatewayId == "" {
		return fmt.Errorf("API Gateway ID not resolved, cannot resolve authorizers")
	}

	index, err := c.client.GetAuthorizers(ctx, &apigatewayv2.GetAuthorizersInput{
		ApiId: aws.String(c.ApiGatewayId),
	})
	if err != nil {
		return err
	}

	for _, found := range index.Items {
		foundNames = append(foundNames, strings.ToLower(*found.Name))
		foundIds = append(foundIds, *found.AuthorizerId)
		foundTypes = append(foundTypes, string(found.AuthorizerType))
	}

	if len(foundIds) != len(foundNames) || len(foundIds) != len(foundTypes) {
		return fmt.Errorf("inconsistent authorizer data: ids %d, names %d, types %d",
			len(foundIds), len(foundNames), len(foundTypes))
	}

	for _, given := range c.ApiGatewayAuthTypes {
		if slices.Contains(foundIds, given) {
			index := slices.Index(foundIds, given)
			c.ApiGatewayAuthorizerId = append(c.ApiGatewayAuthorizerId, foundIds[index])
			c.ApiGatewayAuthType = append(c.ApiGatewayAuthType, foundTypes[index])
			continue
		}

		if slices.Contains(foundNames, strings.ToLower(given)) {
			index := slices.Index(foundNames, strings.ToLower(given))
			c.ApiGatewayAuthorizerId = append(c.ApiGatewayAuthorizerId, foundIds[index])
			c.ApiGatewayAuthType = append(c.ApiGatewayAuthType, foundTypes[index])
			continue
		}

		log.Error().
			Str("api", c.ApiGatewayId).
			Strs("valid_ids", foundIds).
			Strs("valid_names", foundNames).
			Msg("resolve auth")

		return fmt.Errorf("authorizer %s not found for API %s", given, c.ApiGatewayId)
	}

	return nil
}

// onlyGreedyProxies validates that a route contains and ends with {proxy+} pattern
func onlyGreedyProxies(value interface{}) error {
	route, ok := value.(string)
	if !ok {
		return fmt.Errorf("route must be a string")
	}

	if !strings.Contains(route, "{proxy+}") {
		return fmt.Errorf("route must contain {proxy+} pattern: %s", route)
	}

	if !strings.HasSuffix(route, "{proxy+}") {
		return fmt.Errorf("route must end with {proxy+} pattern: %s", route)
	}

	return nil
}
