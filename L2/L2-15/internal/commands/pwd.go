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

// PwdCommand is structure for execution pwd
type PwdCommand struct {
	logicalPath  bool
	physicalPath bool
}

// Parse is function for parsing args
func (c *PwdCommand) Parse(args []string) error {
	fs := flag.NewFlagSet("pwd", flag.ContinueOnError)
	fs.BoolVar(&c.logicalPath, "L", false, "for the logical path")
	fs.BoolVar(&c.physicalPath, "P", false, "for the physical path")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if strings.Join(fs.Args(), " ") != "" {
		return fmt.Errorf("pwd: to many arguments")
	}

	return nil
}

// Execute is function for execution command with context ctx, read from input io.Reader, write to stdout io.Writer
func (c *PwdCommand) Execute(_ context.Context, _ io.Reader, stdout io.Writer) error {
	var path string

	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	if c.physicalPath {
		path, err = filepath.EvalSymlinks(dir)
		if err != nil {
			return err
		}
	} else {
		path = dir
	}

	_, err = fmt.Fprintln(stdout, path)
	return err
}
