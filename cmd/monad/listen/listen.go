package listen

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/bkeane/monad/internal/logging"
	"github.com/bkeane/monad/pkg/handler"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/justinas/alice"
	"github.com/rs/zerolog/log"
)

type Root struct {
	Port int32 `arg:"-p,env:PORT" help:"port to listen on" default:"8080"`
}

func (r Root) Route(ctx context.Context, awsconfig aws.Config) (*string, error) {
	if StdinPresent() {
		return StdinHandler(ctx, awsconfig)
	}

	return PortHandler(ctx, awsconfig, r.Port)
}

func PortHandler(ctx context.Context, awsconfig aws.Config, port int32) (*string, error) {
	log.Info().Msgf("listening on port %d", port)
	return nil, http.ListenAndServe(fmt.Sprintf(":%d", port), Server(ctx, awsconfig))
}

func StdinHandler(ctx context.Context, awsconfig aws.Config) (*string, error) {
	stdin := bufio.NewReader(os.Stdin)
	req := httptest.NewRequest(http.MethodPost, "/events", stdin)
	rec := httptest.NewRecorder()

	handler := Server(ctx, awsconfig)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", rec.Code, rec.Body.String())
	}

	result := rec.Body.String()
	return &result, nil
}

func StdinPresent() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	return (stat.Mode()&os.ModeCharDevice) == 0 || stat.Size() > 0
}

func Server(ctx context.Context, awsconfig aws.Config) http.Handler {
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/", notFound)
	httpMux.HandleFunc("GET /health", healthCheck)
	httpMux.HandleFunc("POST /events", handler.Init(ctx, awsconfig).HttpMount)
	middleware := alice.New(logging.HTTP)
	return middleware.Then(httpMux)
}

func notFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not found", http.StatusNotFound)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("health check ok"))
}
