package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/bkeane/monad/pkg/basis/defaults"
	"github.com/caarlos0/env/v11"
	"github.com/rs/zerolog/log"
)

type Basis interface {
	Defaults() (*defaults.Basis, error)
}

type Scaffold struct {
	WritePolicy bool `env:"MONAD_SCAFFOLD_POLICY"`
	WriteRole   bool `env:"MONAD_SCAFFOLD_ROLE"`
	WriteEnv    bool `env:"MONAD_SCAFFOLD_ENV"`
	defaults    *defaults.Basis
}

func Derive(basis Basis) (*Scaffold, error) {
	var s Scaffold

	// Parse environment variables into struct fields
	if err := env.Parse(&s); err != nil {
		return nil, err
	}

	defaults, err := basis.Defaults()
	if err != nil {
		return nil, err
	}

	s.defaults = defaults
	return &s, nil
}

func (s *Scaffold) List() ([]string, error) {
	entries, err := Templates.ReadDir("templates")
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}

	var languages []string
	for _, entry := range entries {
		if entry.IsDir() {
			languages = append(languages, entry.Name())
		}
	}

	return languages, nil
}

func (s *Scaffold) Create(language, targetDir string) error {
	if targetDir == "" {
		targetDir = "."
	}

	languages, err := s.List()
	if err != nil {
		return err
	}

	if !slices.Contains(languages, language) {
		log.Error().Strs("valid", languages).Str("given", language).Msg("scaffold")
		return fmt.Errorf("invalid language type '%s'", language)
	}

	scaffoldPath := filepath.Join("templates", language)
	if _, err := Templates.Open(scaffoldPath); err != nil {
		return fmt.Errorf("invalid scaffold type '%s'", language)
	}

	if err := s.copyScaffold(scaffoldPath, targetDir); err != nil {
		return err
	}

	if s.WritePolicy {
		if err := s.writePolicy(targetDir); err != nil {
			return err
		}
	}

	if s.WriteRole {
		if err := s.writeRole(targetDir); err != nil {
			return err
		}
	}

	if s.WriteEnv {
		if err := s.writeEnv(targetDir); err != nil {
			return err
		}
	}

	log.Info().Msgf("initialized new %s scaffold", language)
	return nil
}

func (s *Scaffold) copyScaffold(scaffoldPath, targetDir string) error {
	// Ensure target directory exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	return fs.WalkDir(Templates, scaffoldPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath := strings.TrimPrefix(path, scaffoldPath)
		if relPath == "" || d.IsDir() {
			return nil
		}

		destPath := filepath.Join(targetDir, filepath.Base(relPath))
		if _, err := os.Stat(destPath); err == nil {
			log.Info().Str("file", destPath).Msg("skipping existing file")
			return nil
		}

		content, err := Templates.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file: %w", err)
		}

		log.Info().Str("file", destPath).Msg("creating")
		return os.WriteFile(destPath, content, 0644)
	})
}

func (s *Scaffold) writePolicy(targetDir string) error {
	return s.writeTemplate("policy.json.tmpl", s.defaults.PolicyTemplate(), targetDir)
}

func (s *Scaffold) writeRole(targetDir string) error {
	return s.writeTemplate("role.json.tmpl", s.defaults.RoleTemplate(), targetDir)
}

func (s *Scaffold) writeEnv(targetDir string) error {
	return s.writeTemplate(".env.tmpl", s.defaults.EnvTemplate(), targetDir)
}

func (s *Scaffold) writeTemplate(filename, content, targetDir string) error {
	if targetDir == "" {
		targetDir = "."
	}
	filepath := filepath.Join(targetDir, filename)
	if _, err := os.Stat(filepath); err == nil {
		log.Info().Str("file", filepath).Msg("skipping existing file")
		return nil
	}

	log.Info().Str("file", filepath).Msg("creating")
	return os.WriteFile(filepath, []byte(content), 0644)
}
