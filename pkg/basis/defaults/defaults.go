package defaults

import (
	"embed"

	v "github.com/go-ozzo/ozzo-validation/v4"
)

//go:embed embed/*
var defaults embed.FS

// Basis

type Basis struct {
	Policy string
	Role   string
	Rule string
	Env    string
}

//
// Derive
//

func Derive() (*Basis, error) {
	var err error
	var basis Basis

	basis.Policy, err = read("embed/policy.json.tmpl")
	if err != nil {
		return nil, err
	}

	basis.Role, err = read("embed/role.json.tmpl")
	if err != nil {
		return nil, err
	}

	basis.Rule, err = read("embed/rule.json.tmpl")
	if err != nil {
		return nil, err
	}

	basis.Env, err = read("embed/env.tmpl")
	if err != nil {
		return nil, err
	}

	err = basis.Validate()
	if err != nil {
		return nil, err
	}

	return &basis, nil
}

//
// Validations
//

func (s *Basis) Validate() error {
	return v.ValidateStruct(s,
		v.Field(&s.Policy, v.Required),
		v.Field(&s.Role, v.Required),
		v.Field(&s.Rule, v.Required),
		v.Field(&s.Env, v.Required),
	)
}

// Accessors

func (s *Basis) PolicyTemplate() string {
	return s.Policy
}

func (s *Basis) RoleTemplate() string {
	return s.Role
}

func (s *Basis) RuleTemplate() string {
	return s.Rule
}

func (s *Basis) EnvTemplate() string {
	return s.Env
}

//
// Helpers
//

func read(path string) (string, error) {
	data, err := defaults.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
