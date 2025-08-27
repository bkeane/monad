package eventbridge

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/bkeane/monad/pkg/basis/caller"
	"github.com/bkeane/monad/pkg/basis/defaults"
	"github.com/bkeane/monad/pkg/basis/resource"

	eventbridgetypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/caarlos0/env/v11"
	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/rs/zerolog/log"
)

type Basis interface {
	Caller() (*caller.Basis, error)
	Defaults() (*defaults.Basis, error)
	Resource() (*resource.Basis, error)
	Render(string) (string, error)
}

//
// Convention
//

type Config struct {
	client                  *eventbridge.Client
	EventBridgeRegionName   string `env:"MONAD_BUS_REGION"`
	EventBridgeBusName      string `env:"MONAD_BUS_NAME" flag:"--bus" usage:"EventBridge bus name" hint:"name"`
	EventBridgeRuleName     string
	EventBridgeRulePath     string `env:"MONAD_RULE" flag:"--rule" usage:"EventBridge rule template file path" hint:"path"`
	EventBridgeRuleTemplate string
	EventBridgeRuleDocument string
	caller                  *caller.Basis
	defaults                *defaults.Basis
	resource                *resource.Basis
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

	cfg.caller, err = basis.Caller()
	if err != nil {
		return nil, err
	}

	cfg.defaults, err = basis.Defaults()
	if err != nil {
		return nil, err
	}

	cfg.resource, err = basis.Resource()
	if err != nil {
		return nil, err
	}

	cfg.client = eventbridge.NewFromConfig(cfg.caller.AwsConfig())

	if cfg.EventBridgeRegionName == "" {
		cfg.EventBridgeRegionName = cfg.caller.AwsConfig().Region
	}

	if cfg.EventBridgeBusName == "" {
		cfg.EventBridgeBusName = "default"
	}

	if cfg.EventBridgeRuleName == "" {
		cfg.EventBridgeRuleName = cfg.resource.Name()
	}

	if cfg.EventBridgeRulePath != "" {
		bytes, err := os.ReadFile(cfg.EventBridgeRulePath)
		if err != nil {
			return nil, err
		}

		cfg.EventBridgeRuleTemplate = string(bytes)
	}

	cfg.EventBridgeRuleDocument, err = basis.Render(cfg.EventBridgeRuleTemplate)
	if err != nil {
		return nil, err
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
		v.Field(&c.EventBridgeBusName, v.By(c.exists)),
	)
}

// exists is a validation to ensure the bus provided is present.
func (c *Config) exists(value interface{}) error {
	ctx := context.Background()

	busName, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid bus name format: %s", c.EventBridgeBusName)
	}

	var validBusNames []string
	buses, err := c.client.ListEventBuses(ctx, &eventbridge.ListEventBusesInput{})
	if err != nil {
		return err
	}

	for _, bus := range buses.EventBuses {
		validBusNames = append(validBusNames, *bus.Name)
	}

	if !slices.Contains(validBusNames, busName) {
		log.Error().
			Str("given", busName).
			Strs("valid_names", validBusNames).
			Msg("bus name not found")

		return fmt.Errorf("bus not found")
	}

	return nil
}

//
// Accessors
//

// Client returns the AWS EventBridge service client
func (c *Config) Client() *eventbridge.Client { return c.client }

// Region returns the AWS region for EventBridge deployment
func (c *Config) Region() string { return c.EventBridgeRegionName }

// BusName returns the EventBridge custom bus name
func (c *Config) BusName() string { return c.EventBridgeBusName }

// RuleTemplate returns the EventBridge rule name
func (c *Config) RuleName() string { return c.EventBridgeRuleName }

// PermissionStatementId returns the Lambda permission statement ID for EventBridge
func (c *Config) PermissionStatementId() string {
	return strings.Join([]string{"eventbridge", c.BusName(), c.resource.Name()}, "-")
}

// RuleDocument returns the eventbridge rule definition
func (c *Config) RuleDocument() string {
	return c.EventBridgeRuleDocument
}

// Tags returns standardized EventBridge resource tags
func (c *Config) Tags() []eventbridgetypes.Tag {
	var tags []eventbridgetypes.Tag
	for key, value := range c.resource.Tags() {
		tags = append(tags, eventbridgetypes.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}
	return tags
}
