// externalCMD.go

package commands

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// ExternalCmd is structure for execution external commands
type ExternalCmd struct {
	name string
	args []string
}

// Parse is function for parsing args
func (c *ExternalCmd) Parse(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("external: not name command for execution")
	}
	c.name = args[0]
	c.args = args[1:]
	return nil
}

// Execute is function for execution command with context ctx, read from input io.Reader, write to stdout io.Writer
func (c *ExternalCmd) Execute(ctx context.Context, stdin io.Reader, stdout io.Writer) error {

	inputBytes, err := io.ReadAll(stdin)
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, c.name, c.args...)
	cmd.Stdin = bytes.NewReader(inputBytes)
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
