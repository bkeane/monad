# Flag Package

The `flag` package provides automatic CLI flag generation from struct field annotations using reflection. It integrates seamlessly with [urfave/cli v3](https://github.com/urfave/cli) to create type-safe command-line interfaces.

## Overview

This package allows you to define CLI flags declaratively using struct tags, following the same pattern as your existing `env:` tags. The `Parse()` function extracts these annotations and generates appropriate `cli.Flag` instances that can be used directly with urfave/cli commands.

## Supported Field Types

- `string` → `cli.StringFlag`
- `int`, `int32`, `int64` → `cli.IntFlag`
- `bool` → `cli.BoolFlag`
- `[]string` → `cli.StringSliceFlag`
- Pointer types (automatically unwrapped)

## Struct Tag Annotations

Use these struct tags to define flag behavior:

| Tag | Description | Required | Example |
|-----|-------------|----------|---------|
| `flag` | Flag name (with `--` prefix) | Yes | `flag:"--region"` |
| `usage` | Help text description | No | `usage:"AWS region"` |
| `env` | Environment variable name | No | `env:"MONAD_REGION"` |
| `default` | Default value | No | `default:"us-east-1"` |

### Special Values

- `flag:"-"` - Exclude field from flag generation
- No `flag` tag - Field is automatically skipped
- Unexported fields are always skipped

## Basic Usage

```go
package main

import (
    "context"
    "os"
    
    "github.com/bkeane/monad/pkg/flag"
    "github.com/urfave/cli/v3"
)

type Config struct {
    Region  string `env:"MONAD_REGION" flag:"--region" usage:"AWS region" default:"us-east-1"`
    Memory  int32  `env:"MONAD_MEMORY" flag:"--memory" usage:"Memory in MB" default:"128"`
    Verbose bool   `env:"MONAD_VERBOSE" flag:"--verbose" usage:"Enable verbose logging"`
    
    // This field will be skipped (no flag tag)
    Internal string `env:"MONAD_INTERNAL"`
    
    // This field will be excluded explicitly
    Client interface{} `flag:"-"`
}

func main() {
    config := &Config{}
    
    cmd := &cli.Command{
        Name:  "deploy",
        Usage: "Deploy a service",
        Flags: flag.Parse(config),
        Action: func(ctx context.Context, cmd *cli.Command) error {
            // Your command logic here
            return nil
        },
    }
    
    cmd.Run(context.Background(), os.Args)
}
```

## Integration with Existing Codebase

Since your structs already use `env:` tags for environment variable parsing, you can simply add `flag:` and `usage:` tags alongside them:

```go
// Before
type Lambda struct {
    region string `env:"MONAD_LAMBDA_REGION"`
    memory int32  `env:"MONAD_MEMORY"`
}

// After - Make fields exported and add flag annotations
type Lambda struct {
    Region string `env:"MONAD_LAMBDA_REGION" flag:"--region" usage:"AWS region"`
    Memory int32  `env:"MONAD_MEMORY" flag:"--memory" usage:"Memory in MB" default:"128"`
}
```

## Advanced Usage

### Multiple Commands with Different Flag Sets

```go
// Different structs for different commands
type DeployConfig struct {
    Region  string `flag:"--region" usage:"AWS region"`
    Memory  int32  `flag:"--memory" usage:"Memory in MB"`
    Timeout int32  `flag:"--timeout" usage:"Timeout in seconds"`
}

type ECRConfig struct {
    Region     string `flag:"--region" usage:"AWS region"`
    RegistryID string `flag:"--registry-id" usage:"ECR registry ID"`
}

func main() {
    cmd := &cli.Command{
        Name: "monad",
        Commands: []*cli.Command{
            {
                Name:  "deploy",
                Flags: flag.Parse(&DeployConfig{}),
                Action: deployAction,
            },
            {
                Name:  "ecr",
                Flags: flag.Parse(&ECRConfig{}),
                Action: ecrAction,
            },
        },
    }
}
```

### Working with Nested Structs

For complex configurations with nested structs, you can flatten them or parse specific sub-structs:

```go
type Config struct {
    Lambda *LambdaConfig
    IAM    *IAMConfig
}

// Parse specific sub-configurations
lambdaFlags := flag.Parse(&LambdaConfig{})
iamFlags := flag.Parse(&IAMConfig{})

// Combine flags
allFlags := append(lambdaFlags, iamFlags...)
```

## Environment Variable Integration

The package automatically integrates environment variables through urfave/cli's `ValueSourceChain`. When you specify an `env:` tag, the flag will:

1. Check command-line arguments first
2. Fall back to environment variable if present
3. Use default value if neither is provided

```bash
# These are equivalent:
./monad deploy --region us-west-2
MONAD_REGION=us-west-2 ./monad deploy
```

## Implementation Notes

- Uses Go's `reflect` package to inspect struct fields at runtime
- Only processes exported (capitalized) fields
- Maintains type safety through Go's type system
- Integrates with existing `caarlos0/env` parsing workflow
- Compatible with urfave/cli v3's `ValueSourceChain` architecture

## Error Handling

The package handles common edge cases gracefully:

- Non-struct types return empty flag slice
- Nil pointers return empty flag slice
- Unsupported field types log warnings and are skipped
- Invalid default values are ignored (zero values used instead)

## Testing

Run the test suite to verify functionality:

```bash
go test ./pkg/flag/
```

The tests cover various field types, edge cases, and integration scenarios.