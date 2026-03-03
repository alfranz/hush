package output

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

func PrintResult(w io.Writer, label string, exitCode int, filteredOutput []byte, warningCount int, warningOutput []byte) {
	if exitCode == 0 {
		if warningCount > 0 {
			printWarningSuccess(w, label, warningCount, warningOutput)
			return
		}
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
		fmt.Fprintf(w, "  %s\n", indentOutput(output))
	}
}

func printWarningSuccess(w io.Writer, label string, warningCount int, output []byte) {
	suffix := "warnings"
	if warningCount == 1 {
		suffix = "warning"
	}
	fmt.Fprintf(w, "⚠ %s (%d %s)\n", label, warningCount, suffix)
	if len(output) == 0 {
		return
	}

	shown := countLines(output)
	fmt.Fprintf(w, "  %s\n", indentOutput(output))
	if warningCount > shown {
		fmt.Fprintf(w, "  ... and %d more\n", warningCount-shown)
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
	s = strings.TrimSuffix(s, "\n")
	return strings.ReplaceAll(s, "\n", "\n  ")
}

func countLines(b []byte) int {
	count := 0
	for line := range bytes.SplitSeq(b, []byte("\n")) {
		if len(line) > 0 {
			count++
		}
	}
	return count
}
