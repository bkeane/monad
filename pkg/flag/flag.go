package flag

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/urfave/cli/v3"
)

// Parse extracts flag definitions from a struct using reflection
// and returns a slice of cli.Flag that can be used with urfave/cli
func Parse(v interface{}) []cli.Flag {
	var flags []cli.Flag
	
	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)
	
	// Handle pointer to struct
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}
	
	// Only process structs
	if val.Kind() != reflect.Struct {
		return flags
	}
	
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}
		
		// Check if field has flag annotation
		flagTag := field.Tag.Get("flag")
		if flagTag == "" || flagTag == "-" {
			continue
		}
		
		// Extract annotations
		usage := field.Tag.Get("usage")
		envVar := field.Tag.Get("env")
		defaultValue := field.Tag.Get("default")
		
		// Create appropriate flag type based on field type
		flag := createFlag(field.Type, flagTag, usage, envVar, defaultValue)
		if flag != nil {
			flags = append(flags, flag)
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
		return &cli.StringFlag{
			Name:    name,
			Usage:   usage,
			Sources: sources,
			Value:   defaultValue,
		}
		
	case reflect.Int, reflect.Int32, reflect.Int64:
		var defaultInt int
		if defaultValue != "" {
			if parsed, err := strconv.Atoi(defaultValue); err == nil {
				defaultInt = parsed
			}
		}
		return &cli.IntFlag{
			Name:    name,
			Usage:   usage,
			Sources: sources,
			Value:   defaultInt,
		}
		
	case reflect.Bool:
		var defaultBool bool
		if defaultValue != "" {
			if parsed, err := strconv.ParseBool(defaultValue); err == nil {
				defaultBool = parsed
			}
		}
		return &cli.BoolFlag{
			Name:    name,
			Usage:   usage,
			Sources: sources,
			Value:   defaultBool,
		}
		
	case reflect.Slice:
		// Handle string slices
		if fieldType.Elem().Kind() == reflect.String {
			var defaultSlice []string
			if defaultValue != "" {
				defaultSlice = strings.Split(defaultValue, ",")
			}
			return &cli.StringSliceFlag{
				Name:    name,
				Usage:   usage,
				Sources: sources,
				Value:   defaultSlice,
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