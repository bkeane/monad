package param

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/rs/zerolog/log"

	"github.com/bkeane/monad/internal/uriopt"
	v "github.com/go-ozzo/ozzo-validation/v4"
)

type EventBridgeConfig struct {
	Client       *eventbridge.Client `arg:"-" json:"-"`
	BusName      string              `arg:"--bus,env:MONAD_BUS" placeholder:"name" help:"eventbridge bus name" default:"default"`
	RuleTemplate string              `arg:"--rule,env:MONAD_RULE" placeholder:"template" help:"string | file://rule.json" default:"no-rule"`
	Region       string              `arg:"--bus-region,env:MONAD_BUS_REGION" placeholder:"name" help:"eventbridge region" default:"caller-region"`
}

func (e *EventBridgeConfig) Process(ctx context.Context, awsconfig aws.Config) error {
	e.Client = eventbridge.NewFromConfig(awsconfig)

	if e.BusName == "" {
		e.BusName = "default"
	}

	if e.Region == "" {
		e.Region = awsconfig.Region
	}

	if e.BusName != "" {
		var validBusNames []string

		buses, err := e.Client.ListEventBuses(ctx, &eventbridge.ListEventBusesInput{})
		if err != nil {
			return err
		}

		for _, bus := range buses.EventBuses {
			validBusNames = append(validBusNames, *bus.Name)
		}

		if !slices.Contains(validBusNames, e.BusName) {
			log.Error().
				Str("given", e.BusName).
				Strs("valid_names", validBusNames).
				Msg("bus name not found")

			return fmt.Errorf("bus name %s not found", e.BusName)
		}
	}

	if e.RuleTemplate != "" {
		var content string
		var err error

		if strings.HasSuffix(e.RuleTemplate, ".json") {
			content, err = uriopt.Json(e.RuleTemplate)
			if err != nil {
				return fmt.Errorf("failed to read provided rule template: %w", err)
			}
		} else {
			content, err = uriopt.String(e.RuleTemplate)
			if err != nil {
				return fmt.Errorf("failed to read provided rule template: %w", err)
			}
		}

		e.RuleTemplate = content
	}

	return e.Validate()
}

func (e *EventBridgeConfig) Validate() error {
	return v.ValidateStruct(e,
		v.Field(&e.Client, v.Required),
		v.Field(&e.BusName, v.Required),
		v.Field(&e.Region, v.Required),
	)
}
