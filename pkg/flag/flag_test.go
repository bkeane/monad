package flag

import (
	"context"
	"os"
	"testing"

	"github.com/urfave/cli/v3"
)

// Test struct with various field types and annotations
type TestConfig struct {
	// String field with all annotations
	Region string `env:"TEST_REGION" flag:"--region" usage:"AWS region" default:"us-east-1"`
	
	// Int field
	Memory int32 `env:"TEST_MEMORY" flag:"--memory" usage:"Memory in MB" default:"128"`
	
	// Int field with custom hint
	Retention int32 `env:"TEST_RETENTION" flag:"--retention" usage:"Log retention period" default:"7" hint:"days"`
	
	// Bool field
	Verbose bool `env:"TEST_VERBOSE" flag:"--verbose" usage:"Enable verbose logging" default:"false"`
	
	// String slice
	Tags []string `env:"TEST_TAGS" flag:"--tags" usage:"Resource tags" default:"env=test,app=monad"`
	
	// Field without flag annotation (should be skipped)
	Internal string `env:"TEST_INTERNAL"`
	
	// Field with flag:"-" (should be skipped)
	Client interface{} `env:"TEST_CLIENT" flag:"-"`
	
	// unexported field (should be skipped)
	private string `env:"TEST_PRIVATE" flag:"--private"`
}

func TestFlags(t *testing.T) {
	flags := Flags[TestConfig]()
	
	// Should have 5 flags (Region, Memory, Retention, Verbose, Tags)
	expectedCount := 5
	if len(flags) != expectedCount {
		t.Errorf("Expected %d flags, got %d", expectedCount, len(flags))
	}
	
	// Test each flag type
	flagMap := make(map[string]cli.Flag)
	for _, flag := range flags {
		switch f := flag.(type) {
		case *cli.StringFlag:
			flagMap[f.Name] = f
		case *CustomStringFlag:
			flagMap[f.Name] = f
		case *cli.IntFlag:
			flagMap[f.Name] = f
		case *CustomIntFlag:
			flagMap[f.Name] = f
		case *cli.BoolFlag:
			flagMap[f.Name] = f
		case *CustomBoolFlag:
			flagMap[f.Name] = f
		case *cli.StringSliceFlag:
			flagMap[f.Name] = f
		case *CustomStringSliceFlag:
			flagMap[f.Name] = f
		}
	}
	
	// Test string flag
	if regionFlag, ok := flagMap["region"].(*cli.StringFlag); ok {
		if regionFlag.Usage != "AWS region" {
			t.Errorf("Expected usage 'AWS region', got '%s'", regionFlag.Usage)
		}
		if regionFlag.Value != "us-east-1" {
			t.Errorf("Expected default 'us-east-1', got '%s'", regionFlag.Value)
		}
		envKeys := regionFlag.Sources.EnvKeys()
		if len(envKeys) != 1 || envKeys[0] != "TEST_REGION" {
			t.Errorf("Expected env var 'TEST_REGION', got %v", envKeys)
		}
	} else {
		t.Error("Expected region to be a StringFlag")
	}
	
	// Test int flag
	if memoryFlag, ok := flagMap["memory"].(*cli.IntFlag); ok {
		if memoryFlag.Value != 128 {
			t.Errorf("Expected default 128, got %d", memoryFlag.Value)
		}
	} else {
		t.Error("Expected memory to be an IntFlag")
	}
	
	// Test custom int flag with hint
	if retentionFlag, ok := flagMap["retention"].(*CustomIntFlag); ok {
		if retentionFlag.Value != 7 {
			t.Errorf("Expected default 7, got %d", retentionFlag.Value)
		}
		if retentionFlag.TypeName() != "days" {
			t.Errorf("Expected type name 'days', got '%s'", retentionFlag.TypeName())
		}
	} else {
		t.Error("Expected retention to be a CustomIntFlag")
	}
	
	// Test bool flag
	if verboseFlag, ok := flagMap["verbose"].(*cli.BoolFlag); ok {
		if verboseFlag.Value != false {
			t.Errorf("Expected default false, got %t", verboseFlag.Value)
		}
	} else {
		t.Error("Expected verbose to be a BoolFlag")
	}
	
	// Test string slice flag
	if tagsFlag, ok := flagMap["tags"].(*cli.StringSliceFlag); ok {
		expected := []string{"env=test", "app=monad"}
		if len(tagsFlag.Value) != len(expected) {
			t.Errorf("Expected %d tags, got %d", len(expected), len(tagsFlag.Value))
		}
		for i, tag := range tagsFlag.Value {
			if tag != expected[i] {
				t.Errorf("Expected tag '%s', got '%s'", expected[i], tag)
			}
		}
	} else {
		t.Error("Expected tags to be a StringSliceFlag")
	}
}

