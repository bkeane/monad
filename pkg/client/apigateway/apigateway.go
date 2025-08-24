package apigateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/smithy-go"
	"github.com/rs/zerolog/log"
)

type ApiGatewayConfig interface {
	Client() *apigatewayv2.Client
	ApiId() string
	Route() []string
	Auth() []string
	AuthType() []string
	AuthorizerId() []string
	ForwardedPrefixes() ([]string, error)
	PermissionSourceArns() ([]string, error)
	PermissionStatementId(apiId string) string
}

type LambdaConfig interface {
	FunctionArn() string
	Client() *lambda.Client
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

type Summary struct {
	RoutesDeleted       []Route
	RoutesCreated       []Route
	IntegrationsDeleted []Integration
	IntegrationsCreated []Integration
	PermissionsDeleted  []Permission
	PermissionsCreated  []Permission
}

type Client struct {
	apigateway ApiGatewayConfig
	lambda     LambdaConfig
}

//
// Derive
//

func Derive(apigateway ApiGatewayConfig, lambda LambdaConfig) *Client {
	return &Client{
		apigateway: apigateway,
		lambda:     lambda,
	}
}

func (s *Client) Mount(ctx context.Context) error {
	// Call internal unmount silently (don't log deletes)
	if _, err := s.unmount(ctx); err != nil {
		return err
	}

	// Call internal mount and log only the creates as action=put
	summary, err := s.mount(ctx)
	if err != nil {
		return err
	}

	// Log only the creates as action=put for consistency with other services
	for _, route := range summary.RoutesCreated {
		authType := strings.ToLower(route.AuthorizationType)
		log.Info().
			Str("id", route.ApiId).
			Str("route", route.RouteKey).
			Str("auth", authType).
			Str("action", "put").
			Msg("apigatewayv2")
	}

	return nil
}

func (s *Client) Unmount(ctx context.Context) error {
	// Call internal unmount and log the deletes as action=delete
	summary, err := s.unmount(ctx)
	if err != nil {
		return err
	}

	// Log the deletes as action=delete
	for _, route := range summary.RoutesDeleted {
		authType := strings.ToLower(route.AuthorizationType)
		log.Info().
			Str("id", route.ApiId).
			Str("route", route.RouteKey).
			Str("auth", authType).
			Str("action", "delete").
			Msg("apigatewayv2")
	}

	return nil
}

// Internal methods that return summaries of work done
func (s *Client) mount(ctx context.Context) (Summary, error) {
	var summary Summary

	api, err := s.GetApi(ctx, s.apigateway.ApiId())
	if err != nil {
		return summary, err
	}

	routeKeys := s.apigateway.Route()
	authTypes := s.apigateway.AuthType()
	authorizerIds := s.apigateway.AuthorizerId()

	// Validate 1:1 pairing (this should already be validated in param validation, but double-check)
	if len(routeKeys) != len(authTypes) || len(routeKeys) != len(authorizerIds) {
		return summary, fmt.Errorf("route/auth configuration mismatch: %d routes, %d auth types, %d authorizer ids",
			len(routeKeys), len(authTypes), len(authorizerIds))
	}

	// Create integrations and routes for each route/auth pair
	for i := 0; i < len(routeKeys); i++ {
		integration, err := s.CreateIntegration(ctx, api, i)
		if err != nil {
			return summary, fmt.Errorf("failed to create integration for route %d: %w", i, err)
		}
		summary.IntegrationsCreated = append(summary.IntegrationsCreated, integration)

		route, err := s.CreateRoute(ctx, api, integration, i)
		if err != nil {
			return summary, fmt.Errorf("failed to create route %d: %w", i, err)
		}
		summary.RoutesCreated = append(summary.RoutesCreated, route)

		permission, err := s.CreatePermission(ctx, api, i)
		if err != nil {
			return summary, fmt.Errorf("failed to create permission for route %d: %w", i, err)
		}
		summary.PermissionsCreated = append(summary.PermissionsCreated, permission)
	}

	return summary, nil
}

func (s *Client) unmount(ctx context.Context) (Summary, error) {
	var summary Summary

	apis, err := s.GetApis(ctx)
	if err != nil {
		return summary, err
	}

	routes, err := s.GetRoutes(ctx, apis)
	if err != nil {
		return summary, err
	}

	integrations, err := s.GetIntegrations(ctx, apis)
	if err != nil {
		return summary, err
	}

	permissions, err := s.GetPermissions(ctx, apis)
	if err != nil {
		return summary, err
	}

	for _, route := range routes {
		if _, err := s.DeleteRoute(ctx, route); err != nil {
			return summary, err
		}
		summary.RoutesDeleted = append(summary.RoutesDeleted, route)
	}

	for _, integration := range integrations {
		if _, err := s.DeleteIntegration(ctx, integration); err != nil {
			return summary, err
		}
		summary.IntegrationsDeleted = append(summary.IntegrationsDeleted, integration)
	}

	for _, permission := range permissions {
		if err := s.DeletePermissions(ctx, permission); err != nil {
			return summary, err
		}
		summary.PermissionsDeleted = append(summary.PermissionsDeleted, permission)
	}

	return summary, nil
}

// Internal create functions (no logging)
func (s *Client) CreateIntegration(ctx context.Context, api Api, routeIndex int) (Integration, error) {
	forwardedPrefixes, err := s.apigateway.ForwardedPrefixes()
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
		IntegrationUri:       aws.String(s.lambda.FunctionArn()),
		PayloadFormatVersion: aws.String("2.0"),
		RequestParameters: map[string]string{
			"overwrite:path":                      "/$request.path.proxy",
			"overwrite:header.X-Forwarded-Prefix": forwardedPrefixes[routeIndex],
		},
	}

	client := s.apigateway.Client()
	integration, err := client.CreateIntegration(ctx, create)
	if err != nil {
		return Integration{}, err
	}

	return Integration{
		ApiId:         api.ApiId,
		IntegrationId: *integration.IntegrationId,
	}, nil
}

