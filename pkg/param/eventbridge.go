package param

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/rs/zerolog/log"

	"github.com/bkeane/monad/internal/uriopt"
	v "github.com/go-ozzo/ozzo-validation/v4"
)

type EventBridge struct {
	Client       *eventbridge.Client `arg:"-" json:"-"`
	BusName      string              `arg:"--bus-name" placeholder:"name" help:"eventbridge bus name" default:"default"`
	RuleTemplate string              `arg:"--bus-rule" placeholder:"template" help:"{} | file://rule.json" default:"no-rule"`
	Region       string              `arg:"--bus-region" placeholder:"name" help:"eventbridge region name" default:"caller-region"`
}

func (e *EventBridge) Validate(ctx context.Context, awsconfig aws.Config) error {
	e.Client = eventbridge.NewFromConfig(awsconfig)

	if e.BusName == "" {
		e.BusName = "default"
	}

	if e.BusName != "" {
		var validBusNames []string

		buses, err := e.Client.ListEventBuses(ctx, &eventbridge.ListEventBusesInput{})
		if err != nil {
			return err
		}

		for _, bus := range buses.EventBuses {
			validBusNames = append(validBusNames, *bus.Name)
			if *bus.Name == e.BusName {
				e.BusName = *bus.Name
				break
			}
		}

		log.Error().
			Str("given", e.BusName).
			Strs("valid_names", validBusNames).
			Msg("bus name not found")

		return fmt.Errorf("bus name %s not found", e.BusName)
	}

	if e.RuleTemplate != "" {
		content, err := uriopt.Json(e.RuleTemplate)
		if err != nil {
			return fmt.Errorf("failed to read provided rule template")
		}
		e.RuleTemplate = content
	}

	return v.ValidateStruct(e,
		v.Field(&e.BusName, v.Required),
	)
}
