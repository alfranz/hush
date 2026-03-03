package filter

import (
	"strings"
	"testing"
)

func TestApplyHead(t *testing.T) {
	input := "line1\nline2\nline3\nline4\nline5\n"
	got := string(Apply([]byte(input), Options{Head: 3}))
	if got != "line1\nline2\nline3" {
		t.Errorf("expected first 3 lines, got: %q", got)
	}
}

func TestApplyTail(t *testing.T) {
	input := "line1\nline2\nline3\nline4\nline5\n"
	got := string(Apply([]byte(input), Options{Tail: 2}))
	if got != "line4\nline5" {
		t.Errorf("expected last 2 lines, got: %q", got)
	}
}

func TestApplyGrep(t *testing.T) {
	input := "INFO ok\nERROR bad\nINFO fine\nERROR worse\n"
	got := string(Apply([]byte(input), Options{Grep: "ERROR"}))
	if got != "ERROR bad\nERROR worse" {
		t.Errorf("expected only ERROR lines, got: %q", got)
	}
}

func TestApplyCombined(t *testing.T) {
	input := "a\nb\nc\nd\ne\nf\ng\n"
	got := string(Apply([]byte(input), Options{Tail: 5, Head: 3}))
	// Head 3 of tail 5 (c,d,e,f,g) = c,d,e
	if got != "c\nd\ne" {
		t.Errorf("expected head 3 of tail 5, got: %q", got)
	}
}

func TestApplyStripANSI(t *testing.T) {
	input := "\x1b[31merror\x1b[0m\n"
	got := string(Apply([]byte(input), Options{StripANSI: true}))
	if strings.Contains(got, "\x1b[") {
		t.Error("expected ANSI codes stripped")
	}
}

func TestApplyNoOp(t *testing.T) {
	input := "hello world\n"
	got := string(Apply([]byte(input), Options{}))
	if got != input {
		t.Errorf("expected passthrough, got: %q", got)
	}
}

func TestMatchLines(t *testing.T) {
	input := []byte("info\nwarning TS1000\nwarning TS2000\nerror\n")
	got := MatchLines(input, `warning TS[0-9]+`)
	if got.Count != 2 {
		t.Fatalf("expected 2 matches, got %d", got.Count)
	}
	if string(got.Lines) != "warning TS1000\nwarning TS2000" {
		t.Fatalf("unexpected matches: %q", string(got.Lines))
	}
}

func TestMatchLinesInvalidPattern(t *testing.T) {
	got := MatchLines([]byte("warning\n"), "[")
	if got.Count != 0 {
		t.Fatalf("expected 0 matches, got %d", got.Count)
	}
	if len(got.Lines) != 0 {
		t.Fatalf("expected empty lines, got %q", string(got.Lines))
	}
}
