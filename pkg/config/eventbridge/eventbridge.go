package eventbridge

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode"

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
	EventBridgeRulePaths    []string `env:"MONAD_RULE" flag:"--rule" usage:"EventBridge rule template file paths" hint:"path"`
	EventBridgeRulesMap     map[string]string
	caller                  *caller.Basis
	defaults                *defaults.Basis
	resource                *resource.Basis
}

//
// Helper functions
//

// extractRuleName converts filename to rule name by removing all extensions
// e.g., "s3.json.tmpl" -> "s3", "schedule.yaml" -> "schedule"
func extractRuleName(filePath string) string {
	filename := filepath.Base(filePath)
	// Keep removing extensions until no more dots
	for strings.Contains(filename, ".") {
		filename = strings.TrimSuffix(filename, filepath.Ext(filename))
	}
	return filename
}

// chomp removes leading and trailing whitespace
func chomp(s string) string {
	s = strings.TrimLeftFunc(s, unicode.IsSpace)
	s = strings.TrimRightFunc(s, unicode.IsSpace)
	return s
}

// processRuleContent applies chomping to schedule expressions
func processRuleContent(content string) string {
	// First chomp to check the prefix without leading whitespace
	chomped := chomp(content)
	// If it's a schedule expression, return the chomped version
	if strings.HasPrefix(chomped, "cron(") || strings.HasPrefix(chomped, "rate(") {
		return chomped
	}
	// Leave event patterns as-is (JSON should not be chomped)
	return content
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

	cfg.EventBridgeRulesMap = make(map[string]string)

	if len(cfg.EventBridgeRulePaths) == 0 {
		// Use default rule
		defaultTemplate := cfg.defaults.RuleTemplate()
		defaultDocument, err := basis.Render(defaultTemplate)
		if err != nil {
			return nil, err
		}
		cfg.EventBridgeRulesMap[cfg.resource.Name()] = processRuleContent(defaultDocument)
	} else {
		// Process multiple rule files with duplicate name detection
		for _, path := range cfg.EventBridgeRulePaths {
			baseRuleName := extractRuleName(path)
			// Prefix with resource name to avoid cross-repo/branch collisions
			ruleName := fmt.Sprintf("%s-%s", cfg.resource.Name(), baseRuleName)
			
			// Check for duplicate names
			if _, exists := cfg.EventBridgeRulesMap[ruleName]; exists {
				return nil, fmt.Errorf("duplicate rule name '%s' derived from file '%s'", ruleName, path)
			}
			
			bytes, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}
			
			template := string(bytes)
			document, err := basis.Render(template)
			if err != nil {
				return nil, err
			}
			
			cfg.EventBridgeRulesMap[ruleName] = processRuleContent(document)
		}
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
		v.Field(&c.EventBridgeBusName, v.By(c.emptyOrExists)),
		v.Field(&c.EventBridgeRulesMap, 
			v.By(c.validateRulesMap),
		),
		v.Field(&c.EventBridgeRegionName, v.Required),
	)
}

// exists is a validation to ensure the bus provided is present.
func (c *Config) emptyOrExists(value interface{}) error {
	ctx := context.Background()

	busName, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid bus name format: %s", c.EventBridgeBusName)
	}

	if busName == "" {
		return nil
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

func (c *Config) validateRulesMap(value interface{}) error {
	rules, ok := value.(map[string]string)
	if !ok {
		return fmt.Errorf("rules must be a map[string]string")
	}
	for name, content := range rules {
		if name == "" {
			return fmt.Errorf("rule name cannot be empty")
		}
		if content == "" {
			return fmt.Errorf("rule content cannot be empty for rule '%s'", name)
		}
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
func (c *Config) BusName() string { 
	return c.EventBridgeBusName 
}

// Rules returns all configured EventBridge rules as name->content map
func (c *Config) Rules() map[string]string {
	return c.EventBridgeRulesMap
}

// PermissionStatementId returns the Lambda permission statement ID for EventBridge
func (c *Config) PermissionStatementId() string {
	return strings.Join([]string{"eventbridge", c.BusName(), c.resource.Name()}, "-")
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
