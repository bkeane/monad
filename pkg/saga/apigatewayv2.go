package saga

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/bkeane/monad/pkg/param"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
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
	ApiId             string
	RouteId           string
	RouteKey          string
	AuthorizationType string
	AuthorizerId      string
}

type Integration struct {
	ApiId         string
	IntegrationId string
}

type Permission struct {
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

	if s.config.ApiGateway().ID() != "" {
		apiId = s.config.ApiGateway().ID()
		action = "put"
	} else {
		apiId = "*"
		action = "delete"
	}

	routeKeys, err := s.config.ApiGateway().RouteKeys()
	if err != nil {
		return err
	}

	// Log each route/auth pair individually to show 1:1 relationship
	auth := s.config.ApiGateway().Auth()
	for i, route := range routeKeys {
		log.Info().
			Str("id", apiId).
			Str("route", route).
			Str("auth", auth[i]).
			Str("action", action).
			Msg("apigatewayv2")
	}

	return s.Ensure(ctx)
}

func (s *ApiGatewayV2) Undo(ctx context.Context) error {
	return s.Destroy(ctx)
}

func (s *ApiGatewayV2) Ensure(ctx context.Context) error {
	if s.config.ApiGateway().ID() != "" {
		if err := s.Destroy(ctx); err != nil {
			return err
		}
		return s.Deploy(ctx)
	}

	return s.Destroy(ctx)
}

