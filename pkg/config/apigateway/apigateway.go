package apigateway

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/rs/zerolog/log"
)

type Basis interface {
	AwsConfig() aws.Config
	Region() string
	Name() string
	AccountId() string
}

//
// Convention
//

type Config struct {
	client       *apigatewayv2.Client
	basis        Basis
	api          string
	region       string
	route        []string
	auth         []string
	apiId        string
	authType     []string
	authorizerId []string
}

//
// Derive
//

func Derive(ctx context.Context, basis Basis) (*Config, error) {
	var err error
	var cfg Config

	cfg.basis = basis
	cfg.client = apigatewayv2.NewFromConfig(basis.AwsConfig())

	if cfg.region == "" {
		cfg.region = basis.Region()
	}

	if len(cfg.route) == 0 {
		standard := "ANY /{{.Git.Repo}}/{{.Git.Branch}}/{{.Monad.Service}}/{proxy+}"
		cfg.route = append(cfg.route, standard)
	}

	if len(cfg.auth) == 0 {
		standard := "aws_iam"
		cfg.auth = append(cfg.auth, standard)
	}

	if cfg.api != "" {
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
		v.Field(&c.client, v.Required),
		v.Field(&c.route, v.Required, v.Each(v.By(onlyGreedyProxies)), v.Length(len(c.auth), len(c.auth))),
		v.Field(&c.region, v.Required),
		v.Field(&c.auth, v.Required),
		v.Field(&c.authType, v.Each(v.In("NONE", "AWS_IAM", "CUSTOM", "JWT"))),
		v.Field(&c.authType, v.Length(len(c.authorizerId), len(c.authorizerId))),
		v.Field(&c.authorizerId, v.Length(len(c.authType), len(c.authType))),
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
func (c *Config) Api() string { return c.apiId }

// Route returns the list of route patterns
func (c *Config) Route() []string { return c.route }

// Auth returns the list of authorization configurations
func (c *Config) Auth() []string { return c.auth }

// AuthType returns the resolved authorization types
func (c *Config) AuthType() []string { return c.authType }

// AuthorizerId returns the resolved authorizer IDs
func (c *Config) AuthorizerId() []string { return c.authorizerId }

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
	return strings.Join([]string{"apigatewayv2", c.basis.Name(), apiId}, "-")
}

// PermissionSourceArns returns API Gateway execution ARNs for Lambda permissions
func (c *Config) PermissionSourceArns() ([]string, error) {
	var sourceArns []string
	for _, routeKey := range c.Route() {
		routeWithoutVerb := strings.Split(routeKey, " ")[1]
		permissionSourceArn := fmt.Sprintf("arn:aws:execute-api:%s:%s:%s/*/*%s",
			c.region, c.basis.AccountId(), c.Api(), routeWithoutVerb)
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

		if *api.Name == c.api || *api.ApiId == c.api {
			c.apiId = *api.ApiId
			return nil
		}
	}

	log.Error().
		Str("given", c.api).
		Strs("valid_ids", apiIds).
		Strs("valid_names", apiNames).
		Msg("API not found")

	return fmt.Errorf("API %s not found", c.api)
}

// resolveAuth resolves authorizer names to IDs and types
func (c *Config) resolveAuth(ctx context.Context) error {
	foundNames := []string{"none", "aws_iam"}
	foundIds := []string{"", ""}
	foundTypes := []string{"NONE", "AWS_IAM"}

	if c.apiId == "" {
		return fmt.Errorf("API Gateway ID not resolved, cannot resolve authorizers")
	}

	index, err := c.client.GetAuthorizers(ctx, &apigatewayv2.GetAuthorizersInput{
		ApiId: aws.String(c.apiId),
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

	for _, given := range c.auth {
		if slices.Contains(foundIds, given) {
			index := slices.Index(foundIds, given)
			c.authorizerId = append(c.authorizerId, foundIds[index])
			c.authType = append(c.authType, foundTypes[index])
			continue
		}

		if slices.Contains(foundNames, strings.ToLower(given)) {
			index := slices.Index(foundNames, strings.ToLower(given))
			c.authorizerId = append(c.authorizerId, foundIds[index])
			c.authType = append(c.authType, foundTypes[index])
			continue
		}

		log.Error().
			Str("api", c.apiId).
			Strs("valid_ids", foundIds).
			Strs("valid_names", foundNames).
			Msg("resolve auth")

		return fmt.Errorf("authorizer %s not found for API %s", given, c.apiId)
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
