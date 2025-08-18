package flag

import (
	"testing"

	"github.com/urfave/cli/v3"
)

// Test struct with various field types and annotations
type TestConfig struct {
	// String field with all annotations
	Region string `env:"TEST_REGION" flag:"--region" usage:"AWS region" default:"us-east-1"`
	
	// Int field
	Memory int32 `env:"TEST_MEMORY" flag:"--memory" usage:"Memory in MB" default:"128"`
	
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

func TestParse(t *testing.T) {
	config := &TestConfig{}
	flags := Parse(config)
	
	// Should have 4 flags (Region, Memory, Verbose, Tags)
	expectedCount := 4
	if len(flags) != expectedCount {
		t.Errorf("Expected %d flags, got %d", expectedCount, len(flags))
	}
	
	// Test each flag type
	flagMap := make(map[string]cli.Flag)
	for _, flag := range flags {
		switch f := flag.(type) {
		case *cli.StringFlag:
			flagMap[f.Name] = f
		case *cli.IntFlag:
			flagMap[f.Name] = f
		case *cli.BoolFlag:
			flagMap[f.Name] = f
		case *cli.StringSliceFlag:
			flagMap[f.Name] = f
		}
	}
	
	// Test string flag
	if regionFlag, ok := flagMap["--region"].(*cli.StringFlag); ok {
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
		t.Error("Expected --region to be a StringFlag")
	}
	
	// Test int flag
	if memoryFlag, ok := flagMap["--memory"].(*cli.IntFlag); ok {
		if memoryFlag.Value != 128 {
			t.Errorf("Expected default 128, got %d", memoryFlag.Value)
		}
	} else {
		t.Error("Expected --memory to be an IntFlag")
	}
	
	// Test bool flag
	if verboseFlag, ok := flagMap["--verbose"].(*cli.BoolFlag); ok {
		if verboseFlag.Value != false {
			t.Errorf("Expected default false, got %t", verboseFlag.Value)
		}
	} else {
		t.Error("Expected --verbose to be a BoolFlag")
	}
	
	// Test string slice flag
	if tagsFlag, ok := flagMap["--tags"].(*cli.StringSliceFlag); ok {
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
		t.Error("Expected --tags to be a StringSliceFlag")
	}
}

func TestParseNonStruct(t *testing.T) {
	// Test with non-struct type
	flags := Parse("not a struct")
	if len(flags) != 0 {
		t.Errorf("Expected 0 flags for non-struct, got %d", len(flags))
	}
}

func TestParseNil(t *testing.T) {
	// Test with nil
	var nilConfig *TestConfig
	flags := Parse(nilConfig)
	if len(flags) != 0 {
		t.Errorf("Expected 0 flags for nil pointer, got %d", len(flags))
	}
}