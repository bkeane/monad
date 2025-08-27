package lambda

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/bkeane/monad/pkg/basis/caller"
	"github.com/bkeane/monad/pkg/basis/defaults"
	"github.com/bkeane/monad/pkg/basis/resource"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/caarlos0/env/v11"
	v "github.com/go-ozzo/ozzo-validation/v4"
	dotenv "github.com/joho/godotenv"
)

type Basis interface {
	Caller() (*caller.Basis, error)
	Resource() (*resource.Basis, error)
	Defaults() (*defaults.Basis, error)
	Render(string) (string, error)
}

//
// Convention
//

type Config struct {
	client        *lambda.Client
	LambdaRegion  string `env:"MONAD_LAMBDA_REGION"`
	LambdaStorage int32  `env:"MONAD_STORAGE" flag:"--disk" usage:"Lambda storage" hint:"mb"`
	LambdaMemory  int32  `env:"MONAD_MEMORY" flag:"--memory" usage:"Lambda memory" hint:"mb"`
	LambdaTimeout int32  `env:"MONAD_TIMEOUT" flag:"--timeout" usage:"Lambda timeout" hint:"sec"`
	LambdaRetries int32  `env:"MONAD_RETRIES" flag:"--retry" usage:"Lambda async invoke retries" hint:"count"`
	LambdaEnvPath string `env:"MONAD_ENV" flag:"--env" usage:"Lambda env template file path" hint:"path"`
	caller        *caller.Basis
	defaults      *defaults.Basis
	resource      *resource.Basis
	env           map[string]string
}

//
// Derive
//

func Derive(ctx context.Context, basis Basis) (*Config, error) {
	var err error
	var cfg Config

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

	cfg.client = lambda.NewFromConfig(cfg.caller.AwsConfig())

	if err = env.Parse(&cfg); err != nil {
		return nil, err
	}

	if cfg.LambdaRegion == "" {
		cfg.LambdaRegion = cfg.caller.AwsConfig().Region
	}

	if cfg.LambdaStorage == 0 {
		cfg.LambdaStorage = int32(512)
	}

	if cfg.LambdaMemory == 0 {
		cfg.LambdaMemory = int32(128)
	}

	if cfg.LambdaTimeout == 0 {
		cfg.LambdaTimeout = int32(3)
	}

	if cfg.LambdaRetries == 0 {
		cfg.LambdaRetries = int32(0)
	}

	// Env derivation
	var envTemplate string

	if cfg.LambdaEnvPath == "" {
		envTemplate = cfg.defaults.EnvTemplate()

	} else {
		bytes, err := os.ReadFile(cfg.LambdaEnvPath)
		if err != nil {
			return nil, err
		}
		envTemplate = string(bytes)
	}

	templated, err := basis.Render(envTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to render env template: %w", err)
	}

	cfg.env, err = dotenv.Parse(strings.NewReader(templated))
	if err != nil {
		return nil, err
	}

	if err = cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

//
// Validate
//

func (c *Config) Validate() error {
	return v.ValidateStruct(c,
		v.Field(&c.client, v.Required),
		v.Field(&c.LambdaRegion, v.Required),
		v.Field(&c.LambdaStorage, v.Required),
		v.Field(&c.LambdaMemory, v.Required),
		v.Field(&c.LambdaTimeout, v.Required),
		v.Field(&c.LambdaRetries, v.Min(int32(0))),
		v.Field(&c.env, v.Required),
	)
}

//
// Accessors
//

// Client returns the AWS Lambda service client
func (c *Config) Client() *lambda.Client { return c.client }

// Region returns the AWS region for Lambda deployment
func (c *Config) Region() string { return c.LambdaRegion }

// Timeout returns the function timeout in seconds
func (c *Config) Timeout() int32 { return c.LambdaTimeout }

// MemorySize returns the allocated memory in MB
func (c *Config) MemorySize() int32 { return c.LambdaMemory }

// EphemeralStorage returns the ephemeral storage size in MB
func (c *Config) EphemeralStorage() int32 { return c.LambdaStorage }

// Retries returns the number of async invoke retries
func (c *Config) Retries() int32 { return c.LambdaRetries }

// FunctionName returns the Lambda function name using basis resource naming
func (c *Config) FunctionName() string {
	return c.resource.Name()
}

// FunctionArn returns the complete ARN for the Lambda function
func (c *Config) FunctionArn() string {
	return fmt.Sprintf("arn:aws:lambda:%s:%s:function:%s",
		c.LambdaRegion, c.caller.AccountId(), c.FunctionName())
}

// Env returns a map derived from the given env document
func (c *Config) Env() map[string]string {
	return c.env
}

// Tags returns standardized Lambda resource tags
func (c *Config) Tags() map[string]string {
	return c.resource.Tags()
}
