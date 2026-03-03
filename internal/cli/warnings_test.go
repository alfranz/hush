package cli

import (
	"strings"
	"testing"
)

func TestBuildWarningReportNoPattern(t *testing.T) {
	report := buildWarningReport([]byte("warning TS1000\n"), sharedFlags{})
	if report.count != 0 {
		t.Fatalf("expected 0 warnings, got %d", report.count)
	}
}

func TestBuildWarningReportDefaultTail(t *testing.T) {
	var sb strings.Builder
	for i := 0; i < 12; i++ {
		sb.WriteString("warning TS1000\n")
	}

	report := buildWarningReport([]byte(sb.String()), sharedFlags{warnPattern: `warning TS[0-9]+`})
	if report.count != 12 {
		t.Fatalf("expected 12 warnings, got %d", report.count)
	}
	lines := strings.Split(strings.TrimSpace(string(report.lines)), "\n")
	if len(lines) != 10 {
		t.Fatalf("expected default warn tail of 10 lines, got %d", len(lines))
	}
}

func TestBuildWarningReportRespectsFilterFlags(t *testing.T) {
	input := []byte("warning TS1000\nwarning TS2000\nwarning TS3000\n")
	report := buildWarningReport(input, sharedFlags{
		warnPattern: `warning TS[0-9]+`,
		warnTail:    5,
		head:        2,
		grep:        "TS2",
	})

	if report.count != 3 {
		t.Fatalf("expected 3 warnings, got %d", report.count)
	}
	if got := strings.TrimSpace(string(report.lines)); got != "warning TS2000" {
		t.Fatalf("unexpected filtered warnings: %q", got)
	}
}
