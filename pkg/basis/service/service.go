package service

import (
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v11"
	v "github.com/go-ozzo/ozzo-validation/v4"
)

//
// Basis
//

type Basis struct {
	ServiceName string `env:"MONAD_SERVICE" flag:"--service" usage:"Service name" hint:"name"`
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

	if basis.ServiceName == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		basis.ServiceName = filepath.Base(wd)
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
		v.Field(&s.ServiceName, v.Required),
	)
}

//
// Accessors
//

func (s *Basis) Name() string {
	return s.ServiceName
}
