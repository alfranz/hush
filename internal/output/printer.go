package output

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// PrintResult prints the summary line and, on failure or warnings, the output.
// indent goes before each output line. A single command uses no indent,
// because each indented line costs about one token. Batch uses an indent, so
// that output lines cannot be mistaken for summary lines.
func PrintResult(w io.Writer, label string, exitCode int, filteredOutput []byte, warningCount int, warningOutput []byte, indent string) {
	if exitCode == 0 {
		if warningCount > 0 {
			printWarningSuccess(w, label, warningCount, warningOutput, indent)
			return
		}
		printSuccess(w, label)
	} else {
		printFailure(w, label, filteredOutput, indent)
	}
}

func printSuccess(w io.Writer, label string) {
	fmt.Fprintf(w, "✓ %s\n", label)
}

func printFailure(w io.Writer, label string, output []byte, indent string) {
	fmt.Fprintf(w, "✗ %s\n", label)
	if len(output) > 0 {
		fmt.Fprintf(w, "%s\n", indentOutput(output, indent))
	}
}

func printWarningSuccess(w io.Writer, label string, warningCount int, output []byte, indent string) {
	suffix := "warnings"
	if warningCount == 1 {
		suffix = "warning"
	}
	fmt.Fprintf(w, "⚠ %s (%d %s)\n", label, warningCount, suffix)
	if len(output) == 0 {
		return
	}

	shown := countLines(output)
	fmt.Fprintf(w, "%s\n", indentOutput(output, indent))
	if warningCount > shown {
		fmt.Fprintf(w, "%s... and %d more\n", indent, warningCount-shown)
	}
}

func PrintBatchSummary(w io.Writer, passed, total int) {
	if passed == total {
		fmt.Fprintf(w, "✓ %d/%d checks passed\n", passed, total)
	} else {
		fmt.Fprintf(w, "✗ %d/%d checks passed\n", passed, total)
	}
}

func indentOutput(b []byte, indent string) string {
	s := string(b)
	s = strings.TrimSuffix(s, "\n")
	return indent + strings.ReplaceAll(s, "\n", "\n"+indent)
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