func (s *Client) CreateRoute(ctx context.Context, api Api, integration Integration, routeIndex int) (Route, error) {
	client := s.apigateway.Client()
	authTypeMap := map[string]types.AuthorizationType{
		"NONE":    types.AuthorizationTypeNone,
		"AWS_IAM": types.AuthorizationTypeAwsIam,
		"CUSTOM":  types.AuthorizationTypeCustom,
		"JWT":     types.AuthorizationTypeJwt,
	}

	routeKeys := s.apigateway.Route()
	authTypes := s.apigateway.AuthType()
	authorizerIds := s.apigateway.AuthorizerId()

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

func (s *Client) CreatePermission(ctx context.Context, api Api, routeIndex int) (Permission, error) {
	sourceArns, err := s.apigateway.PermissionSourceArns()
	if err != nil {
		return Permission{}, err
	}

	if routeIndex >= len(sourceArns) {
		return Permission{}, fmt.Errorf("route index %d out of bounds, only %d source ARNs available", routeIndex, len(sourceArns))
	}

	// Create unique statement ID for each route
	statementId := fmt.Sprintf("%s-%d", s.apigateway.PermissionStatementId(api.ApiId), routeIndex)

	create := &lambda.AddPermissionInput{
		FunctionName: aws.String(s.lambda.FunctionArn()),
		Action:       aws.String("lambda:InvokeFunction"),
		Principal:    aws.String("apigateway.amazonaws.com"),
		SourceArn:    aws.String(sourceArns[routeIndex]),
		StatementId:  aws.String(statementId),
	}

	_, err = s.lambda.Client().AddPermission(ctx, create)
	if err != nil {
		return Permission{}, err
	}

	return Permission{
		FunctionArn: *create.FunctionName,
		StatementId: *create.StatementId,
	}, nil
}

// Internal delete functions (no logging)
func (s *Client) DeleteRoute(ctx context.Context, route Route) (*apigatewayv2.DeleteRouteOutput, error) {
	input := &apigatewayv2.DeleteRouteInput{
		ApiId:   aws.String(route.ApiId),
		RouteId: aws.String(route.RouteId),
	}

	return s.apigateway.Client().DeleteRoute(ctx, input)
}

func (s *Client) DeleteIntegration(ctx context.Context, integration Integration) (*apigatewayv2.DeleteIntegrationOutput, error) {
	input := &apigatewayv2.DeleteIntegrationInput{
		ApiId:         aws.String(integration.ApiId),
		IntegrationId: aws.String(integration.IntegrationId),
	}

	return s.apigateway.Client().DeleteIntegration(ctx, input)
}

func (s *Client) DeletePermissions(ctx context.Context, permission Permission) error {
	var apiErr smithy.APIError

	input := &lambda.RemovePermissionInput{
		FunctionName: aws.String(permission.FunctionArn),
		StatementId:  aws.String(permission.StatementId),
	}

	_, err := s.lambda.Client().RemovePermission(ctx, input)
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
func (s *Client) GetApis(ctx context.Context) ([]Api, error) {
	var apis []Api
	var nextToken *string

	for {
		input := &apigatewayv2.GetApisInput{
			NextToken: nextToken,
		}

		result, err := s.apigateway.Client().GetApis(ctx, input)
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

func (s *Client) GetRoutes(ctx context.Context, apis []Api) ([]Route, error) {
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

			result, err := s.apigateway.Client().GetRoutes(ctx, input)
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

func (s *Client) GetApi(ctx context.Context, apiId string) (Api, error) {
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

func (s *Client) GetIntegrations(ctx context.Context, apis []Api) ([]Integration, error) {
	var integrations []Integration
	var nextToken *string

	for _, api := range apis {
		for {
			input := &apigatewayv2.GetIntegrationsInput{
				ApiId:     aws.String(api.ApiId),
				NextToken: nextToken,
			}

			result, err := s.apigateway.Client().GetIntegrations(ctx, input)
			if err != nil {
				return nil, err
			}

			for _, integration := range result.Items {
				if s.lambda.FunctionArn() == *integration.IntegrationUri {
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

func (s *Client) GetPermissions(ctx context.Context, apis []Api) ([]Permission, error) {
	var permissions []Permission

	// Get Lambda's resource-based policy
	policy, err := s.lambda.Client().GetPolicy(ctx, &lambda.GetPolicyInput{
		FunctionName: aws.String(s.lambda.FunctionArn()),
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
				FunctionArn: s.lambda.FunctionArn(),
				StatementId: stmt.Sid,
			})
		}
	}

	return permissions, nil
}
