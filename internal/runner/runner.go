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

type Options struct {
	Command string
	Label   string
}

func Run(ctx context.Context, opts Options) (*Result, error) {
	label := opts.Label
	if label == "" {
		label = deriveLabel(opts.Command)
	}

	start := time.Now()
	cmd := exec.CommandContext(ctx, "sh", "-c", opts.Command)
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	exitCode := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, err
		}
	}

	return &Result{
		Label:    label,
		Command:  opts.Command,
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
