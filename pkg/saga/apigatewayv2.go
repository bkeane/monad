package saga

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/bkeane/monad/pkg/config/release"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/smithy-go"
	"github.com/rs/zerolog"
)

type ApiGatewayV2 struct {
	release      release.Config
	lambda       *lambda.Client
	apigatewayv2 *apigatewayv2.Client
	log          *zerolog.Logger
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

func (s ApiGatewayV2) Init(ctx context.Context, r release.Config) *ApiGatewayV2 {
	return &ApiGatewayV2{
		release:      r,
		lambda:       lambda.NewFromConfig(r.AwsConfig),
		apigatewayv2: apigatewayv2.NewFromConfig(r.AwsConfig),
		log:          zerolog.Ctx(ctx),
	}
}

func (s *ApiGatewayV2) Do(ctx context.Context) error {
	return s.Ensure(ctx)
}

func (s *ApiGatewayV2) Undo(ctx context.Context) error {
	return s.Destroy(ctx)
}

func (s *ApiGatewayV2) Ensure(ctx context.Context) error {
	if s.release.Substrate.ApiGateway.Enable != nil {
		if !*s.release.Substrate.ApiGateway.Enable {
			s.log.Info().Msg("destroying apigatewayv2 routes")
			return s.Destroy(ctx)
		}

		if *s.release.Substrate.ApiGateway.Enable {
			s.log.Info().Msg("deploying apigatewayv2 routes")
			if err := s.Destroy(ctx); err != nil {
				return err
			}

			return s.Deploy(ctx)
		}
	}

	s.log.Info().Msg("leaving apigatewayv2 routes unchanged")
	return nil
}

func (s *ApiGatewayV2) Deploy(ctx context.Context) error {
	api, err := s.GetApi(ctx, *s.release.Substrate.ApiGateway.Id)
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
		IntegrationUri:       aws.String(s.release.FunctionArn()),
		PayloadFormatVersion: aws.String("2.0"),
		RequestParameters: map[string]string{
			"overwrite:path":                      "/$request.path.proxy",
			"overwrite:header.X-Forwarded-Prefix": s.release.ForwardedForPrefix(),
		},
	}

	integration, err := s.apigatewayv2.CreateIntegration(ctx, create)
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

	if s.release.Substrate.ApiGateway.AuthType == nil {
		s.log.Debug().Msg("no auth type specified, defaulting to AWS_IAM")
		s.release.Substrate.ApiGateway.AuthType = aws.String("AWS_IAM")
	}

	authType, ok := authTypeMap[*s.release.Substrate.ApiGateway.AuthType]
	if !ok {
		return Route{}, fmt.Errorf("unsupported authorization type %s", *s.release.Substrate.ApiGateway.AuthType)
	}

	create := &apigatewayv2.CreateRouteInput{
		ApiId:             aws.String(api.ApiId),
		RouteKey:          aws.String(s.release.RouteKey()),
		Target:            aws.String(fmt.Sprintf("integrations/%s", integration.IntegrationId)),
		AuthorizationType: authType,
		AuthorizerId:      s.release.Substrate.ApiGateway.AuthorizerId,
	}

	s.log.Debug().
		Str("api_id", api.ApiId).
		Str("route_key", s.release.RouteKey()).
		Str("auth_type", *s.release.Substrate.ApiGateway.AuthType).
		Msg("creating route")

	route, err := s.apigatewayv2.CreateRoute(ctx, create)
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
		FunctionName: aws.String(s.release.FunctionArn()),
		Action:       aws.String("lambda:InvokeFunction"),
		Principal:    aws.String("apigateway.amazonaws.com"),
		SourceArn:    aws.String(s.release.ApiGwPermissionSourceArn()),
		StatementId:  aws.String(s.release.ApiGwPermissionStatementId(api.ApiId)),
	}

	_, err := s.lambda.AddPermission(ctx, create)
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

	s.log.Debug().
		Str("api_id", route.ApiId).
		Str("route_key", s.release.RouteKey()).
		Msg("deleting route")

	return s.apigatewayv2.DeleteRoute(ctx, input)
}

func (s *ApiGatewayV2) DeleteIntegration(ctx context.Context, integration Integration) (*apigatewayv2.DeleteIntegrationOutput, error) {
	input := &apigatewayv2.DeleteIntegrationInput{
		ApiId:         aws.String(integration.ApiId),
		IntegrationId: aws.String(integration.IntegrationId),
	}

	return s.apigatewayv2.DeleteIntegration(ctx, input)
}

func (s *ApiGatewayV2) DeletePermissions(ctx context.Context, permission Permission) error {
	var apiErr smithy.APIError

	input := &lambda.RemovePermissionInput{
		FunctionName: aws.String(permission.FunctionArn),
		StatementId:  aws.String(permission.StatementId),
	}

	_, err := s.lambda.RemovePermission(ctx, input)
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

		result, err := s.apigatewayv2.GetApis(ctx, input)
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

			result, err := s.apigatewayv2.GetRoutes(ctx, input)
			if err != nil {
				return nil, err
			}

			for _, route := range result.Items {
				if slices.Contains(s.release.PossibleRouteKeys(), *route.RouteKey) {
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

			result, err := s.apigatewayv2.GetIntegrations(ctx, input)
			if err != nil {
				return nil, err
			}

			for _, integration := range result.Items {
				if slices.Contains(s.release.PossibleFunctionArns(), *integration.IntegrationUri) {
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
		for _, arn := range s.release.PossibleFunctionArns() {
			permissions = append(permissions, Permission{
				ApiId:       api.ApiId,
				FunctionArn: arn,
				StatementId: s.release.ApiGwPermissionStatementId(api.ApiId),
			})
		}
	}

	return permissions, nil
}
