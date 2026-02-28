package output

import (
	"fmt"
	"io"
	"strings"
)

func PrintResult(w io.Writer, label string, exitCode int, filteredOutput []byte) {
	if exitCode == 0 {
		printSuccess(w, label)
	} else {
		printFailure(w, label, filteredOutput)
	}
}

func printSuccess(w io.Writer, label string) {
	fmt.Fprintf(w, "✓ %s\n", label)
}

func printFailure(w io.Writer, label string, output []byte) {
	fmt.Fprintf(w, "✗ %s\n", label)
	if len(output) > 0 {
		// Indent each line of output with 2 spaces
		fmt.Fprintf(w, "  %s\n", indentOutput(output))
	}
}

func PrintBatchSummary(w io.Writer, passed, total int) {
	if passed == total {
		fmt.Fprintf(w, "✓ %d/%d checks passed\n", passed, total)
	} else {
		fmt.Fprintf(w, "✗ %d/%d checks passed\n", passed, total)
	}
}

func indentOutput(b []byte) string {
	s := string(b)
	// Trim trailing newline to avoid extra blank line
	s = strings.TrimSuffix(s, "\n")
	return strings.ReplaceAll(s, "\n", "\n  ")
}
