package scaffold

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

//go:embed embedded/*
var templates embed.FS

type Root struct {
	Scaffold string `arg:"positional,required" help:"go, python, node, ruby"`
	Name     string `arg:"positional,required" help:"function name"`
}

func (r *Root) Route(ctx context.Context) (*string, error) {
	scaffold := fmt.Sprintf("embedded/%s", r.Scaffold)
	if _, err := templates.Open(scaffold); err != nil {
		return nil, fmt.Errorf("invalid scaffold type '%s'", r.Scaffold)
	}

	if err := os.MkdirAll(r.Name, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	err := fs.WalkDir(templates, scaffold, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath := strings.TrimPrefix(path, scaffold)
		if relPath == "" {
			return nil
		}

		destPath := filepath.Join(r.Name, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		content, err := templates.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file: %w", err)
		}

		return os.WriteFile(destPath, content, 0644)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to copy scaffold: %w", err)
	}

	log.Info().Msgf("Initialized new %s function in ./%s", r.Scaffold, r.Name)
	return nil, nil
}
