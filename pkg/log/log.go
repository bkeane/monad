package log

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

type CloudWatchConfig interface {
	Client() *cloudwatchlogs.Client
	Name() string
}

type Log struct {
	Tail   bool `env:"MONAD_LOG_TAIL" flag:"--tail,-f" usage:"Follow log output"`
	config CloudWatchConfig
	client *cloudwatchlogs.Client
}

func Derive(config CloudWatchConfig) *Log {
	return &Log{
		config: config,
		client: config.Client(),
	}
}

func (l *Log) Fetch(ctx context.Context) error {
	if l.Tail {
		return l.streamLive(ctx)
	}
	return l.getRecentLogs(ctx)
}

func (l *Log) streamLive(ctx context.Context) error {
	// Create a context that can be cancelled by signal
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	input := &cloudwatchlogs.StartLiveTailInput{
		LogGroupIdentifiers: []string{l.config.Name()},
	}

	response, err := l.client.StartLiveTail(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to start live tail: %w", err)
	}

	stream := response.GetStream()
	defer stream.Close()

	fmt.Printf("Starting live tail for log group: %s (press Ctrl+C to exit)\n", l.config.Name())

	// Use goroutine to handle stream events
	eventChan := make(chan types.StartLiveTailResponseStream, 100)
	errChan := make(chan error, 1)

	go func() {
		defer close(eventChan)
		for event := range stream.Events() {
			select {
			case eventChan <- event:
			case <-ctx.Done():
				return
			}
		}
		if err := stream.Err(); err != nil {
			errChan <- err
		}
	}()

	// Process events
	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nStopping live tail...")
			return nil
		case err := <-errChan:
			return fmt.Errorf("stream error: %w", err)
		case event, ok := <-eventChan:
			if !ok {
				return nil
			}

			switch e := event.(type) {
			case *types.StartLiveTailResponseStreamMemberSessionStart:
				// Session started successfully
			case *types.StartLiveTailResponseStreamMemberSessionUpdate:
				for _, logEvent := range e.Value.SessionResults {
					timestamp := time.Unix(*logEvent.Timestamp/1000, 0).Format("2006-01-02 15:04:05")
					message := ""
					if logEvent.Message != nil {
						message = *logEvent.Message
					}
					fmt.Printf("%s %s\n", timestamp, message)
				}
			}
		}
	}
}

func (l *Log) getRecentLogs(ctx context.Context) error {
	// Get logs from the last hour by default
	endTime := time.Now()
	startTime := endTime.Add(-1 * time.Hour)

	logGroupName := l.config.Name()
	startTimeMillis := startTime.UnixMilli()
	endTimeMillis := endTime.UnixMilli()

	input := &cloudwatchlogs.FilterLogEventsInput{
		LogGroupName: &logGroupName,
		StartTime:    &startTimeMillis,
		EndTime:      &endTimeMillis,
	}

	paginator := cloudwatchlogs.NewFilterLogEventsPaginator(l.client, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to get log events: %w", err)
		}

		for _, event := range page.Events {
			fmt.Printf("%s %s\n",
				time.Unix(*event.Timestamp/1000, 0).Format("2006-01-02 15:04:05"),
				*event.Message)
		}
	}

	return nil
}
