package log

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/bkeane/monad/pkg/config/cloudwatch"
	"github.com/rs/zerolog/log"

	"github.com/caarlos0/env/v11"
)

type Config interface {
	Client() *cloudwatchlogs.Client
	Name() string
	Arn() string
}

type LogGroup struct {
	LogGroupTail bool   `env:"MONAD_LOG_TAIL" flag:"--tail,-f" usage:"Follow log output"`
	LogGroupAgo  string `env:"MONAD_LOG_AGO" flag:"--ago" usage:"Show logs from duration ago (e.g., 1h, 30m, 60s)" hint:"duration"`
	cloudwatch   *cloudwatch.Config
}

func Derive(config *cloudwatch.Config) (*LogGroup, error) {
	var lg LogGroup

	err := env.Parse(&lg)
	if err != nil {
		return nil, err
	}

	lg.cloudwatch = config

	return &lg, nil
}

func (l *LogGroup) Dump(ctx context.Context) error {
	// Parse duration or default to 30 seconds
	duration := 30 * time.Second
	if l.LogGroupAgo != "" {
		var err error
		duration, err = time.ParseDuration(l.LogGroupAgo)
		if err != nil {
			return fmt.Errorf("invalid duration format: %w", err)
		}
	}

	return l.dumpLogs(ctx, duration)
}

// dumpLogsForTail dumps historical logs for the tail command, respecting the --ago flag
func (l *LogGroup) dumpLogsForTail(ctx context.Context) error {
	// For tailing, we only dump historical logs if --ago flag is provided
	// If no --ago flag is provided, we start tailing from now (no historical logs)
	if l.LogGroupAgo == "" {
		return nil
	}

	duration, err := time.ParseDuration(l.LogGroupAgo)
	if err != nil {
		return fmt.Errorf("invalid duration format: %w", err)
	}

	return l.dumpLogs(ctx, duration)
}

// dumpLogs is the common implementation for dumping logs for a given duration
func (l *LogGroup) dumpLogs(ctx context.Context, duration time.Duration) error {
	// Get logs from the specified duration ago
	endTime := time.Now()
	startTime := endTime.Add(-duration)

	logGroupName := l.cloudwatch.Name()
	startTimeMillis := startTime.UnixMilli()
	endTimeMillis := endTime.UnixMilli()

	input := &cloudwatchlogs.FilterLogEventsInput{
		LogGroupName: &logGroupName,
		StartTime:    &startTimeMillis,
		EndTime:      &endTimeMillis,
	}

	paginator := cloudwatchlogs.NewFilterLogEventsPaginator(l.cloudwatch.Client(), input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to get log events: %w", err)
		}

		for _, event := range page.Events {
			message := strings.TrimSuffix(*event.Message, "\n")
			fmt.Printf("%s %s\n",
				time.Unix(*event.Timestamp/1000, 0).Format("2006-01-02 15:04:05"),
				message)
		}
	}

	return nil
}

func (l *LogGroup) session(ctx context.Context) (*cloudwatchlogs.StartLiveTailEventStream, error) {
	client := l.cloudwatch.Client()

	request := &cloudwatchlogs.StartLiveTailInput{
		LogGroupIdentifiers: []string{l.cloudwatch.Arn()},
	}

	response, err := client.StartLiveTail(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to start live tail: %w", err)
	}

	return response.GetStream(), nil
}

func (l *LogGroup) Tail(ctx context.Context) error {
	// First, dump existing logs from --ago duration before starting the live tail
	if err := l.dumpLogsForTail(ctx); err != nil {
		return fmt.Errorf("failed to dump historical logs: %w", err)
	}

	session, err := l.session(ctx)
	if err != nil {
		return err
	}
	defer session.Close()

	// Setup signal handling for Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	eventsChan := session.Events()
	for {
		select {
		case <-sigChan:
			log.Info().Msg("Stopping log tail...")
			return nil
		case event := <-eventsChan:
			switch e := event.(type) {
			case *types.StartLiveTailResponseStreamMemberSessionStart:
				// successfully started
			case *types.StartLiveTailResponseStreamMemberSessionUpdate:
				for _, logEvent := range e.Value.SessionResults {
					timestamp := time.Unix(*logEvent.Timestamp/1000, 0).Format("2006-01-02 15:04:05")
					message := ""
					if logEvent.Message != nil {
						message = strings.TrimSuffix(*logEvent.Message, "\n")
					}
					fmt.Printf("%s %s\n", timestamp, message)
				}
			default:
				if err := session.Err(); err != nil {
					return fmt.Errorf("stream error: %w", err)
				} else if event == nil {
					fmt.Println("Stream closed")
					return nil
				} else {
					return fmt.Errorf("unknown event type: %T", e)
				}
			}
		}
	}
}
