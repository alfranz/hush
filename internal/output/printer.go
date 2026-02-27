package output

import (
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/term"
)

type Options struct {
	NoTime bool
	Color  *bool // nil = auto, true = force, false = disable
}

func isTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func useColor(opt *bool) bool {
	if opt != nil {
		return *opt
	}
	return isTTY()
}

func PrintResult(w io.Writer, label string, exitCode int, duration time.Duration, filteredOutput []byte, opts Options) {
	color := useColor(opts.Color)

	if exitCode == 0 {
		printSuccess(w, label, duration, opts.NoTime, color)
	} else {
		printFailure(w, label, duration, filteredOutput, opts.NoTime, color)
	}
}

func printSuccess(w io.Writer, label string, duration time.Duration, noTime, color bool) {
	if color {
		fmt.Fprintf(w, "\x1b[32m✓\x1b[0m %s", label)
	} else {
		fmt.Fprintf(w, "✓ %s", label)
	}
	if !noTime {
		dur := formatDuration(duration)
		if color {
			fmt.Fprintf(w, " \x1b[2m(%s)\x1b[0m", dur)
		} else {
			fmt.Fprintf(w, " (%s)", dur)
		}
	}
	fmt.Fprintln(w)
}

func printFailure(w io.Writer, label string, duration time.Duration, output []byte, noTime, color bool) {
	if color {
		fmt.Fprintf(w, "\x1b[31m✗\x1b[0m %s", label)
	} else {
		fmt.Fprintf(w, "✗ %s", label)
	}
	if !noTime {
		dur := formatDuration(duration)
		if color {
			fmt.Fprintf(w, " \x1b[2m(%s)\x1b[0m", dur)
		} else {
			fmt.Fprintf(w, " (%s)", dur)
		}
	}
	fmt.Fprintln(w)
	if len(output) > 0 {
		// Indent each line of output with 2 spaces
		fmt.Fprintf(w, "  %s\n", indentOutput(output))
	}
}

func PrintBatchSummary(w io.Writer, passed, total int, duration time.Duration, noTime bool, colorOpt *bool) {
	color := useColor(colorOpt)
	if passed == total {
		if color {
			fmt.Fprintf(w, "\x1b[32m✓\x1b[0m %d/%d checks passed", passed, total)
		} else {
			fmt.Fprintf(w, "✓ %d/%d checks passed", passed, total)
		}
	} else {
		if color {
			fmt.Fprintf(w, "\x1b[31m✗\x1b[0m %d/%d checks passed", passed, total)
		} else {
			fmt.Fprintf(w, "✗ %d/%d checks passed", passed, total)
		}
	}
	if !noTime {
		dur := formatDuration(duration)
		if color {
			fmt.Fprintf(w, " \x1b[2m(%s)\x1b[0m", dur)
		} else {
			fmt.Fprintf(w, " (%s)", dur)
		}
	}
	fmt.Fprintln(w)
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%.0fm%.0fs", d.Minutes(), d.Seconds()-float64(int(d.Minutes()))*60)
}

func indentOutput(b []byte) string {
	s := string(b)
	// Trim trailing newline to avoid extra blank line
	if len(s) > 0 && s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}
	// Replace newlines with newline + indent
	result := ""
	for i, c := range s {
		result += string(c)
		if c == '\n' && i < len(s)-1 {
			result += "  "
		}
	}
	return result
}
