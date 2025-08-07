package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"miniShell/api/domain"
	"miniShell/internal/commands"
	"os"
	"os/signal"
	"strings"
	"sync"
)

var commandFactory = map[string]func() domain.Command{
	"cd":   func() domain.Command { return &commands.CdCommand{} },
	"pwd":  func() domain.Command { return &commands.PwdCommand{} },
	"echo": func() domain.Command { return &commands.EchoCmd{} },
	"kill": func() domain.Command { return &commands.KillCmd{} },
	"ps":   func() domain.Command { return &commands.PsCmd{} },
}

func main() {
	if err := prepare(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "prepare error:", err)
		return
	}

	if err := serveMiniShell(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "serve mini shell error:", err)
		return
	}
}

func prepare() error {

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	return os.Chdir(home)
}

func serveMiniShell() error {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		dir, err := os.Getwd()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		}
		fmt.Printf("\uF306 ) \uF07C  %s ) ", dir)

		if !scanner.Scan() {
			return nil
		}
		if err = scanner.Err(); err != nil {
			return err
		}

		line := scanner.Text()
		if line == "" {
			continue
		}
		if strings.ToLower(line) == "exit" {
			return nil
		}

		signalChanCancel := make(chan os.Signal, 1)
		signal.Notify(signalChanCancel, os.Interrupt)

		ctx, cancelFunc := context.WithCancel(context.Background())
		go func() {
			<-signalChanCancel
			cancelFunc()
			signal.Stop(signalChanCancel)
			close(signalChanCancel)
		}()

		pipelineStages := strings.Split(line, "|")

		if err = executePipeline(ctx, pipelineStages); err != nil {
			if errors.Is(err, context.Canceled) {
				_, _ = fmt.Fprintln(os.Stderr, "error pipeline:", err)
			} else {
				fmt.Println()
			}

		}

		signal.Stop(signalChanCancel)
		cancelFunc()
	}
}

func executePipeline(ctx context.Context, stages []string) error {
	cmds := make([]domain.Command, len(stages))
	for i, stage := range stages {
		args := strings.Fields(strings.TrimSpace(stage))
		if len(args) == 0 {
			return fmt.Errorf("invalid command in pipline: '%s'", stage)
		}

		for i, arg := range args {
			if len(arg) >= 2 && arg[0] == '"' && arg[len(arg)-1] == '"' {
				args[i] = arg[1 : len(arg)-1]
			}
		}

		name := args[0]
		cmdArgs := args[1:]

		var cmd domain.Command
		cmdFunc, isBuiltin := commandFactory[name]
		if isBuiltin {
			cmd = cmdFunc()
			if err := cmd.Parse(cmdArgs); err != nil {
				return err
			}
		} else {
			cmd = &commands.ExternalCmd{}
			if err := cmd.Parse(args); err != nil {
				return err
			}
		}
		cmds[i] = cmd
	}

	var wg sync.WaitGroup

	errChan := make(chan error, len(cmds))

	var pipeIn io.Reader = os.Stdin

	for i, cmd := range cmds {
		currentCmd := cmd
		currentPipeIn := pipeIn

		if i < len(cmds)-1 {
			r, w, err := os.Pipe()
			if err != nil {
				return fmt.Errorf("error create pipe: %w", err)
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				defer func(w *os.File) {
					err = w.Close()
					if err != nil {

					}
				}(w)

				if err = currentCmd.Execute(ctx, currentPipeIn, w); err != nil {
					errChan <- err
				}
			}()

			pipeIn = r

		} else {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := currentCmd.Execute(ctx, currentPipeIn, os.Stdout); err != nil {
					errChan <- err
				}
			}()
		}
	}

	wg.Wait()
	close(errChan)

	var allErrors []error
	for err := range errChan {
		allErrors = append(allErrors, err)
	}

	if len(allErrors) > 0 {
		return fmt.Errorf("error execution: %v", allErrors)
	}

	return nil
}
