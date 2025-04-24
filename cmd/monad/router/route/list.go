package route

import (
	"context"
)

type List struct{}

func (e *List) Route(ctx context.Context, r Root) error {
	return nil
}
