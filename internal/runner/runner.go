package runner

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"time"
)

type Result struct {
	Label    string
	Command  string
	ExitCode int
	Output   []byte
	Duration time.Duration
}

// Options selects how to run a command.
// If Args is set, Args[0] runs directly with Args[1:] as its arguments, and
// Command is ignored. Otherwise Command runs with "sh -c".
type Options struct {
	Command string
	Args    []string
	Label   string
}

func Run(ctx context.Context, opts Options) (*Result, error) {
	command := opts.Command
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	if len(opts.Args) > 0 {
		command = strings.Join(opts.Args, " ")
		cmd = exec.CommandContext(ctx, opts.Args[0], opts.Args[1:]...)
	}

	label := opts.Label
	if label == "" {
		label = deriveLabel(command)
	}

	start := time.Now()
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	exitCode := 0
	if err != nil {
		var exitErr *exec.ExitError
		switch {
		case errors.As(err, &exitErr):
			exitCode = exitErr.ExitCode()
		case len(opts.Args) > 0:
			// The program could not start (not found, not executable).
			// Report it like a shell does: exit 127 and the error as output.
			exitCode = 127
			output = []byte(err.Error() + "\n")
		default:
			return nil, err
		}
	}

	return &Result{
		Label:    label,
		Command:  command,
		ExitCode: exitCode,
		Output:   output,
		Duration: duration,
	}, nil
}

func deriveLabel(command string) string {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return "unknown"
	}
	// Use the first token as the label
	token := fields[0]
	// Strip path prefix
	if idx := strings.LastIndex(token, "/"); idx >= 0 {
		token = token[idx+1:]
	}
	return token
}