func (s *ApiGatewayV2) Deploy(ctx context.Context) error {
	api, err := s.GetApi(ctx, s.config.ApiGateway().ID())
	if err != nil {
		return err
	}

	// Get all routes and auth configurations
	routeKeys, err := s.config.ApiGateway().RouteKeys()
	if err != nil {
		return err
	}

	authTypes := s.config.ApiGateway().AuthType()
	authorizerIds := s.config.ApiGateway().AuthorizerId()

	// Validate 1:1 pairing (this should already be validated in param validation, but double-check)
	if len(routeKeys) != len(authTypes) || len(routeKeys) != len(authorizerIds) {
		return fmt.Errorf("route/auth configuration mismatch: %d routes, %d auth types, %d authorizer ids", 
			len(routeKeys), len(authTypes), len(authorizerIds))
	}

	// Create integrations and routes for each route/auth pair
	for i := 0; i < len(routeKeys); i++ {
		integration, err := s.CreateIntegration(ctx, api, i)
		if err != nil {
			return fmt.Errorf("failed to create integration for route %d: %w", i, err)
		}

		_, err = s.CreateRoute(ctx, api, integration, i)
		if err != nil {
			return fmt.Errorf("failed to create route %d: %w", i, err)
		}

		_, err = s.CreatePermission(ctx, api, i)
		if err != nil {
			return fmt.Errorf("failed to create permission for route %d: %w", i, err)
		}
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
func (s *ApiGatewayV2) CreateIntegration(ctx context.Context, api Api, routeIndex int) (Integration, error) {
	forwardedPrefixes, err := s.config.ApiGateway().ForwardedPrefixes()
	if err != nil {
		return Integration{}, err
	}

	if routeIndex >= len(forwardedPrefixes) {
		return Integration{}, fmt.Errorf("route index %d out of bounds, only %d prefixes available", routeIndex, len(forwardedPrefixes))
	}

	create := &apigatewayv2.CreateIntegrationInput{
		ApiId:                aws.String(api.ApiId),
		ConnectionType:       types.ConnectionTypeInternet,
		IntegrationType:      types.IntegrationTypeAwsProxy,
		IntegrationUri:       aws.String(s.config.Lambda().FunctionArn()),
		PayloadFormatVersion: aws.String("2.0"),
		RequestParameters: map[string]string{
			"overwrite:path":                      "/$request.path.proxy",
			"overwrite:header.X-Forwarded-Prefix": forwardedPrefixes[routeIndex],
		},
	}

	client := s.config.ApiGateway().Client()
	integration, err := client.CreateIntegration(ctx, create)
	if err != nil {
		return Integration{}, err
	}

	return Integration{
		ApiId:         api.ApiId,
		IntegrationId: *integration.IntegrationId,
	}, nil
}

func (s *ApiGatewayV2) CreateRoute(ctx context.Context, api Api, integration Integration, routeIndex int) (Route, error) {
	client := s.config.ApiGateway().Client()
	authTypeMap := map[string]types.AuthorizationType{
		"NONE":    types.AuthorizationTypeNone,
		"AWS_IAM": types.AuthorizationTypeAwsIam,
		"CUSTOM":  types.AuthorizationTypeCustom,
		"JWT":     types.AuthorizationTypeJwt,
	}

	routeKeys, err := s.config.ApiGateway().RouteKeys()
	if err != nil {
		return Route{}, err
	}

	authTypes := s.config.ApiGateway().AuthType()
	authorizerIds := s.config.ApiGateway().AuthorizerId()

	if routeIndex >= len(routeKeys) || routeIndex >= len(authTypes) || routeIndex >= len(authorizerIds) {
		return Route{}, fmt.Errorf("route index %d out of bounds", routeIndex)
	}

	authType, ok := authTypeMap[authTypes[routeIndex]]
	if !ok {
		return Route{}, fmt.Errorf("unsupported authorization type %s", authTypes[routeIndex])
	}

	create := &apigatewayv2.CreateRouteInput{
		ApiId:             aws.String(api.ApiId),
		RouteKey:          aws.String(routeKeys[routeIndex]),
		Target:            aws.String(fmt.Sprintf("integrations/%s", integration.IntegrationId)),
		AuthorizationType: authType,
		AuthorizerId:      aws.String(authorizerIds[routeIndex]),
	}

	log.Debug().
		Str("api_id", api.ApiId).
		Str("route_key", routeKeys[routeIndex]).
		Str("auth_type", authTypes[routeIndex]).
		Int("route_index", routeIndex).
		Msg("creating route")

	route, err := client.CreateRoute(ctx, create)
	if err != nil {
		return Route{}, err
	}

	authorizerId := ""
	if route.AuthorizerId != nil {
		authorizerId = *route.AuthorizerId
	}
	
	return Route{
		ApiId:             api.ApiId,
		RouteId:           *route.RouteId,
		RouteKey:          *route.RouteKey,
		AuthorizationType: string(route.AuthorizationType),
		AuthorizerId:      authorizerId,
	}, nil
}

func (s *ApiGatewayV2) CreatePermission(ctx context.Context, api Api, routeIndex int) (Permission, error) {
	sourceArns, err := s.config.ApiGateway().PermissionSourceArns()
	if err != nil {
		return Permission{}, err
	}

	if routeIndex >= len(sourceArns) {
		return Permission{}, fmt.Errorf("route index %d out of bounds, only %d source ARNs available", routeIndex, len(sourceArns))
	}

	// Create unique statement ID for each route
	statementId := fmt.Sprintf("%s-%d", s.config.ApiGateway().PermissionStatementId(api.ApiId), routeIndex)

	create := &lambda.AddPermissionInput{
		FunctionName: aws.String(s.config.Lambda().FunctionArn()),
		Action:       aws.String("lambda:InvokeFunction"),
		Principal:    aws.String("apigateway.amazonaws.com"),
		SourceArn:    aws.String(sourceArns[routeIndex]),
		StatementId:  aws.String(statementId),
	}

	_, err = s.config.Lambda().Client().AddPermission(ctx, create)
	if err != nil {
		return Permission{}, err
	}

	return Permission{
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

	// Convert AWS auth type to lowercase to match deploy format
	authType := strings.ToLower(route.AuthorizationType)
	
	log.Info().
		Str("id", route.ApiId).
		Str("route", route.RouteKey).
		Str("auth", authType).
		Str("action", "delete").
		Msg("apigatewayv2")

	return s.config.ApiGateway().Client().DeleteRoute(ctx, input)
}

func (s *ApiGatewayV2) DeleteIntegration(ctx context.Context, integration Integration) (*apigatewayv2.DeleteIntegrationOutput, error) {
	input := &apigatewayv2.DeleteIntegrationInput{
		ApiId:         aws.String(integration.ApiId),
		IntegrationId: aws.String(integration.IntegrationId),
	}

	return s.config.ApiGateway().Client().DeleteIntegration(ctx, input)
}

func (s *ApiGatewayV2) DeletePermissions(ctx context.Context, permission Permission) error {
	var apiErr smithy.APIError

	input := &lambda.RemovePermissionInput{
		FunctionName: aws.String(permission.FunctionArn),
		StatementId:  aws.String(permission.StatementId),
	}

	_, err := s.config.Lambda().Client().RemovePermission(ctx, input)
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

		result, err := s.config.ApiGateway().Client().GetApis(ctx, input)
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
	var integrations []Integration
	var nextToken *string

	integrations, err := s.GetIntegrations(ctx, apis)
	if err != nil {
		return nil, err
	}

	for _, api := range apis {
		for {
			input := &apigatewayv2.GetRoutesInput{
				ApiId:     aws.String(api.ApiId),
				NextToken: nextToken,
			}

			result, err := s.config.ApiGateway().Client().GetRoutes(ctx, input)
			if err != nil {
				return nil, err
			}

			for _, route := range result.Items {
				for _, integration := range integrations {
					target := fmt.Sprintf("integrations/%s", integration.IntegrationId)
					if target == *route.Target {
						authorizerId := ""
						if route.AuthorizerId != nil {
							authorizerId = *route.AuthorizerId
						}
						routes = append(routes, Route{
							ApiId:             api.ApiId,
							RouteId:           *route.RouteId,
							RouteKey:          *route.RouteKey,
							AuthorizationType: string(route.AuthorizationType),
							AuthorizerId:      authorizerId,
						})
					}
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

			result, err := s.config.ApiGateway().Client().GetIntegrations(ctx, input)
			if err != nil {
				return nil, err
			}

			for _, integration := range result.Items {
				if s.config.Lambda().FunctionArn() == *integration.IntegrationUri {
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
	
	// Get Lambda's resource-based policy
	policy, err := s.config.Lambda().Client().GetPolicy(ctx, &lambda.GetPolicyInput{
		FunctionName: aws.String(s.config.Lambda().FunctionArn()),
	})
	if err != nil {
		var resourceNotFound *lambdatypes.ResourceNotFoundException
		if errors.As(err, &resourceNotFound) {
			return permissions, nil // No policy exists
		}
		return nil, fmt.Errorf("failed to get Lambda policy: %w", err)
	}
	
	// Parse policy JSON to extract ALL API Gateway statement IDs
	var policyDoc struct {
		Statement []struct {
			Sid       string `json:"Sid"`
			Principal struct {
				Service string `json:"Service"`
			} `json:"Principal"`
		} `json:"Statement"`
	}
	
	if err := json.Unmarshal([]byte(*policy.Policy), &policyDoc); err != nil {
		return nil, fmt.Errorf("failed to parse Lambda policy: %w", err)
	}
	
	// Find ALL statements where Principal.Service is "apigateway.amazonaws.com"
	for _, stmt := range policyDoc.Statement {
		if stmt.Principal.Service == "apigateway.amazonaws.com" {
			permissions = append(permissions, Permission{
				FunctionArn: s.config.Lambda().FunctionArn(),
				StatementId: stmt.Sid,
			})
		}
	}
	
	return permissions, nil
}
