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

// CustomIntFlag embeds cli.IntFlag and allows overriding the type name
type CustomIntFlag struct {
	*cli.IntFlag
	typeName string
}

// TypeName returns the custom type name if set, otherwise falls back to default
func (f *CustomIntFlag) TypeName() string {
	if f.typeName != "" {
		return f.typeName
	}
	return f.IntFlag.TypeName()
}

// String overrides the string representation to use custom type name
func (f *CustomIntFlag) String() string {
	if f.typeName != "" {
		original := f.IntFlag.String()
		// Replace " int\t" with our custom type name
		return strings.ReplaceAll(original, " int\t", " "+f.typeName+"\t")
	}
	return f.IntFlag.String()
}

// CustomStringFlag embeds cli.StringFlag and allows overriding the type name
type CustomStringFlag struct {
	*cli.StringFlag
	typeName string
}

// TypeName returns the custom type name if set, otherwise falls back to default
func (f *CustomStringFlag) TypeName() string {
	if f.typeName != "" {
		return f.typeName
	}
	return f.StringFlag.TypeName()
}

// String overrides the string representation to use custom type name
func (f *CustomStringFlag) String() string {
	if f.typeName != "" {
		original := f.StringFlag.String()
		// Replace " string\t" with our custom type name
		return strings.ReplaceAll(original, " string\t", " "+f.typeName+"\t")
	}
	return f.StringFlag.String()
}

// CustomBoolFlag embeds cli.BoolFlag and allows overriding the type name
type CustomBoolFlag struct {
	*cli.BoolFlag
	typeName string
}

// TypeName returns the custom type name if set, otherwise falls back to default
func (f *CustomBoolFlag) TypeName() string {
	if f.typeName != "" {
		return f.typeName
	}
	return f.BoolFlag.TypeName()
}

// CustomStringSliceFlag embeds cli.StringSliceFlag and allows overriding the type name
type CustomStringSliceFlag struct {
	*cli.StringSliceFlag
	typeName string
}

// TypeName returns the custom type name if set, otherwise falls back to default
func (f *CustomStringSliceFlag) TypeName() string {
	if f.typeName != "" {
		return f.typeName
	}
	return f.StringSliceFlag.TypeName()
}

// String overrides the string representation to use custom type name
func (f *CustomStringSliceFlag) String() string {
	if f.typeName != "" {
		original := f.StringSliceFlag.String()
		// Replace " string " or " string\t" with our custom type name for slice flags
		result := strings.ReplaceAll(original, " string\t", " "+f.typeName+"\t")
		result = strings.ReplaceAll(result, " string ", " "+f.typeName+" ")
		return result
	}
	return f.StringSliceFlag.String()
}

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
			hint := field.Tag.Get("hint")
			
			// Parse flag name and aliases (comma-separated)
			flagParts := strings.Split(flagTag, ",")
			var cleanFlagName string
			var aliases []string
			
			for i, part := range flagParts {
				cleanPart := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(part, "--"), "-"))
				if i == 0 {
					cleanFlagName = cleanPart
				} else {
					aliases = append(aliases, cleanPart)
				}
			}
			
			flag := createFlag(field.Type, cleanFlagName, usage, envVar, defaultValue, hint, aliases)
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
func createFlag(fieldType reflect.Type, name, usage, envVar, defaultValue, hint string, aliases []string) cli.Flag {
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
		
		baseFlag := &cli.StringFlag{
			Name:        name,
			Aliases:     aliases,
			Usage:       usage,
			Sources:     sources,
			Value:       value,
			HideDefault: hideDefault,
		}
		
		// Use custom flag if hint is provided
		if hint != "" {
			return &CustomStringFlag{
				StringFlag: baseFlag,
				typeName:   hint,
			}
		}
		return baseFlag
		
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
		
		baseFlag := &cli.IntFlag{
			Name:        name,
			Aliases:     aliases,
			Usage:       usage,
			Sources:     sources,
			Value:       defaultInt,
			HideDefault: hideDefault,
		}
		
		// Use custom flag if hint is provided
		if hint != "" {
			return &CustomIntFlag{
				IntFlag:  baseFlag,
				typeName: hint,
			}
		}
		return baseFlag
		
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
		
		baseFlag := &cli.BoolFlag{
			Name:        name,
			Aliases:     aliases,
			Usage:       usage,
			Sources:     sources,
			Value:       defaultBool,
			HideDefault: hideDefault,
		}
		
		// Use custom flag if hint is provided
		if hint != "" {
			return &CustomBoolFlag{
				BoolFlag: baseFlag,
				typeName: hint,
			}
		}
		return baseFlag
		
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
			
			baseFlag := &cli.StringSliceFlag{
				Name:        name,
				Aliases:     aliases,
				Usage:       usage,
				Sources:     sources,
				Value:       defaultSlice,
				HideDefault: hideDefault,
			}
			
			// Use custom flag if hint is provided
			if hint != "" {
				return &CustomStringSliceFlag{
					StringSliceFlag: baseFlag,
					typeName:        hint,
				}
			}
			return baseFlag
		}
		
	case reflect.Ptr:
		// Handle pointer types by checking the underlying type
		return createFlag(fieldType.Elem(), name, usage, envVar, defaultValue, hint, aliases)
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
			// Parse the flag name (handle aliases by taking first name)
			flagNames := strings.Split(flagTag, ",")
			cleanFlagName := strings.TrimPrefix(strings.TrimSpace(flagNames[0]), "--")
			
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