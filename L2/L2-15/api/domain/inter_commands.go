package domain

import (
	"context"
	"io"
)

// Command is interface for realizing commands for execution in pipeline
type Command interface {
	Parse(args []string) error

	Execute(ctx context.Context, stdin io.Reader, stdout io.Writer) error
}
