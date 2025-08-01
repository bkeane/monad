package logging

import (
	"context"
	"strings"

	"github.com/aws/smithy-go/logging"
	"github.com/rs/zerolog"
)

type RetryLogger struct {
	Log *zerolog.Logger
}

func AwsConfig(ctx context.Context) logging.Logger {
	return &RetryLogger{
		Log: zerolog.Ctx(ctx),
	}
}

func (l *RetryLogger) WithContext(ctx context.Context) logging.Logger {
	return &RetryLogger{
		Log: zerolog.Ctx(ctx),
	}
}

func (l *RetryLogger) Logf(classification logging.Classification, format string, v ...interface{}) {
	switch classification {
	case logging.Warn:
		l.Log.Warn().Msgf(format, v...)
	case logging.Debug:
		if strings.Contains(format, "retrying request") {
			l.Log.Debug().Msgf(format, v...)
		} else {
			l.Log.Trace().Msgf(format, v...)
		}
	default:
		l.Log.Error().Msgf(format, v...)
	}
}
