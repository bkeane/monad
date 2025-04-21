package saga

import (
	"context"
	"errors"
	"fmt"

	"github.com/bkeane/monad/pkg/param"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/smithy-go"
	"github.com/rs/zerolog/log"
)

type ApiGatewayV2 struct {
	config param.Aws
}

type Api struct {
	ApiId string
}

type Route struct {
	ApiId   string
	RouteId string
}

type Integration struct {
	ApiId         string
	IntegrationId string
}

type Permission struct {
	ApiId       string
	FunctionArn string
	StatementId string
}

func (s ApiGatewayV2) Init(ctx context.Context, c param.Aws) *ApiGatewayV2 {
	return &ApiGatewayV2{
		config: c,
	}
}

func (s *ApiGatewayV2) Do(ctx context.Context) error {
	var action string
	var apiId string

	if s.config.ApiGateway.Id != "" {
		apiId = s.config.ApiGateway.Id
		action = "put"
	} else {
		apiId = "*"
		action = "delete"
	}

	log.Info().
		Str("id", apiId).
		Str("route", s.config.RouteKey()).
		Str("auth", s.config.ApiGateway.Auth).
		Str("action", action).
		Msg("apigatewayv2")

	return s.Ensure(ctx)
}

func (s *ApiGatewayV2) Undo(ctx context.Context) error {
	log.Info().
		Str("id", s.config.ApiGateway.Id).
		Str("route", s.config.RouteKey()).
		Str("auth", s.config.ApiGateway.Auth).
		Str("action", "delete").
		Msg("apigatewayv2")
	return s.Destroy(ctx)
}

func (s *ApiGatewayV2) Ensure(ctx context.Context) error {
	if s.config.ApiGateway.Id != "" {
		if err := s.Destroy(ctx); err != nil {
			return err
		}
		return s.Deploy(ctx)
	}

	return s.Destroy(ctx)
}

func (s *ApiGatewayV2) Deploy(ctx context.Context) error {
	api, err := s.GetApi(ctx, s.config.ApiGateway.Id)
	if err != nil {
		return err
	}

	integration, err := s.CreateIntegration(ctx, api)
	if err != nil {
		return err
	}

	_, err = s.CreateRoute(ctx, api, integration)
	if err != nil {
		return err
	}

	_, err = s.CreatePermission(ctx, api)
	if err != nil {
		return err
	}

	return nil
}

func (s *ApiGatewayV2) Destroy(ctx context.Context) error {
	apis, err := s.GetApis(ctx)
	if err != nil {
		return err
	}

	routes, err := s.GetRoutes(ctx, apis)
	if err != nil {
		return err
	}

	integrations, err := s.GetIntegrations(ctx, apis)
	if err != nil {
		return err
	}

	permissions, err := s.GetPermissions(ctx, apis)
	if err != nil {
		return err
	}

	for _, route := range routes {
		if _, err := s.DeleteRoute(ctx, route); err != nil {
			return err
		}
	}

	for _, integration := range integrations {
		if _, err := s.DeleteIntegration(ctx, integration); err != nil {
			return err
		}
	}

	for _, permission := range permissions {
		if err := s.DeletePermissions(ctx, permission); err != nil {
			return err
		}
	}

	return nil
}

// Create functions
func (s *ApiGatewayV2) CreateIntegration(ctx context.Context, api Api) (Integration, error) {
	create := &apigatewayv2.CreateIntegrationInput{
		ApiId:                aws.String(api.ApiId),
		ConnectionType:       types.ConnectionTypeInternet,
		IntegrationType:      types.IntegrationTypeAwsProxy,
		IntegrationUri:       aws.String(s.config.FunctionArn()),
		PayloadFormatVersion: aws.String("2.0"),
		RequestParameters: map[string]string{
			"overwrite:path":                      "/$request.path.proxy",
			"overwrite:header.X-Forwarded-Prefix": s.config.ForwardedForPrefix(),
		},
	}

	integration, err := s.config.ApiGateway.Client.CreateIntegration(ctx, create)
	if err != nil {
		return Integration{}, err
	}

	return Integration{
		ApiId:         api.ApiId,
		IntegrationId: *integration.IntegrationId,
	}, nil
}

func (s *ApiGatewayV2) CreateRoute(ctx context.Context, api Api, integration Integration) (Route, error) {
	authTypeMap := map[string]types.AuthorizationType{
		"NONE":    types.AuthorizationTypeNone,
		"AWS_IAM": types.AuthorizationTypeAwsIam,
		"CUSTOM":  types.AuthorizationTypeCustom,
		"JWT":     types.AuthorizationTypeJwt,
	}

	if s.config.ApiGateway.AuthType == "" {
		log.Debug().Msg("no auth type specified, defaulting to AWS_IAM")
		s.config.ApiGateway.AuthType = "AWS_IAM"
	}

	authType, ok := authTypeMap[s.config.ApiGateway.AuthType]
	if !ok {
		return Route{}, fmt.Errorf("unsupported authorization type %s", s.config.ApiGateway.AuthType)
	}

	create := &apigatewayv2.CreateRouteInput{
		ApiId:             aws.String(api.ApiId),
		RouteKey:          aws.String(s.config.RouteKey()),
		Target:            aws.String(fmt.Sprintf("integrations/%s", integration.IntegrationId)),
		AuthorizationType: authType,
		AuthorizerId:      aws.String(s.config.ApiGateway.AuthorizerId),
	}

	log.Debug().
		Str("api_id", api.ApiId).
		Str("route_key", s.config.RouteKey()).
		Str("auth_type", s.config.ApiGateway.AuthType).
		Msg("creating route")

	route, err := s.config.ApiGateway.Client.CreateRoute(ctx, create)
	if err != nil {
		return Route{}, err
	}

	return Route{
		ApiId:   api.ApiId,
		RouteId: *route.RouteId,
	}, nil
}

