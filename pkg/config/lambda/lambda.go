package lambda

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	v "github.com/go-ozzo/ozzo-validation/v4"
	dotenv "github.com/joho/godotenv"
)

type Basis interface {
	AwsConfig() aws.Config
	AccountId() string
	Name() string
	EnvTemplate() (string, error)
	Tags() map[string]string
}

//
// Convention
//

type Config struct {
	basis       Basis
	client      *lambda.Client
	region      string `env:"MONAD_LAMBDA_REGION"`
	storage     int32  `env:"MONAD_STORAGE"`
	memory      int32  `env:"MONAD_MEMORY"`
	timeout     int32  `env:"MONAD_TIMEOUT"`
	retries     int32  `env:"MONAD_RETRIES"`
	envPath     string `env:"MONAD_ENV"`
	envTemplate string
	envMap      map[string]string
}

//
// Derive
//

func Derive(ctx context.Context, basis Basis) (*Config, error) {
	var err error
	var cfg Config

	cfg.basis = basis
	cfg.client = lambda.NewFromConfig(basis.AwsConfig())

	if cfg.region == "" {
		cfg.region = basis.AwsConfig().Region
	}

	if cfg.storage == 0 {
		cfg.storage = int32(512)
	}

	if cfg.memory == 0 {
		cfg.memory = int32(128)
	}

	if cfg.timeout == 0 {
		cfg.timeout = int32(3)
	}

	if cfg.retries == 0 {
		cfg.retries = int32(0)
	}

	// Env derivation
	if cfg.envPath == "" {
		cfg.envTemplate, err = basis.EnvTemplate()
		if err != nil {
			return nil, err
		}
	} else {
		bytes, err := os.ReadFile(cfg.envPath)
		if err != nil {
			return nil, err
		}
		cfg.envTemplate = string(bytes)
	}

	cfg.envMap, err = dotenv.Parse(strings.NewReader(cfg.envTemplate))
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
		v.Field(&c.region, v.Required),
		v.Field(&c.storage, v.Required),
		v.Field(&c.memory, v.Required),
		v.Field(&c.timeout, v.Required),
		v.Field(&c.retries, v.Min(int32(0))),
	)
}

//
// Accessors
//

// Client returns the AWS Lambda service client
func (c *Config) Client() *lambda.Client { return c.client }

// Region returns the AWS region for Lambda deployment
func (c *Config) Region() string { return c.region }

// Timeout returns the function timeout in seconds
func (c *Config) Timeout() int32 { return c.timeout }

// MemorySize returns the allocated memory in MB
func (c *Config) MemorySize() int32 { return c.memory }

// EphemeralStorage returns the ephemeral storage size in MB
func (c *Config) EphemeralStorage() int32 { return c.storage }

// Retries returns the number of async invoke retries
func (c *Config) Retries() int32 { return c.retries }

// FunctionName returns the Lambda function name using axiom resource naming
func (c *Config) FunctionName() string { return c.basis.Name() }

// FunctionArn returns the complete ARN for the Lambda function
func (c *Config) FunctionArn() string {
	return fmt.Sprintf("arn:aws:lambda:%s:%s:function:%s",
		c.region, c.basis.AccountId(), c.FunctionName())
}

// Env returns a map derived from the given env document
func (c *Config) Env() map[string]string {
	return c.envMap
}

// Tags returns standardized Lambda resource tags
func (c *Config) Tags() map[string]string {
	return c.basis.Tags()
}
