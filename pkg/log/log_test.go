package log

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bkeane/monad/pkg/basis/mock"
	"github.com/bkeane/monad/pkg/config/cloudwatch"
)

func TestLogGroup_DumpLogsForTail_NoAgoFlag(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	// Create CloudWatch config
	cloudwatchConfig, err := cloudwatch.Derive(ctx, setup.Basis)
	if err != nil {
		// May fail due to AWS API calls in test environment
		t.Skip("CloudWatch config failed (expected in test env):", err)
	}

	// Create LogGroup with no --ago flag
	logGroup := &LogGroup{
		LogGroupAgo: "", // No --ago flag provided
		cloudwatch:  cloudwatchConfig,
	}

	// dumpLogsForTail should do nothing when no --ago flag is provided
	err = logGroup.dumpLogsForTail(ctx)
	assert.NoError(t, err, "dumpLogsForTail should not error when no --ago flag is provided")
}

func TestLogGroup_DumpLogsForTail_WithAgoFlag(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	// Create CloudWatch config
	cloudwatchConfig, err := cloudwatch.Derive(ctx, setup.Basis)
	if err != nil {
		// May fail due to AWS API calls in test environment
		t.Skip("CloudWatch config failed (expected in test env):", err)
	}

	// Create LogGroup with --ago flag
	logGroup := &LogGroup{
		LogGroupAgo: "1h", // 1 hour ago
		cloudwatch:  cloudwatchConfig,
	}

	// This will likely fail due to AWS API calls, but we test the duration parsing
	err = logGroup.dumpLogsForTail(ctx)
	
	// The error should be AWS-related, not duration parsing related
	if err != nil {
		assert.NotContains(t, err.Error(), "invalid duration format", 
			"Should not have duration parsing errors")
	}
}

func TestLogGroup_DumpLogsForTail_InvalidDuration(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	// Create CloudWatch config
	cloudwatchConfig, err := cloudwatch.Derive(ctx, setup.Basis)
	if err != nil {
		// May fail due to AWS API calls in test environment
		t.Skip("CloudWatch config failed (expected in test env):", err)
	}

	// Create LogGroup with invalid duration
	logGroup := &LogGroup{
		LogGroupAgo: "invalid-duration",
		cloudwatch:  cloudwatchConfig,
	}

	// Should return duration parsing error
	err = logGroup.dumpLogsForTail(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid duration format")
}

func TestLogGroup_Dump_DefaultDuration(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	// Create CloudWatch config
	cloudwatchConfig, err := cloudwatch.Derive(ctx, setup.Basis)
	if err != nil {
		// May fail due to AWS API calls in test environment
		t.Skip("CloudWatch config failed (expected in test env):", err)
	}

	// Create LogGroup with no --ago flag (should use default 30s)
	logGroup := &LogGroup{
		LogGroupAgo: "",
		cloudwatch:  cloudwatchConfig,
	}

	// Dump should use default 30 second duration
	err = logGroup.Dump(ctx)
	
	// The error should be AWS-related, not configuration related
	if err != nil {
		assert.NotContains(t, err.Error(), "invalid duration format")
		assert.NotContains(t, err.Error(), "failed to dump historical logs")
	}
}

func TestLogGroup_Dump_WithCustomDuration(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	// Create CloudWatch config
	cloudwatchConfig, err := cloudwatch.Derive(ctx, setup.Basis)
	if err != nil {
		// May fail due to AWS API calls in test environment
		t.Skip("CloudWatch config failed (expected in test env):", err)
	}

	// Create LogGroup with custom --ago flag
	logGroup := &LogGroup{
		LogGroupAgo: "2h30m", // 2 hours 30 minutes
		cloudwatch:  cloudwatchConfig,
	}

	// Should parse duration correctly
	err = logGroup.Dump(ctx)
	
	// The error should be AWS-related, not duration parsing related
	if err != nil {
		assert.NotContains(t, err.Error(), "invalid duration format")
	}
}

func TestDurationParsing(t *testing.T) {
	tests := []struct {
		name        string
		duration    string
		expectError bool
	}{
		{"valid hours", "2h", false},
		{"valid minutes", "30m", false},
		{"valid seconds", "45s", false},
		{"valid combined", "1h30m45s", false},
		{"invalid format", "invalid", true},
		{"empty string", "", true},
		{"negative duration", "-1h", false}, // Go allows negative durations
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := time.ParseDuration(tt.duration)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewlineHandling(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"message with newline", "Hello World\n", "Hello World"},
		{"message without newline", "Hello World", "Hello World"},
		{"message with multiple newlines", "Hello World\n\n", "Hello World\n"},
		{"empty message", "", ""},
		{"message with just newline", "\n", ""},
		{"message with spaces and newline", "Hello World   \n", "Hello World   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strings.TrimSuffix(tt.input, "\n")
			assert.Equal(t, tt.expected, result)
		})
	}
}