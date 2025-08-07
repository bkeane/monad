package cli

import (
	"fmt"
	"reflect"
	"strings"
)

//
// Annotation Parser
//

type FieldBinding struct {
	Flag   string
	EnvVar string
	Desc   string
}

func ParseBindings(structType interface{}) ([]FieldBinding, error) {
	var bindings []FieldBinding
	
	t := reflect.TypeOf(structType)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		
		bindTag := field.Tag.Get("bind")
		descTag := field.Tag.Get("desc")
		
		if bindTag == "" {
			continue // Skip fields without bind tag
		}
		
		binding, err := parseBindTag(bindTag, descTag)
		if err != nil {
			return nil, fmt.Errorf("field %s: %w", field.Name, err)
		}
		
		bindings = append(bindings, binding)
	}
	
	return bindings, nil
}

func parseBindTag(bindTag, descTag string) (FieldBinding, error) {
	// Parse: bind="--owner,MONAD_OWNER"
	parts := strings.Split(bindTag, ",")
	if len(parts) != 2 {
		return FieldBinding{}, fmt.Errorf("bind tag must have format '--flag,ENV_VAR', got: %s", bindTag)
	}
	
	flag := strings.TrimSpace(parts[0])
	envVar := strings.TrimSpace(parts[1])
	
	// Validate flag format
	if !strings.HasPrefix(flag, "--") {
		return FieldBinding{}, fmt.Errorf("flag must start with '--', got: %s", flag)
	}
	
	// Validate env var format (basic check)
	if envVar == "" {
		return FieldBinding{}, fmt.Errorf("environment variable cannot be empty")
	}
	
	return FieldBinding{
		Flag:   flag,
		EnvVar: envVar,
		Desc:   descTag,
	}, nil
}

//
// Help Generation
//

func GenerateHelpText(structType interface{}) (string, error) {
	bindings, err := ParseBindings(structType)
	if err != nil {
		return "", err
	}
	
	var lines []string
	for _, binding := range bindings {
		// Format: "  --flag                 description [env: ENV_VAR]"
		flagPart := fmt.Sprintf("  %s", binding.Flag)
		envPart := fmt.Sprintf("[env: %s]", binding.EnvVar)
		
		// Pad flag to align descriptions
		paddedFlag := fmt.Sprintf("%-20s", flagPart)
		line := fmt.Sprintf("%s %s %s", paddedFlag, binding.Desc, envPart)
		lines = append(lines, line)
	}
	
	return strings.Join(lines, "\n"), nil
}

//
// Flag Resolution
//

func FlagToEnvVar(structType interface{}, flagName string) (string, bool) {
	bindings, err := ParseBindings(structType)
	if err != nil {
		return "", false
	}
	
	for _, binding := range bindings {
		if binding.Flag == flagName {
			return binding.EnvVar, true
		}
	}
	
	return "", false
}

func GetValidFlags(structType interface{}) ([]string, error) {
	bindings, err := ParseBindings(structType)
	if err != nil {
		return nil, err
	}
	
	var flags []string
	for _, binding := range bindings {
		flags = append(flags, binding.Flag)
	}
	
	return flags, nil
}