func TestParseNonStruct(t *testing.T) {
	// Test with non-struct type
	flags := Flags[string]()
	if len(flags) != 0 {
		t.Errorf("Expected 0 flags for non-struct, got %d", len(flags))
	}
}

func TestParseNil(t *testing.T) {
	// Test with nil - should still find flags from the type definition  
	flags := Flags[TestConfig]()
	
	// Should find flags from the type even if the pointer is nil
	expectedCount := 5 // Region, Memory, Retention, Verbose, Tags
	if len(flags) != expectedCount {
		t.Errorf("Expected %d flags for nil pointer (type-based), got %d", expectedCount, len(flags))
	}
}

// Test recursive parsing with nested structs
func TestParseRecursive(t *testing.T) {
	type NestedConfig struct {
		Host string `env:"TEST_HOST" flag:"--host" usage:"Server host" default:"localhost"`
		Port int    `env:"TEST_PORT" flag:"--port" usage:"Server port" default:"8080"`
	}
	
	type ParentConfig struct {
		Name   string        `env:"TEST_NAME" flag:"--name" usage:"Application name"`
		Nested *NestedConfig
	}
	
	flags := Flags[ParentConfig]()
	
	// Should find flags from both parent and nested struct
	expectedCount := 3 // --name, --host, --port
	if len(flags) != expectedCount {
		t.Errorf("Expected %d flags from recursive parsing, got %d", expectedCount, len(flags))
	}
	
	// Check that we got flags from both levels
	flagNames := make(map[string]bool)
	for _, flag := range flags {
		switch f := flag.(type) {
		case *cli.StringFlag:
			flagNames[f.Name] = true
		case *cli.IntFlag:
			flagNames[f.Name] = true
		}
	}
	
	expectedFlags := []string{"name", "host", "port"}
	for _, expectedFlag := range expectedFlags {
		if !flagNames[expectedFlag] {
			t.Errorf("Expected to find flag %s in recursive parsing", expectedFlag)
		}
	}
}

// Test that recursive parsing handles nil nested pointers (type-based)
func TestParseRecursiveNilNested(t *testing.T) {
	type NestedConfig struct {
		Value string `flag:"--nested-value" usage:"Nested value"`
	}
	
	type ParentConfig struct {
		Name   string        `flag:"--name" usage:"Application name"`
		Nested *NestedConfig // This will be nil but we parse the type
	}
	
	flags := Flags[ParentConfig]()
	
	// Should find both parent and nested flags (type-based parsing)
	expectedCount := 2 // --name and --nested-value
	if len(flags) != expectedCount {
		t.Errorf("Expected %d flags with nil nested (type-based), got %d", expectedCount, len(flags))
	}
	
	// Verify we got both flags
	flagNames := make(map[string]bool)
	for _, flag := range flags {
		if f, ok := flag.(*cli.StringFlag); ok {
			flagNames[f.Name] = true
		}
	}
	
	if !flagNames["name"] {
		t.Error("Expected to find name flag")
	}
	if !flagNames["nested-value"] {
		t.Error("Expected to find nested-value flag from type definition")
	}
}

