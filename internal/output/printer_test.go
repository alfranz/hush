package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintResultSuccess(t *testing.T) {
	var buf bytes.Buffer
	PrintResult(&buf, "echo", 0, nil)
	got := buf.String()
	if got != "✓ echo\n" {
		t.Errorf("expected '✓ echo\\n', got: %q", got)
	}
}

func TestPrintResultFailure(t *testing.T) {
	var buf bytes.Buffer
	PrintResult(&buf, "test", 1, []byte("FAIL: something broke\n"))
	got := buf.String()
	if !strings.Contains(got, "✗ test") {
		t.Errorf("expected failure marker, got: %q", got)
	}
	if !strings.Contains(got, "FAIL: something broke") {
		t.Errorf("expected output, got: %q", got)
	}
}

func TestPrintResultNoDuration(t *testing.T) {
	var buf bytes.Buffer
	PrintResult(&buf, "echo", 0, nil)
	got := buf.String()
	if strings.Contains(got, "(") {
		t.Errorf("expected no duration, got: %q", got)
	}
}

func TestPrintResultNoANSI(t *testing.T) {
	var buf bytes.Buffer
	PrintResult(&buf, "echo", 0, nil)
	got := buf.String()
	if strings.Contains(got, "\x1b[") {
		t.Errorf("expected no ANSI codes, got: %q", got)
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
