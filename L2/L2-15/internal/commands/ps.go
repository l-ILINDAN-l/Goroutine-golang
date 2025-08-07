package commands

import (
	"context"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/shirou/gopsutil/v3/process"
)

// PsCmd is structure for execution ps
type PsCmd struct{}

// Parse is function for parsing args
func (cmd *PsCmd) Parse(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("ps: doesn`t accept arguments")
	}
	return nil
}

// Execute is function for execution command with context ctx, read from input io.Reader, write to stdout io.Writer
func (cmd *PsCmd) Execute(_ context.Context, _ io.Reader, stdout io.Writer) error {
	processes, err := process.Processes()
	if err != nil {
		return fmt.Errorf("couldn't get a list of processes: %w", err)
	}

	w := tabwriter.NewWriter(stdout, 0, 0, 3, ' ', tabwriter.Debug)

	_, err = fmt.Fprintln(w, "PID\tUSER\tCPU(%)\tMEM(%)\tCOMMAND")
	if err != nil {
		return err
	}

	for _, p := range processes {
		pid := p.Pid
		name, _ := p.Name()
		cpu, _ := p.CPUPercent()
		mem, _ := p.MemoryPercent()
		user, _ := p.Username()

		_, err = fmt.Fprintf(w, "%d\t%s\t%.2f\t%.2f\t%s\n",
			pid,
			user,
			cpu,
			mem,
			name,
		)
		if err != nil {
			continue
		}
	}

	return w.Flush()
}
