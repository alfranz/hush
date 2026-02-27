package filter

import (
	"strings"
	"testing"
)

func TestCollapseTracebacks(t *testing.T) {
	input := `Some output
Traceback (most recent call last):
  File "test.py", line 10, in <module>
    foo()
  File "test.py", line 5, in foo
    bar()
  File "test.py", line 2, in bar
    raise ValueError("bad")
ValueError: bad
More output`

	got := string(collapseTracebacks([]byte(input)))
	if !strings.Contains(got, "Traceback ... (3 frames) test.py:2 → ValueError: bad") {
		t.Errorf("expected collapsed traceback with location, got:\n%s", got)
	}
	if !strings.Contains(got, "Some output") {
		t.Error("expected surrounding output to be preserved")
	}
	if !strings.Contains(got, "More output") {
		t.Error("expected surrounding output to be preserved")
	}
}

func TestRemoveProgressLines(t *testing.T) {
	input := "line1\nprogress\r50%\nline2\n50%|████ | 5/10\nline3"
	got := string(removeProgressLines([]byte(input)))
	if strings.Contains(got, "progress") {
		t.Errorf("expected progress line with \\r removed, got:\n%s", got)
	}
	if strings.Contains(got, "████") {
		t.Errorf("expected tqdm line removed, got:\n%s", got)
	}
	if !strings.Contains(got, "line1") || !strings.Contains(got, "line2") || !strings.Contains(got, "line3") {
		t.Errorf("expected normal lines preserved, got:\n%s", got)
	}
}

func TestRemoveTimestamps(t *testing.T) {
	input := "2024-01-15T10:30:00 INFO Starting\n2024-01-15 10:30:00.123 DEBUG detail\nno timestamp here"
	got := string(removeTimestamps([]byte(input)))
	if strings.Contains(got, "2024-01-15") {
		t.Errorf("expected timestamps removed, got:\n%s", got)
	}
	if !strings.Contains(got, "INFO Starting") {
		t.Error("expected message preserved after timestamp removal")
	}
	if !strings.Contains(got, "no timestamp here") {
		t.Error("expected non-timestamp lines preserved")
	}
}

func TestApplyAgentMode(t *testing.T) {
	input := "\x1b[31m2024-01-15T10:30:00 ERROR something failed\x1b[0m"
	got := string(ApplyAgentMode([]byte(input)))
	if strings.Contains(got, "\x1b[") {
		t.Error("expected ANSI stripped")
	}
	if strings.Contains(got, "2024-01-15") {
		t.Error("expected timestamp stripped")
	}
	if !strings.Contains(got, "ERROR something failed") {
		t.Errorf("expected message preserved, got: %q", got)
	}
}
