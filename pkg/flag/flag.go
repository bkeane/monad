package flag

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/urfave/cli/v3"
)

// Global configuration for flag behavior
var (
	disableDefaults bool
)

// DisableDefaults globally disables default value setting and display across all flags.
// When called, flags will not have default values set and will not show defaults in help.
// This keeps your CLI clean while allowing your business logic to handle defaults independently.
func DisableDefaults() {
	disableDefaults = true
}

// EnableDefaults re-enables default value setting and display (default behavior).
func EnableDefaults() {
	disableDefaults = false
}

// Flags extracts flag definitions from a struct type using reflection.
// It recursively traverses nested structs to find all flag annotations.
// Returns a slice of cli.Flag that can be used with urfave/cli.
func Flags[T any]() []cli.Flag {
	visited := make(map[reflect.Type]bool)
	var zero T
	typ := reflect.TypeOf(zero)
	return parseType(typ, visited)
}

// parseType traverses a type definition recursively to extract flag annotations
func parseType(typ reflect.Type, visited map[reflect.Type]bool) []cli.Flag {
	var flags []cli.Flag
	
	if typ == nil {
		return flags
	}
	
	// Handle pointer to struct - get the underlying type
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	
	// Only process structs
	if typ.Kind() != reflect.Struct {
		return flags
	}
	
	// Prevent infinite recursion
	if visited[typ] {
		return flags
	}
	visited[typ] = true
	
	// Process all fields in the struct type
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}
		
		// Check if field has flag annotation
		flagTag := field.Tag.Get("flag")
		
		// If field has flag annotation, create flag
		if flagTag != "" && flagTag != "-" {
			usage := field.Tag.Get("usage")
			envVar := field.Tag.Get("env")
			defaultValue := field.Tag.Get("default")
			
			// Clean the flag name (remove -- prefix if present)
			cleanFlagName := strings.TrimPrefix(flagTag, "--")
			
			flag := createFlag(field.Type, cleanFlagName, usage, envVar, defaultValue)
			if flag != nil {
				flags = append(flags, flag)
			}
		}
		
		// Recursively process nested struct types
		fieldType := field.Type
		
		// Handle pointer to struct - traverse the pointed-to type
		if fieldType.Kind() == reflect.Ptr && fieldType.Elem().Kind() == reflect.Struct {
			nestedFlags := parseType(fieldType.Elem(), visited)
			flags = append(flags, nestedFlags...)
		} else if fieldType.Kind() == reflect.Struct {
			// Handle embedded struct - traverse the struct type
			nestedFlags := parseType(fieldType, visited)
			flags = append(flags, nestedFlags...)
		}
	}
	
	return flags
}

// createFlag creates the appropriate cli.Flag based on the field type
func createFlag(fieldType reflect.Type, name, usage, envVar, defaultValue string) cli.Flag {
	var sources cli.ValueSourceChain
	if envVar != "" {
		sources = cli.EnvVars(envVar)
	}
	
	switch fieldType.Kind() {
	case reflect.String:
		var value string
		var hideDefault bool
		if !disableDefaults && defaultValue != "" {
			value = defaultValue
		} else if disableDefaults {
			hideDefault = true
		}
		return &cli.StringFlag{
			Name:        name,
			Usage:       usage,
			Sources:     sources,
			Value:       value,
			HideDefault: hideDefault,
		}
		
	case reflect.Int, reflect.Int32, reflect.Int64:
		var defaultInt int
		var hideDefault bool
		if !disableDefaults && defaultValue != "" {
			if parsed, err := strconv.Atoi(defaultValue); err == nil {
				defaultInt = parsed
			}
		} else if disableDefaults {
			hideDefault = true
		}
		return &cli.IntFlag{
			Name:        name,
			Usage:       usage,
			Sources:     sources,
			Value:       defaultInt,
			HideDefault: hideDefault,
		}
		
	case reflect.Bool:
		var defaultBool bool
		var hideDefault bool
		if !disableDefaults && defaultValue != "" {
			if parsed, err := strconv.ParseBool(defaultValue); err == nil {
				defaultBool = parsed
			}
		} else if disableDefaults {
			hideDefault = true
		}
		return &cli.BoolFlag{
			Name:        name,
			Usage:       usage,
			Sources:     sources,
			Value:       defaultBool,
			HideDefault: hideDefault,
		}
		
	case reflect.Slice:
		// Handle string slices
		if fieldType.Elem().Kind() == reflect.String {
			var defaultSlice []string
			var hideDefault bool
			if !disableDefaults && defaultValue != "" {
				defaultSlice = strings.Split(defaultValue, ",")
			} else if disableDefaults {
				hideDefault = true
			}
			return &cli.StringSliceFlag{
				Name:        name,
				Usage:       usage,
				Sources:     sources,
				Value:       defaultSlice,
				HideDefault: hideDefault,
			}
		}
		
	case reflect.Ptr:
		// Handle pointer types by checking the underlying type
		return createFlag(fieldType.Elem(), name, usage, envVar, defaultValue)
	}
	
	// For unsupported types, return nil and log a warning
	fmt.Printf("Warning: unsupported field type %s for flag %s\n", fieldType.String(), name)
	return nil
}

