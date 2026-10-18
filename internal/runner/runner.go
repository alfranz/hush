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

// maxLabelLen is the longest derived label, in characters.
const maxLabelLen = 40

// deriveLabel uses the whole command as the label, so that the label shows
// which command failed. It drops the directory of the program, collapses
// whitespace and cuts the result to maxLabelLen characters.
func deriveLabel(command string) string {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return "unknown"
	}
	// Strip path prefix
	if idx := strings.LastIndex(fields[0], "/"); idx >= 0 {
		fields[0] = fields[0][idx+1:]
	}
	label := []rune(strings.Join(fields, " "))
	if len(label) > maxLabelLen {
		return string(label[:maxLabelLen-3]) + "..."
	}
	return string(label)
}
