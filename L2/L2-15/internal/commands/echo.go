package commands

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// EchoCmd is structure for execution echo
type EchoCmd struct {
	text string
}

// Parse is function for parsing args
func (c *EchoCmd) Parse(args []string) error {
	c.text = strings.Join(args, " ")
	return nil
}

// Execute is function for execution command with context ctx, read from input io.Reader, write to stdout io.Writer
func (c *EchoCmd) Execute(_ context.Context, _ io.Reader, stdout io.Writer) error {
	_, err := fmt.Fprintln(stdout, c.text)
	return err
}
