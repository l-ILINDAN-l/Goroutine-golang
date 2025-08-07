package commands

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CdCommand is structure for execution cd
type CdCommand struct {
	physicalPath bool
	targetPath   string
}

// Parse is function for parsing args
func (c *CdCommand) Parse(args []string) error {
	fs := flag.NewFlagSet("cd", flag.ContinueOnError)
	fs.BoolVar(&c.physicalPath, "P", false, "use physical path")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 1 {
		return fmt.Errorf("cd: too many arguments")
	}

	c.targetPath = fs.Arg(0)
	return nil
}

// Execute is function for execution command with context ctx, read from input io.Reader, write to stdout io.Writer
func (c *CdCommand) Execute(_ context.Context, _ io.Reader, _ io.Writer) error {
	var pathToChange string
	var err error

	if c.targetPath == "" {
		pathToChange, err = os.UserHomeDir()
		if err != nil {
			return err
		}
	} else {
		pathToChange, err = expandTilde(c.targetPath)
		if err != nil {
			return err
		}
	}

	if c.physicalPath {
		pathToChange, err = filepath.EvalSymlinks(pathToChange)
		if err != nil {
			return err
		}
	}

	return os.Chdir(pathToChange)
}

func expandTilde(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, path[1:]), nil
}
