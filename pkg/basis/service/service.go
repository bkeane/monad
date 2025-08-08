package service

import (
	"os"
	"path/filepath"

	v "github.com/go-ozzo/ozzo-validation/v4"
)

// Data

type Data struct {
	name string `bind:"--service,MONAD_SERVICE" hint:"name" desc:"service name"`
}

//
// Derive
//

func Derive() (*Data, error) {
	var data Data

	if data.name == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		data.name = filepath.Base(wd)
	}

	return &data, nil
}

//
// Validations
//

func (s *Data) Validate() error {
	return v.ValidateStruct(s,
		v.Field(&s.name, v.Required),
	)
}

//
// Accessors
//

func (s *Data) Name() string { return s.name } // Service name
