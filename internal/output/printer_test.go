package output

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func ptrTo[T any](v T) *T { return &v }

func TestPrintResultSuccess(t *testing.T) {
	var buf bytes.Buffer
	PrintResult(&buf, "echo", 0, 100*time.Millisecond, nil, Options{Color: ptrTo(false)})
	got := buf.String()
	if !strings.Contains(got, "✓ echo") {
		t.Errorf("expected success marker, got: %q", got)
	}
	if !strings.Contains(got, "(0.1s)") {
		t.Errorf("expected duration, got: %q", got)
	}
}

func TestPrintResultFailure(t *testing.T) {
	var buf bytes.Buffer
	PrintResult(&buf, "test", 1, 500*time.Millisecond, []byte("FAIL: something broke\n"), Options{Color: ptrTo(false)})
	got := buf.String()
	if !strings.Contains(got, "✗ test") {
		t.Errorf("expected failure marker, got: %q", got)
	}
	if !strings.Contains(got, "FAIL: something broke") {
		t.Errorf("expected output, got: %q", got)
	}
}

func TestPrintResultNoTime(t *testing.T) {
	var buf bytes.Buffer
	PrintResult(&buf, "echo", 0, 100*time.Millisecond, nil, Options{NoTime: true, Color: ptrTo(false)})
	got := buf.String()
	if strings.Contains(got, "(") {
		t.Errorf("expected no duration, got: %q", got)
	}
}

func TestPrintBatchSummaryAllPass(t *testing.T) {
	var buf bytes.Buffer
	PrintBatchSummary(&buf, 3, 3, time.Second, false, ptrTo(false))
	got := buf.String()
	if !strings.Contains(got, "✓ 3/3 checks passed") {
		t.Errorf("expected all pass summary, got: %q", got)
	}
}

func TestPrintBatchSummaryWithFailure(t *testing.T) {
	var buf bytes.Buffer
	PrintBatchSummary(&buf, 1, 3, time.Second, false, ptrTo(false))
	got := buf.String()
	if !strings.Contains(got, "✗ 1/3 checks passed") {
		t.Errorf("expected failure summary, got: %q", got)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{100 * time.Millisecond, "0.1s"},
		{1500 * time.Millisecond, "1.5s"},
		{30 * time.Second, "30.0s"},
	}
	for _, tt := range tests {
		got := formatDuration(tt.d)
		if got != tt.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}
