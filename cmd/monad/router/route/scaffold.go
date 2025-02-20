package route

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

//go:embed scaffolds/*
var templates embed.FS

type Scaffold struct {
	Scaffold string `arg:"positional,required" help:"go, python, node, ruby"`
	Name     string `arg:"positional,required" help:"function name"`
}

func (s *Scaffold) Route(ctx context.Context, r Root) error {
	scaffold := fmt.Sprintf("scaffolds/%s", s.Scaffold)
	if _, err := templates.Open(scaffold); err != nil {
		return fmt.Errorf("invalid scaffold type '%s'", s.Scaffold)
	}

	if err := os.MkdirAll(s.Name, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	err := fs.WalkDir(templates, scaffold, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath := strings.TrimPrefix(path, scaffold)
		if relPath == "" {
			return nil
		}

		destPath := filepath.Join(s.Name, relPath)

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
		return fmt.Errorf("failed to copy scaffold: %w", err)
	}

	log.Info().Msgf("Initialized new %s function in ./%s", s.Scaffold, s.Name)
	return nil
}
