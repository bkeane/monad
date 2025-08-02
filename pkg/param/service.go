package param

import (
	"fmt"
	"path/filepath"
	"strings"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/rs/zerolog/log"
)

type ServiceConfig struct {
	Name      string `arg:"--service,env:MONAD_SERVICE" placeholder:"name" help:"service name" default:"basename $PWD"`
	Image     string `arg:"--image,env:MONAD_IMAGE" placeholder:"path" help:"service ecr image path" default:"${owner}/${repo}/${service}:${branch}"`
	ImagePath string `arg:"-"`
	ImageTag  string `arg:"-"`
}

func (s *ServiceConfig) Process(git GitConfig) error {
	if s.Name == "" {
		s.Name = filepath.Base(git.cwd)
	}

	if s.Image == "" {
		s.Image = fmt.Sprintf("%s/%s/%s:%s", git.Owner, git.Repository, s.Name, git.Branch)
	}

	if !strings.Contains(s.Image, ":") {
		s.Image = fmt.Sprintf("%s:%s", s.Image, git.Branch)
	}

	parts := strings.Split(s.Image, ":")
	s.ImagePath = parts[0]
	s.ImageTag = parts[1]

	log.Info().
		Str("image", s.ImagePath).
		Str("tag", s.ImageTag).
		Str("name", s.Name).
		Msg("service")

	return s.Validate()
}

func (s *ServiceConfig) Validate() error {
	return v.ValidateStruct(s,
		v.Field(&s.Name, v.Required),
		v.Field(&s.Image, v.Required),
		v.Field(&s.ImagePath, v.Required),
		v.Field(&s.ImageTag, v.Required),
	)
}
