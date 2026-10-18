package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintResultSuccess(t *testing.T) {
	var buf bytes.Buffer
	PrintResult(&buf, "echo", 0, nil, 0, nil, "")
	got := buf.String()
	if got != "✓ echo\n" {
		t.Errorf("expected '✓ echo\\n', got: %q", got)
	}
}

func TestPrintResultFailure(t *testing.T) {
	var buf bytes.Buffer
	PrintResult(&buf, "test", 1, []byte("FAIL: something broke\n"), 0, nil, "")
	want := "✗ test\nFAIL: something broke\n"
	if got := buf.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPrintResultNoDuration(t *testing.T) {
	var buf bytes.Buffer
	PrintResult(&buf, "echo", 0, nil, 0, nil, "")
	got := buf.String()
	if strings.Contains(got, "(") {
		t.Errorf("expected no duration, got: %q", got)
	}
}

func TestPrintResultNoANSI(t *testing.T) {
	var buf bytes.Buffer
	PrintResult(&buf, "echo", 0, nil, 0, nil, "")
	got := buf.String()
	if strings.Contains(got, "\x1b[") {
		t.Errorf("expected no ANSI codes, got: %q", got)
	}
}

func TestPrintResultWarningSuccess(t *testing.T) {
	var buf bytes.Buffer
	PrintResult(&buf, "types", 0, nil, 3, []byte("w1\nw2"), "")
	want := "⚠ types (3 warnings)\nw1\nw2\n... and 1 more\n"
	if got := buf.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// Batch output indents each command's output, so that its lines cannot be
// mistaken for hush summary lines.
func TestPrintResultIndent(t *testing.T) {
	tests := []struct {
		name     string
		exitCode int
		output   []byte
		warnings []byte
		want     string
	}{
		{"failure", 1, []byte("a\nb\n"), nil, "✗ cmd\n  a\n  b\n"},
		{"warnings", 0, nil, []byte("w1\nw2"), "⚠ cmd (3 warnings)\n  w1\n  w2\n  ... and 1 more\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			warnCount := 0
			if tt.warnings != nil {
				warnCount = 3
			}
			PrintResult(&buf, "cmd", tt.exitCode, tt.output, warnCount, tt.warnings, "  ")
			if got := buf.String(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintBatchSummaryAllPass(t *testing.T) {
	var buf bytes.Buffer
	PrintBatchSummary(&buf, 3, 3)
	got := buf.String()
	if got != "✓ 3/3 checks passed\n" {
		t.Errorf("expected '✓ 3/3 checks passed\\n', got: %q", got)
	}
}

func TestPrintBatchSummaryWithFailure(t *testing.T) {
	var buf bytes.Buffer
	PrintBatchSummary(&buf, 1, 3)
	got := buf.String()
	if got != "✗ 1/3 checks passed\n" {
		t.Errorf("expected '✗ 1/3 checks passed\\n', got: %q", got)
	}
}