// Test Export functionality
func TestExport(t *testing.T) {
	type ExportConfig struct {
		Name    string `env:"TEST_EXPORT_NAME" flag:"--name" usage:"Application name"`
		Port    int    `env:"TEST_EXPORT_PORT" flag:"--port" usage:"Server port"`
		Verbose bool   `env:"TEST_EXPORT_VERBOSE" flag:"--verbose" usage:"Verbose mode"`
	}
	
	// Clear any existing env vars
	os.Unsetenv("TEST_EXPORT_NAME")
	os.Unsetenv("TEST_EXPORT_PORT") 
	os.Unsetenv("TEST_EXPORT_VERBOSE")
	
	// Create a CLI command with our flags
	cmd := &cli.Command{
		Name:  "test",
		Flags: Flags[ExportConfig](),
		Action: func(ctx context.Context, c *cli.Command) error {
			// Export flags to env vars
			return Export[ExportConfig](c)
		},
	}
	
	// Simulate running with flags set
	args := []string{"test", "--name", "myapp", "--port", "8080", "--verbose"}
	err := cmd.Run(context.Background(), args)
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}
	
	// Check that env vars were set correctly
	if name := os.Getenv("TEST_EXPORT_NAME"); name != "myapp" {
		t.Errorf("Expected TEST_EXPORT_NAME=myapp, got %s", name)
	}
	
	if port := os.Getenv("TEST_EXPORT_PORT"); port != "8080" {
		t.Errorf("Expected TEST_EXPORT_PORT=8080, got %s", port)
	}
	
	if verbose := os.Getenv("TEST_EXPORT_VERBOSE"); verbose != "true" {
		t.Errorf("Expected TEST_EXPORT_VERBOSE=true, got %s", verbose)
	}
	
	// Clean up
	os.Unsetenv("TEST_EXPORT_NAME")
	os.Unsetenv("TEST_EXPORT_PORT")
	os.Unsetenv("TEST_EXPORT_VERBOSE")
}

// Test Export with nested structs
func TestExportNested(t *testing.T) {
	type NestedConfig struct {
		Host string `env:"TEST_NESTED_HOST" flag:"--host" usage:"Server host"`
	}
	
	type ParentConfig struct {
		Name   string         `env:"TEST_PARENT_NAME" flag:"--name" usage:"App name"`
		Nested *NestedConfig
	}
	
	// Clear env vars
	os.Unsetenv("TEST_PARENT_NAME")
	os.Unsetenv("TEST_NESTED_HOST")
	
	cmd := &cli.Command{
		Name:  "test",
		Flags: Flags[ParentConfig](),
		Action: func(ctx context.Context, c *cli.Command) error {
			return Export[ParentConfig](c)
		},
	}
	
	args := []string{"test", "--name", "testapp", "--host", "localhost"}
	err := cmd.Run(context.Background(), args)
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}
	
	// Check both parent and nested env vars were set
	if name := os.Getenv("TEST_PARENT_NAME"); name != "testapp" {
		t.Errorf("Expected TEST_PARENT_NAME=testapp, got %s", name)
	}
	
	if host := os.Getenv("TEST_NESTED_HOST"); host != "localhost" {
		t.Errorf("Expected TEST_NESTED_HOST=localhost, got %s", host)
	}
	
	// Clean up
	os.Unsetenv("TEST_PARENT_NAME")
	os.Unsetenv("TEST_NESTED_HOST")
}

// Test Export only affects flags that are actually set
func TestExportOnlySetFlags(t *testing.T) {
	type SelectiveConfig struct {
		SetFlag   string `env:"TEST_SET_FLAG" flag:"--set-flag" usage:"This will be set"`
		UnsetFlag string `env:"TEST_UNSET_FLAG" flag:"--unset-flag" usage:"This won't be set"`
	}
	
	// Clear and set a baseline
	os.Unsetenv("TEST_SET_FLAG")
	os.Unsetenv("TEST_UNSET_FLAG")
	os.Setenv("TEST_UNSET_FLAG", "original")
	
	cmd := &cli.Command{
		Name:  "test",
		Flags: Flags[SelectiveConfig](),
		Action: func(ctx context.Context, c *cli.Command) error {
			return Export[SelectiveConfig](c)
		},
	}
	
	// Only set one flag
	args := []string{"test", "--set-flag", "new_value"}
	err := cmd.Run(context.Background(), args)
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}
	
	// Only the set flag should be exported
	if setFlag := os.Getenv("TEST_SET_FLAG"); setFlag != "new_value" {
		t.Errorf("Expected TEST_SET_FLAG=new_value, got %s", setFlag)
	}
	
	// The unset flag should retain its original value
	if unsetFlag := os.Getenv("TEST_UNSET_FLAG"); unsetFlag != "original" {
		t.Errorf("Expected TEST_UNSET_FLAG=original (unchanged), got %s", unsetFlag)
	}
	
	// Clean up
	os.Unsetenv("TEST_SET_FLAG")
	os.Unsetenv("TEST_UNSET_FLAG")
}