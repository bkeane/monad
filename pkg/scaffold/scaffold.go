package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

type Basis interface {
	PolicyTemplate() string
	RoleTemplate() string
	EnvTemplate() string
}

type Scaffold struct {
	basis       Basis
	WritePolicy bool `env:"MONAD_SCAFFOLD_POLICY"`
	WriteRole   bool `env:"MONAD_SCAFFOLD_ROLE`
	WriteEnv    bool `env:"MONAD_SCAFFOLD_ENV`
}

func Derive(basis Basis) *Scaffold {
	return &Scaffold{basis: basis}
}

func (s *Scaffold) Create(language, targetDir string) error {
	if targetDir == "" {
		targetDir = "."
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
	return s.writeTemplate("policy.json.tmpl", s.basis.PolicyTemplate(), targetDir)
}

func (s *Scaffold) writeRole(targetDir string) error {
	return s.writeTemplate("role.json.tmpl", s.basis.RoleTemplate(), targetDir)
}

func (s *Scaffold) writeEnv(targetDir string) error {
	return s.writeTemplate(".env.tmpl", s.basis.EnvTemplate(), targetDir)
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
