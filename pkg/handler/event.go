package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bkeane/monad/pkg/config/release"
	"github.com/bkeane/monad/pkg/event"
	"github.com/bkeane/monad/pkg/saga"
	"github.com/bkeane/substrate/pkg/substrate"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/rs/zerolog"
)

type Handler struct {
	awsconfig aws.Config
	log       *zerolog.Logger
}

func Init(ctx context.Context, awsconfig aws.Config) *Handler {
	return &Handler{
		awsconfig: awsconfig,
		log:       zerolog.Ctx(ctx),
	}
}

func (h *Handler) Event(ctx context.Context, evt json.RawMessage) ([]byte, error) {
	h.log.Debug().Msg("event handler called")

	msg, substrate, err := substrate.Rx(ctx, h.awsconfig, evt)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal message from eventbridge event: %w", err)
	}

	switch msg := msg.(type) {
	case *event.DeployRequest:
		return h.DeployEvent(ctx, substrate, msg)

	case *event.DestroyRequest:
		return h.DestroyEvent(ctx, substrate, msg)

	default:
		return nil, fmt.Errorf("unsupported message type: %T", msg)

	}
}

func (h *Handler) DeployEvent(ctx context.Context, substrate *substrate.Substrate, msg *event.DeployRequest) ([]byte, error) {
	r, err := release.Config{}.Parse(ctx, h.awsconfig, msg.ImageUri, substrate)
	if err != nil {
		return nil, err
	}

	resp := &event.DeployResponse{
		Status:      "success",
		ImageUri:    r.ImageUri(),
		FunctionArn: r.FunctionArn(),
		PolicyArn:   r.PolicyArn(),
		RoleArn:     r.RoleArn(),
		EniRoleArn:  r.EniRoleArn(),
	}

	if err := saga.Init(ctx, r).Do(ctx); err != nil {
		resp.Status = "failure"
	}

	return json.Marshal(resp)
}

func (h *Handler) DestroyEvent(ctx context.Context, substrate *substrate.Substrate, msg *event.DestroyRequest) ([]byte, error) {
	r, err := release.Config{}.Parse(ctx, h.awsconfig, msg.ImageUri, substrate)
	if err != nil {
		return nil, err
	}

	resp := &event.DestroyResponse{
		Status:      "success",
		FunctionArn: r.FunctionArn(),
		PolicyArn:   r.PolicyArn(),
		RoleArn:     r.RoleArn(),
	}

	if err := saga.Init(ctx, r).Undo(ctx); err != nil {
		resp.Status = "failure"
	}

	return json.Marshal(resp)
}
