package service

import (
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v11"
	v "github.com/go-ozzo/ozzo-validation/v4"
)

// Basis

type Basis struct {
	Name string `env:"MONAD_SERVICE" flag:"--service" usage:"Service name"`
}

//
// Derive
//

func Derive() (*Basis, error) {
	var err error
	var basis Basis

	if err = env.Parse(&basis); err != nil {
		return nil, err
	}

	if basis.Name == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		basis.Name = filepath.Base(wd)
	}

	return &basis, nil
}

//
// Validations
//

func (s *Basis) Validate() error {
	return v.ValidateStruct(s,
		v.Field(&s.Name, v.Required),
	)
}
