package pkg

import (
	"context"

	"github.com/bkeane/monad/pkg/basis"
	"github.com/bkeane/monad/pkg/config"
	"github.com/bkeane/monad/pkg/log"
	"github.com/bkeane/monad/pkg/registry"
	"github.com/bkeane/monad/pkg/saga"
	"github.com/bkeane/monad/pkg/scaffold"
	"github.com/bkeane/monad/pkg/state"
	"github.com/bkeane/monad/pkg/step"
)

func Basis(ctx context.Context) (*basis.Basis, error) {
	return basis.Derive(ctx)
}

func Config(ctx context.Context) (*config.Config, error) {
	basis, err := Basis(ctx)
	if err != nil {
		return nil, err
	}

	return config.Derive(ctx, basis)
}

func Saga(ctx context.Context) (*saga.Saga, error) {
	config, err := Config(ctx)
	if err != nil {
		return nil, err
	}

	steps, err := step.Derive(ctx, config)
	if err != nil {
		return nil, err
	}

	return saga.Derive(ctx, steps), nil
}

func Scaffold(ctx context.Context) (*scaffold.Scaffold, error) {
	basis, err := Basis(ctx)
	if err != nil {
		return nil, err
	}

	return scaffold.Derive(basis)
}

func Registry(ctx context.Context) (*registry.Client, error) {
	config, err := Config(ctx)
	if err != nil {
		return nil, err
	}

	ecrConfig, err := config.Ecr(ctx)
	if err != nil {
		return nil, err
	}

	return registry.Derive(ecrConfig), nil
}

func State(ctx context.Context) (*state.State, error) {
	basis, err := Basis(ctx)
	if err != nil {
		return nil, err
	}

	return state.Init(ctx, basis)
}

func Log(ctx context.Context) (*log.LogGroup, error) {
	config, err := Config(ctx)
	if err != nil {
		return nil, err
	}

	cloudwatchConfig, err := config.CloudWatch(ctx)
	if err != nil {
		return nil, err
	}

	return log.Derive(cloudwatchConfig)
}
