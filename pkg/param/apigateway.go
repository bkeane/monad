package param

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/rs/zerolog/log"
)

type ApiGateway struct {
	Client       *apigatewayv2.Client `arg:"-" json:"-"`
	Api          string               `arg:"--api" placeholder:"id|name" help:"api gateway" default:"no-gateway"`
	Id           string               `arg:"-" json:"-"` // `arg:"--api-id" placeholder:"id" help:"api gateway id" default:"no-gateway-id"`
	Auth         string               `arg:"--auth" placeholder:"id|name" help:"none | aws_iam | NAME | ID" default:"aws_iam"`
	Region       string               `arg:"--api-region" placeholder:"name" help:"api gateway region name" default:"caller-region"`
	AuthType     string               `arg:"-" json:"-"` // `arg:"--api-auth-type" placeholder:"name" default:"AWS_IAM" help:"NONE | AWS_IAM | CUSTOM | JWT"`
	AuthorizerId string               `arg:"-" json:"-"` // `arg:"--api-auth-id" placeholder:"id" help:"authorizer id to use" default:"no-authorizer-id"`
}

func (a *ApiGateway) Validate(ctx context.Context, awsconfig aws.Config) error {
	a.Client = apigatewayv2.NewFromConfig(awsconfig)

	if a.Region == "" {
		a.Region = awsconfig.Region
	}

	if a.Api != "" && a.Auth == "" {
		a.Auth = "aws_iam"
	}

	// pre-compute validations
	err := v.ValidateStruct(a,
		v.Field(&a.Api, v.When(a.Auth != "", v.Required)),
		v.Field(&a.Region, v.Required),
		v.Field(&a.Client, v.Required),
	)

	if err != nil {
		return err
	}

	// Resolve api and authorizer names/ids
	if a.Api != "" {
		a.Id, err = findApiId(ctx, a.Client, a.Api)
		if err != nil {
			return err
		}

		a.AuthorizerId, a.AuthType, err = findAuthIdAndType(ctx, a.Client, a.Id, a.Auth)
		if err != nil {
			return err
		}
	}

	return v.ValidateStruct(a,
		v.Field(&a.Id, v.When(a.AuthType != "" || a.AuthorizerId != "", v.Required)),
		v.Field(&a.AuthType, v.When(a.AuthType != "", v.In("NONE", "AWS_IAM", "CUSTOM", "JWT")), v.When(a.AuthorizerId != "", v.Required)),
		v.Field(&a.AuthorizerId, v.When(a.AuthType != "" && (a.AuthType == "CUSTOM" || a.AuthType == "JWT"), v.Required)),
	)
}

func findApiId(ctx context.Context, client *apigatewayv2.Client, apiIdOrName string) (string, error) {
	var apiIds []string
	var apiNames []string

	apis, err := client.GetApis(ctx, &apigatewayv2.GetApisInput{})
	if err != nil {
		return "", err
	}

	for _, api := range apis.Items {
		apiIds = append(apiIds, *api.ApiId)
		apiNames = append(apiNames, *api.Name)

		if *api.Name == apiIdOrName || *api.ApiId == apiIdOrName {
			return *api.ApiId, nil
		}
	}

	log.Error().
		Str("given", apiIdOrName).
		Strs("valid_ids", apiIds).
		Strs("valid_names", apiNames).
		Msg("api not found")

	return "", fmt.Errorf("api %s not found", apiIdOrName)
}

func findAuthIdAndType(ctx context.Context, client *apigatewayv2.Client, apiId string, authIdOrName string) (authId string, authType string, err error) {
	switch strings.ToLower(authIdOrName) {
	case "none", "aws_iam":
		return "", strings.ToUpper(authIdOrName), nil
	default:
		var authorizerIds []string
		var authorizerNames []string

		authorizerNames = append(authorizerNames, "none", "aws_iam")

		list, err := client.GetAuthorizers(ctx, &apigatewayv2.GetAuthorizersInput{
			ApiId: aws.String(apiId),
		})

		if err != nil {
			return "", "", err
		}

		for _, authorizer := range list.Items {
			authorizerIds = append(authorizerIds, *authorizer.AuthorizerId)
			authorizerNames = append(authorizerNames, *authorizer.Name)

			if *authorizer.Name == authIdOrName || *authorizer.AuthorizerId == authIdOrName {
				authId = *authorizer.AuthorizerId
				authType = string(authorizer.AuthorizerType)
				return authId, authType, nil
			}
		}

		log.Error().
			Str("api", apiId).
			Strs("valid_ids", authorizerIds).
			Strs("valid_names", authorizerNames).
			Str("given", authIdOrName).
			Msg("authorizer not found")

		return "", "", fmt.Errorf("authorizer %s not found", authIdOrName)
	}
}
