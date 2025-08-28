package mock

import (
	"github.com/bkeane/monad/pkg/basis/defaults"
)

// NewMockDefaults creates a realistic mock defaults.Basis for testing
func NewMockDefaults() *defaults.Basis {
	// Use the real defaults derivation which loads embedded templates
	defaultsBasis, _ := defaults.Derive()
	return defaultsBasis
}

// NewMockDefaultsSimple creates a simple mock defaults without embedded content
func NewMockDefaultsSimple() *defaults.Basis {
	return &defaults.Basis{
		Env:    "TEST_VAR={{.Service.Name}}\nAWS_REGION={{.Account.Region}}",
		Policy: `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow", "Action": "logs:*", "Resource": "*"}]}`,
		Role:   `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow", "Principal": {"Service": "lambda.amazonaws.com"}, "Action": "sts:AssumeRole"}]}`,
	}
}