package param

import (
	"fmt"
	"path/filepath"
	"strings"

	v "github.com/go-ozzo/ozzo-validation/v4"
)

type Target struct {
	Service   string `arg:"--service,env:MONAD_SERVICE" placeholder:"name" help:"deployed service name" default:"${basename $PWD}"`
	Image     string `arg:"--image,env:MONAD_IMAGE" placeholder:"path" help:"deployed service ecr image" default:"${owner}/${repo}/${service}:${branch}"`
	ImagePath string `arg:"-"`
	ImageTag  string `arg:"-"`
}

func (t *Target) Validate(git Git) error {
	if t.Service == "" {
		t.Service = filepath.Base(git.cwd)
	}

	if t.Image == "" {
		t.Image = fmt.Sprintf("%s/%s/%s:%s", git.Owner, git.Repository, t.Service, git.Branch)
	}

	if !strings.Contains(t.Image, ":") {
		t.Image = fmt.Sprintf("%s:%s", t.Image, git.Branch)
	}

	parts := strings.Split(t.Image, ":")
	t.ImagePath = parts[0]
	t.ImageTag = parts[1]

	return v.ValidateStruct(t,
		v.Field(&t.Service, v.Required),
		v.Field(&t.Image, v.Required),
		v.Field(&t.ImagePath, v.Required),
		v.Field(&t.ImageTag, v.Required),
	)
}
