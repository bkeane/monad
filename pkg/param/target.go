package param

import (
	"fmt"
	"path/filepath"

	"github.com/bkeane/monad/pkg/registry"
	v "github.com/go-ozzo/ozzo-validation/v4"
)

type Target struct {
	Service      string                 `arg:"--service" placeholder:"name" help:"service name" default:"${basename $PWD}"`
	Image        string                 `arg:"--image" placeholder:"path" help:"ecr image path" default:"${owner}/${repo}/${service}"`
	ImagePointer *registry.ImagePointer `arg:"-"`
}

func (t *Target) Validate(git Git) error {
	if t.Service == "" {
		t.Service = filepath.Base(git.cwd)
	}

	if t.Image == "" {
		t.Image = fmt.Sprintf("%s/%s/%s", git.Owner, git.Repository, t.Service)
	}

	return v.ValidateStruct(t,
		v.Field(&t.Service, v.Required),
		v.Field(&t.Image, v.Required),
	)
}
