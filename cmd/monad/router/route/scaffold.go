package route

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/bkeane/monad/pkg/param"

	"github.com/rs/zerolog/log"
)

type Scaffold struct {
	Scaffold string `arg:"positional,required" help:"go, python, node, ruby, shell"`
	Policy   bool   `arg:"--policy" help:"create policy.json.tmpl"`
	Role     bool   `arg:"--role" help:"create role.json.tmpl"`
	Env      bool   `arg:"--env" help:"create .env.tmpl"`
}

func (s *Scaffold) Route(ctx context.Context, r Root) error {
	scaffoldPath := filepath.Join("scaffolds", s.Scaffold)
	if _, err := param.Scaffolds.Open(scaffoldPath); err != nil {
		return fmt.Errorf("invalid scaffold type '%s'", s.Scaffold)
	}

	err := fs.WalkDir(param.Scaffolds, scaffoldPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath := strings.TrimPrefix(path, scaffoldPath)
		if relPath == "" {
			return nil
		}

		destPath := filepath.Base(relPath)

		if d.IsDir() {
			return nil
		}

		if _, err := os.Stat(destPath); err == nil {
			log.Info().Str("file", destPath).Msgf("skipping")
			return nil
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("failed to check file existence: %w", err)
		}

		content, err := param.Scaffolds.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file: %w", err)
		}

		return os.WriteFile(destPath, content, 0644)
	})

	writeDefault := func(name string) error {
		defaultPath := filepath.Join("defaults", name)
		policy, err := param.Defaults.ReadFile(defaultPath)
		if err != nil {
			return fmt.Errorf("failed to open policy template: %w", err)
		}

		if _, err := os.Stat(name); err == nil {
			log.Info().Str("file", name).Msgf("skipping")
			return nil
		}

		return os.WriteFile(name, policy, 0644)
	}

	if s.Policy {
		if err := writeDefault("policy.json.tmpl"); err != nil {
			return fmt.Errorf("failed to write policy template: %w", err)
		}
	}

	if s.Role {
		if err := writeDefault("role.json.tmpl"); err != nil {
			return fmt.Errorf("failed to write role template: %w", err)
		}
	}

	if s.Env {
		if err := writeDefault(".env.tmpl"); err != nil {
			return fmt.Errorf("failed to write env template: %w", err)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to copy scaffold: %w", err)
	}

	log.Info().Msgf("initialized new %s scaffold in current directory", s.Scaffold)
	return nil
}
