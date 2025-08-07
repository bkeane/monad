package route

import (
	"context"

	"github.com/bkeane/monad/pkg/scaffold"
)

type Scaffold struct {
	Scaffold string `arg:"positional,required" help:"go, python, node, ruby, shell"`
	Policy   bool   `arg:"--policy" help:"create policy.json.tmpl"`
	Role     bool   `arg:"--role" help:"create role.json.tmpl"`
	Env      bool   `arg:"--env" help:"create .env.tmpl"`
}

func (s *Scaffold) Route(ctx context.Context, r Root) error {
	scaffolder := scaffold.New(s.Scaffold)
	
	if s.Policy {
		scaffolder = scaffolder.WithPolicy()
	}
	
	if s.Role {
		scaffolder = scaffolder.WithRole()
	}
	
	if s.Env {
		scaffolder = scaffolder.WithEnv()
	}
	
	return scaffolder.Create()
}
