package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/bkeane/monad/pkg/basis"
	"github.com/rs/zerolog/log"
)

//
// Scaffolder
//

type Scaffolder struct {
	basis    *basis.Basis
	Language string
	Policy   bool
	Role     bool
	Env      bool
}

func New(basis *basis.Basis, language string) *Scaffolder {
	return &Scaffolder{
		basis:    basis,
		Language: language,
	}
}

func (s *Scaffolder) WithPolicy() *Scaffolder {
	s.Policy = true
	return s
}

func (s *Scaffolder) WithRole() *Scaffolder {
	s.Role = true
	return s
}

func (s *Scaffolder) WithEnv() *Scaffolder {
	s.Env = true
	return s
}

//
// Create
//

func (s *Scaffolder) Create() error {
	scaffoldPath := filepath.Join("templates", s.Language)
	if _, err := Templates.Open(scaffoldPath); err != nil {
		return fmt.Errorf("invalid scaffold type '%s'", s.Language)
	}

	if err := s.copyScaffold(scaffoldPath); err != nil {
		return err
	}

	if err := s.writeTemplates(); err != nil {
		return err
	}

	log.Info().Msgf("initialized new %s scaffold", s.Language)
	return nil
}

func (s *Scaffolder) copyScaffold(scaffoldPath string) error {
	return fs.WalkDir(Templates, scaffoldPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath := strings.TrimPrefix(path, scaffoldPath)
		if relPath == "" || d.IsDir() {
			return nil
		}

		destPath := filepath.Base(relPath)
		if fileExists(destPath) {
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

func (s *Scaffolder) writeTemplates() error {
	if s.Policy {
		content, err := s.basis.PolicyTemplate()
		if err != nil {
			return fmt.Errorf("failed to read policy template: %w", err)
		}
		if err := s.writeTemplate("policy.json.tmpl", content); err != nil {
			return err
		}
	}

	if s.Role {
		content, err := s.basis.RoleTemplate()
		if err != nil {
			return fmt.Errorf("failed to read role template: %w", err)
		}
		if err := s.writeTemplate("role.json.tmpl", content); err != nil {
			return err
		}
	}

	if s.Env {
		content, err := s.basis.EnvTemplate()
		if err != nil {
			return fmt.Errorf("failed to read env template: %w", err)
		}
		if err := s.writeTemplate(".env.tmpl", content); err != nil {
			return err
		}
	}

	return nil
}

func (s *Scaffolder) writeTemplate(filename, content string) error {
	if fileExists(filename) {
		log.Info().Str("file", filename).Msg("skipping existing file")
		return nil
	}

	log.Info().Str("file", filename).Msg("creating")
	return os.WriteFile(filename, []byte(content), 0644)
}

//
// Helpers
//

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}