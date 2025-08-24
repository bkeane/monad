package lambda

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"

	"github.com/caarlos0/env/v11"
	v "github.com/go-ozzo/ozzo-validation/v4"
	dotenv "github.com/joho/godotenv"
)

type Basis interface {
	AwsConfig() aws.Config
	AccountId() string
	Name() string
	EnvTemplate() string
	Render(string) (string, error)
	Tags() map[string]string
}

//
// Convention
//

type Config struct {
	basis          Basis
	client         *lambda.Client
	RegionName     string `env:"MONAD_LAMBDA_REGION"`
	StorageSize    int32  `env:"MONAD_STORAGE" flag:"--disk" usage:"Ephemeral storage size in MB"`
	MemorySizeMB   int32  `env:"MONAD_MEMORY" flag:"--memory" usage:"Memory size in MB"`
	TimeoutSeconds int32  `env:"MONAD_TIMEOUT" flag:"--timeout" usage:"Function timeout in seconds"`
	RetryCount     int32  `env:"MONAD_RETRIES" flag:"--retry" usage:"Async function invoke retries"`
	EnvPath        string `env:"MONAD_ENV" flag:"--env" usage:"Environment template file path"`
	envTemplate    string
	envMap         map[string]string
}

//
// Derive
//

func Derive(ctx context.Context, basis Basis) (*Config, error) {
	var err error
	var cfg Config

	cfg.basis = basis
	cfg.client = lambda.NewFromConfig(basis.AwsConfig())

	if err = env.Parse(&cfg); err != nil {
		return nil, err
	}

	if cfg.RegionName == "" {
		cfg.RegionName = basis.AwsConfig().Region
	}

	if cfg.StorageSize == 0 {
		cfg.StorageSize = int32(512)
	}

	if cfg.MemorySizeMB == 0 {
		cfg.MemorySizeMB = int32(128)
	}

	if cfg.TimeoutSeconds == 0 {
		cfg.TimeoutSeconds = int32(3)
	}

	if cfg.RetryCount == 0 {
		cfg.RetryCount = int32(0)
	}

	// Env derivation
	if cfg.EnvPath == "" {
		cfg.envTemplate = basis.EnvTemplate()

	} else {
		bytes, err := os.ReadFile(cfg.EnvPath)
		if err != nil {
			return nil, err
		}
		cfg.envTemplate = string(bytes)
	}

	// Render the environment template before parsing
	renderedEnvTemplate, err := cfg.basis.Render(cfg.envTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to render env template: %w", err)
	}

	cfg.envMap, err = dotenv.Parse(strings.NewReader(renderedEnvTemplate))
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
		v.Field(&c.basis, v.Required),
		v.Field(&c.client, v.Required),
		v.Field(&c.RegionName, v.Required),
		v.Field(&c.StorageSize, v.Required),
		v.Field(&c.MemorySizeMB, v.Required),
		v.Field(&c.TimeoutSeconds, v.Required),
		v.Field(&c.RetryCount, v.Min(int32(0))),
	)
}

//
// Accessors
//

// Client returns the AWS Lambda service client
func (c *Config) Client() *lambda.Client { return c.client }

// Region returns the AWS region for Lambda deployment
func (c *Config) Region() string { return c.RegionName }

// Timeout returns the function timeout in seconds
func (c *Config) Timeout() int32 { return c.TimeoutSeconds }

// MemorySize returns the allocated memory in MB
func (c *Config) MemorySize() int32 { return c.MemorySizeMB }

// EphemeralStorage returns the ephemeral storage size in MB
func (c *Config) EphemeralStorage() int32 { return c.StorageSize }

// Retries returns the number of async invoke retries
func (c *Config) Retries() int32 { return c.RetryCount }

// FunctionName returns the Lambda function name using axiom resource naming
func (c *Config) FunctionName() string { return c.basis.Name() }

// FunctionArn returns the complete ARN for the Lambda function
func (c *Config) FunctionArn() string {
	return fmt.Sprintf("arn:aws:lambda:%s:%s:function:%s",
		c.RegionName, c.basis.AccountId(), c.FunctionName())
}

// Env returns a map derived from the given env document
func (c *Config) Env() map[string]string {
	return c.envMap
}

// Tags returns standardized Lambda resource tags
func (c *Config) Tags() map[string]string {
	return c.basis.Tags()
}
