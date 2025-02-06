package logging

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/aws/smithy-go/logging"
	"github.com/google/uuid"
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

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func HTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrappedWriter := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		requestID := uuid.New().String()
		logger := zerolog.New(os.Stdout).With().Str("request_id", requestID).Logger()
		ctx := logger.WithContext(r.Context())

		logger.Debug().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("request")

		next.ServeHTTP(wrappedWriter, r.WithContext(ctx))

		logger.Debug().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", wrappedWriter.statusCode).
			Msg("response")
	})
}
