package param

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

type ApiGatewayConfig struct {
	Client       *apigatewayv2.Client `arg:"-" json:"-"`
	Api          string               `arg:"--api,env:MONAD_API" placeholder:"id|name" help:"api gateway" default:"no-gateway"`
	Region       string               `arg:"--api-region,env:MONAD_API_REGION" placeholder:"name" help:"api gateway region" default:"caller-region"`
	Route        []string             `arg:"--route,env:MONAD_ROUTE,separate" placeholder:"pattern" help:"ANY /echo/{proxy+} | GET /api/{proxy+} (multiple flags supported) [default: [ANY /{{.Git.Repo}}/{{.Git.Branch}}/{{.Monad.Service}}/{proxy+}]"`
	Auth         []string             `arg:"--auth,env:MONAD_AUTH,separate" placeholder:"id|name" help:"none | aws_iam | jwt-auth (multiple flags supported) [default: [aws_iam]]"`
	ApiId        string               `arg:"-" json:"-"` // computed value
	AuthType     []string             `arg:"-" json:"-"` // computed value
	AuthorizerId []string             `arg:"-" json:"-"` // computed value
}

func (a *ApiGatewayConfig) Process(ctx context.Context, awsconfig aws.Config) error {
	a.Client = apigatewayv2.NewFromConfig(awsconfig)

	// Set default values if not provided
	if a.Region == "" {
		a.Region = awsconfig.Region
	}

	if len(a.Route) == 0 {
		standard := "ANY /{{.Git.Repo}}/{{.Git.Branch}}/{{.Monad.Service}}/{proxy+}"
		a.Route = append(a.Route, standard)
	}

	if len(a.Auth) == 0 {
		standard := "aws_iam"
		a.Auth = append(a.Auth, standard)
	}

	// Resolve api and authorizer names/ids given an api name or id
	if a.Api != "" {
		if err := resolve(ctx, a); err != nil {
			return err
		}
	}

	return a.Validate()
}

func (a *ApiGatewayConfig) Validate() error {
	return v.ValidateStruct(a,
		v.Field(&a.Client, v.Required),
		v.Field(&a.Route, v.Required, v.Each(v.By(onlyGreedyProxies)), v.Length(len(a.Auth), len(a.Auth))),
		v.Field(&a.Region, v.Required),
		v.Field(&a.Auth, v.Required),
		v.Field(&a.AuthType, v.Each(v.In("NONE", "AWS_IAM", "CUSTOM", "JWT"))),
		v.Field(&a.AuthType, v.Length(len(a.AuthorizerId), len(a.AuthorizerId))),
		v.Field(&a.AuthorizerId, v.Length(len(a.AuthType), len(a.AuthType))),
	)
}

func resolve(ctx context.Context, param *ApiGatewayConfig) error {
	if param == nil || param.Client == nil {
		return fmt.Errorf("api gateway client is not initialized in param struct")
	}

	if err := resolveApi(ctx, param); err != nil {
		return err
	}

	if err := resolveAuth(ctx, param); err != nil {
		return err
	}

	return nil
}

func resolveApi(ctx context.Context, param *ApiGatewayConfig) error {
	var apiIds []string
	var apiNames []string

	apis, err := param.Client.GetApis(ctx, &apigatewayv2.GetApisInput{})
	if err != nil {
		return err
	}

	for _, api := range apis.Items {
		apiIds = append(apiIds, *api.ApiId)
		apiNames = append(apiNames, *api.Name)

		if *api.Name == param.Api || *api.ApiId == param.Api {
			param.ApiId = *api.ApiId
			return nil
		}
	}

	log.Error().
		Str("given", param.Api).
		Strs("valid_ids", apiIds).
		Strs("valid_names", apiNames).
		Msg("api not found")

	return fmt.Errorf("api %s not found", param.Api)
}

func resolveAuth(ctx context.Context, param *ApiGatewayConfig) error {
	foundNames := []string{"none", "aws_iam"}
	foundIds := []string{"", ""}
	foundTypes := []string{"NONE", "AWS_IAM"}

	if param.ApiId == "" {
		return fmt.Errorf("api gateway ID is not yet resolved, cannot resolve authorizers")
	}

	index, err := param.Client.GetAuthorizers(ctx, &apigatewayv2.GetAuthorizersInput{
		ApiId: aws.String(param.ApiId),
	})

	if err != nil {
		return err
	}

	for _, found := range index.Items {
		foundNames = append(foundNames, strings.ToLower(*found.Name))
		foundIds = append(foundIds, *found.AuthorizerId)
		foundTypes = append(foundTypes, string(found.AuthorizerType))
	}

	// For development purposes, we can delete this check later, but for now we raise an error if the lengths are inconsistent
	if len(foundIds) != len(foundNames) || len(foundIds) != len(foundTypes) {
		return fmt.Errorf("inconsistent authorizer data: ids %d, names %d, types %d", len(foundIds), len(foundNames), len(foundTypes))
	}

	for _, given := range param.Auth {
		// If the given auth is an ID, populate rest accordingly
		if slices.Contains(foundIds, given) {
			index := slices.Index(foundIds, given)
			param.AuthorizerId = append(param.AuthorizerId, foundIds[index])
			param.AuthType = append(param.AuthType, foundTypes[index])
			continue
		}

		// If the given auth is a name, populate rest accordingly
		if slices.Contains(foundNames, strings.ToLower(given)) {
			index := slices.Index(foundNames, strings.ToLower(given))
			param.AuthorizerId = append(param.AuthorizerId, foundIds[index])
			param.AuthType = append(param.AuthType, foundTypes[index])
			continue
		}

		// If the given auth is neither a known ID nor a name, return an error
		log.Error().
			Str("api", param.ApiId).
			Strs("valid_ids", foundIds).
			Strs("valid_names", foundNames).
			Msg("resolve auth")

		return fmt.Errorf("authorizer %s not found for api %s", given, param.ApiId)
	}

	return nil
}

// onlyGreedyProxies validates that a route contains and ends with {proxy+} pattern
func onlyGreedyProxies(value interface{}) error {
	route, ok := value.(string)
	if !ok {
		return fmt.Errorf("route must be a string")
	}

	// Check if route contains {proxy+} pattern
	if !strings.Contains(route, "{proxy+}") {
		return fmt.Errorf("route must contain {proxy+} pattern: %s", route)
	}

	// Ensure it ends with {proxy+} (not in the middle)
	if !strings.HasSuffix(route, "{proxy+}") {
		return fmt.Errorf("route must end with {proxy+} pattern: %s", route)
	}

	return nil
}
