package mysql

import (
	"context"
	"errors"
)

// Context owns MySQL repository context validation.
type Context struct{}

// Validate rejects nil or cancelled operation contexts.
func (Context) Validate(ctx context.Context, operation string) error {
	if ctx == nil {
		return errors.New(operation + ": context is nil")
	}
	if cause := context.Cause(ctx); cause != nil {
		return errors.Join(errors.New(operation+": context ended"), cause)
	}
	return nil
}