// Before returns a BeforeFunc that exports all set CLI flags to corresponding environment variables.
// Uses the same type traversal as Flags to find flag-to-env mappings.
// Returns a function compatible with cli.Command.Before.
func Before[T any]() func(context.Context, *cli.Command) (context.Context, error) {
	return func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		visited := make(map[reflect.Type]bool)
		var zero T
		typ := reflect.TypeOf(zero)
		err := exportType(typ, visited, cmd)
		return ctx, err
	}
}

// Export reads all set CLI flags and exports their values to corresponding environment variables.
// Uses the same type traversal as Flags to find flag-to-env mappings.
// This function is kept for direct testing - use Before() for CLI integration.
func Export[T any](cmd *cli.Command) error {
	visited := make(map[reflect.Type]bool)
	var zero T
	typ := reflect.TypeOf(zero)
	return exportType(typ, visited, cmd)
}

// exportType traverses a type definition to find flag-to-env mappings and export set values
func exportType(typ reflect.Type, visited map[reflect.Type]bool, cmd *cli.Command) error {
	if typ == nil {
		return nil
	}
	
	// Handle pointer to struct - get the underlying type
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	
	// Only process structs
	if typ.Kind() != reflect.Struct {
		return nil
	}
	
	// Prevent infinite recursion
	if visited[typ] {
		return nil
	}
	visited[typ] = true
	
	// Process all fields in the struct type
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}
		
		// Check if field has flag annotation
		flagTag := field.Tag.Get("flag")
		envVar := field.Tag.Get("env")
		
		// If field has both flag and env annotations, check if flag is set
		if flagTag != "" && flagTag != "-" && envVar != "" {
			// Clean the flag name (remove -- prefix if present)
			cleanFlagName := strings.TrimPrefix(flagTag, "--")
			
			if cmd.IsSet(cleanFlagName) {
				value := getFlagValue(cmd, cleanFlagName, field.Type)
				if value != "" {
					if err := os.Setenv(envVar, value); err != nil {
						return fmt.Errorf("failed to export %s to %s: %w", cleanFlagName, envVar, err)
					}
				}
			}
		}
		
		// Recursively process nested struct types
		fieldType := field.Type
		
		// Handle pointer to struct - traverse the pointed-to type
		if fieldType.Kind() == reflect.Ptr && fieldType.Elem().Kind() == reflect.Struct {
			if err := exportType(fieldType.Elem(), visited, cmd); err != nil {
				return err
			}
		} else if fieldType.Kind() == reflect.Struct {
			// Handle embedded struct - traverse the struct type
			if err := exportType(fieldType, visited, cmd); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// getFlagValue retrieves the string representation of a flag value from the command
func getFlagValue(cmd *cli.Command, flagName string, fieldType reflect.Type) string {
	switch fieldType.Kind() {
	case reflect.String:
		return cmd.String(flagName)
	case reflect.Int, reflect.Int32, reflect.Int64:
		return strconv.Itoa(cmd.Int(flagName))
	case reflect.Bool:
		return strconv.FormatBool(cmd.Bool(flagName))
	case reflect.Slice:
		if fieldType.Elem().Kind() == reflect.String {
			return strings.Join(cmd.StringSlice(flagName), ",")
		}
	}
	return ""
}