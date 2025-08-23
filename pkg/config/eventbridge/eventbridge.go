package eventbridge

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/caarlos0/env/v11"

	eventbridgetypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/rs/zerolog/log"
)

type Basis interface {
	AwsConfig() aws.Config
	Region() string
	Name() string
	Render(string) (string, error)
	Tags() map[string]string
}

//
// Convention
//

type Config struct {
	basis        Basis
	client       *eventbridge.Client
	RegionName      string `env:"MONAD_BUS_REGION"`
	EventBusName    string `env:"MONAD_BUS_NAME" flag:"--bus" usage:"EventBridge bus name"`
	ruleName     string
	ruleDocument string
	ruleTemplate string
	RulePath     string `env:"MONAD_RULE" flag:"--rule" usage:"EventBridge rule template file path"`
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

	cfg.basis = basis
	cfg.client = eventbridge.NewFromConfig(basis.AwsConfig())

	if cfg.RegionName == "" {
		cfg.RegionName = basis.Region()
	}

	if cfg.EventBusName == "" {
		cfg.EventBusName = "default"
	}

	if cfg.ruleName == "" {
		cfg.ruleName = basis.Name()
	}

	if cfg.RulePath != "" {
		bytes, err := os.ReadFile(cfg.RulePath)
		if err != nil {
			return nil, err
		}

		cfg.ruleTemplate = string(bytes)
	}

	cfg.ruleDocument, err = basis.Render(cfg.ruleTemplate)
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
		v.Field(&c.basis, v.Required),
		v.Field(&c.client, v.Required),
		v.Field(&c.EventBusName, v.By(c.exists)),
	)
}

// exists is a validation to ensure the bus provided is present.
func (c *Config) exists(value interface{}) error {
	ctx := context.Background()

	busName, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid bus name format: %s", c.EventBusName)
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
func (c *Config) Region() string { return c.RegionName }

// BusName returns the EventBridge custom bus name
func (c *Config) BusName() string { return c.EventBusName }

// RuleTemplate returns the EventBridge rule name
func (c *Config) RuleName() string { return c.ruleName }

// PermissionStatementId returns the Lambda permission statement ID for EventBridge
func (c *Config) PermissionStatementId() string {
	return strings.Join([]string{"eventbridge", c.BusName(), c.basis.Name()}, "-")
}

// RuleDocument returns the eventbridge rule definition
func (c *Config) RuleDocument() string {
	return c.ruleDocument
}

// Tags returns standardized EventBridge resource tags
func (c *Config) Tags() []eventbridgetypes.Tag {
	var tags []eventbridgetypes.Tag
	for key, value := range c.basis.Tags() {
		tags = append(tags, eventbridgetypes.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}
	return tags
}