func (s *ApiGatewayV2) CreatePermission(ctx context.Context, api Api) (Permission, error) {
	create := &lambda.AddPermissionInput{
		FunctionName: aws.String(s.config.FunctionArn()),
		Action:       aws.String("lambda:InvokeFunction"),
		Principal:    aws.String("apigateway.amazonaws.com"),
		SourceArn:    aws.String(s.config.ApiGwPermissionSourceArn()),
		StatementId:  aws.String(s.config.ApiGwPermissionStatementId(api.ApiId)),
	}

	_, err := s.config.Lambda.Client.AddPermission(ctx, create)
	if err != nil {
		return Permission{}, err
	}

	return Permission{
		ApiId:       api.ApiId,
		FunctionArn: *create.FunctionName,
		StatementId: *create.StatementId,
	}, nil
}

// Delete functions
func (s *ApiGatewayV2) DeleteRoute(ctx context.Context, route Route) (*apigatewayv2.DeleteRouteOutput, error) {
	input := &apigatewayv2.DeleteRouteInput{
		ApiId:   aws.String(route.ApiId),
		RouteId: aws.String(route.RouteId),
	}

	log.Debug().
		Str("action", "delete").
		Str("api_id", route.ApiId).
		Str("route_key", s.config.RouteKey()).
		Msg("apigatewayv2")

	return s.config.ApiGateway.Client.DeleteRoute(ctx, input)
}

func (s *ApiGatewayV2) DeleteIntegration(ctx context.Context, integration Integration) (*apigatewayv2.DeleteIntegrationOutput, error) {
	input := &apigatewayv2.DeleteIntegrationInput{
		ApiId:         aws.String(integration.ApiId),
		IntegrationId: aws.String(integration.IntegrationId),
	}

	return s.config.ApiGateway.Client.DeleteIntegration(ctx, input)
}

func (s *ApiGatewayV2) DeletePermissions(ctx context.Context, permission Permission) error {
	var apiErr smithy.APIError

	input := &lambda.RemovePermissionInput{
		FunctionName: aws.String(permission.FunctionArn),
		StatementId:  aws.String(permission.StatementId),
	}

	_, err := s.config.Lambda.Client.RemovePermission(ctx, input)
	if err != nil {
		switch errors.As(err, &apiErr) {
		case apiErr.ErrorCode() == "ResourceNotFoundException":
			return nil
		default:
			return err
		}
	}

	return err
}

// GET functions
func (s *ApiGatewayV2) GetApis(ctx context.Context) ([]Api, error) {
	var apis []Api
	var nextToken *string

	for {
		input := &apigatewayv2.GetApisInput{
			NextToken: nextToken,
		}

		result, err := s.config.ApiGateway.Client.GetApis(ctx, input)
		if err != nil {
			return nil, err
		}

		for _, api := range result.Items {
			apis = append(apis, Api{
				ApiId: *api.ApiId,
			})
		}

		if result.NextToken == nil {
			break
		}
		nextToken = result.NextToken
	}

	return apis, nil
}

func (s *ApiGatewayV2) GetRoutes(ctx context.Context, apis []Api) ([]Route, error) {
	var routes []Route
	var nextToken *string

	for _, api := range apis {
		for {
			input := &apigatewayv2.GetRoutesInput{
				ApiId:     aws.String(api.ApiId),
				NextToken: nextToken,
			}

			result, err := s.config.ApiGateway.Client.GetRoutes(ctx, input)
			if err != nil {
				return nil, err
			}

			for _, route := range result.Items {
				if s.config.RouteKey() == *route.RouteKey {
					routes = append(routes, Route{
						ApiId:   api.ApiId,
						RouteId: *route.RouteId,
					})
				}
			}

			if result.NextToken == nil {
				break
			}
			nextToken = result.NextToken
		}
	}

	return routes, nil
}

func (s *ApiGatewayV2) GetApi(ctx context.Context, apiId string) (Api, error) {
	apis, err := s.GetApis(ctx)
	if err != nil {
		return Api{}, err
	}

	for _, api := range apis {
		if api.ApiId == apiId {
			return api, nil
		}
	}

	return Api{}, fmt.Errorf("api not found")
}

func (s *ApiGatewayV2) GetIntegrations(ctx context.Context, apis []Api) ([]Integration, error) {
	var integrations []Integration
	var nextToken *string

	for _, api := range apis {
		for {
			input := &apigatewayv2.GetIntegrationsInput{
				ApiId:     aws.String(api.ApiId),
				NextToken: nextToken,
			}

			result, err := s.config.ApiGateway.Client.GetIntegrations(ctx, input)
			if err != nil {
				return nil, err
			}

			for _, integration := range result.Items {
				if s.config.FunctionArn() == *integration.IntegrationUri {
					integrations = append(integrations, Integration{
						ApiId:         api.ApiId,
						IntegrationId: *integration.IntegrationId,
					})
				}
			}

			if result.NextToken == nil {
				break
			}
			nextToken = result.NextToken
		}
	}

	return integrations, nil
}

func (s *ApiGatewayV2) GetPermissions(ctx context.Context, apis []Api) ([]Permission, error) {
	var permissions []Permission

	for _, api := range apis {
		permissions = append(permissions, Permission{
			ApiId:       api.ApiId,
			FunctionArn: s.config.FunctionArn(),
			StatementId: s.config.ApiGwPermissionStatementId(api.ApiId),
		})
	}

	return permissions, nil
}
