package commands

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"syscall"
)

// KillCmd is structure for execution kill
type KillCmd struct {
	pid int
}

// Parse is function for parsing args
func (cmd *KillCmd) Parse(args []string) error {
	if len(args) == 1 {
		pid, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("kill: invalid PID '%s'", args[0])
		}
		cmd.pid = pid
		return nil
	}
	if len(args) > 1 {
		return errors.New("kill: to many arguments")
	}
	return nil
}

// Execute is function for execution command with context ctx, read from input io.Reader, write to stdout io.Writer
func (cmd *KillCmd) Execute(_ context.Context, stdin io.Reader, stdout io.Writer) error {
	if cmd.pid == 0 {
		inputBytes, err := io.ReadAll(stdin)
		if err != nil {
			return fmt.Errorf("kill: couldn't read PID from stdin: %w", err)
		}

		pidStr := string(bytes.TrimSpace(inputBytes))
		if pidStr == "" {
			return errors.New("kill: PID was not transmitted")
		}

		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			return fmt.Errorf("kill: invalid PID from stdin '%s'", pidStr)
		}
		cmd.pid = pid
	}

	return syscall.Kill(cmd.pid, syscall.SIGTERM)
}
