package runner

import (
	"testing"
)

func TestRunSuccess(t *testing.T) {
	r, err := Run(t.Context(), Options{Command: "echo hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", r.ExitCode)
	}
	if string(r.Output) != "hello\n" {
		t.Errorf("unexpected output: %q", r.Output)
	}
}

func TestRunFailure(t *testing.T) {
	r, err := Run(t.Context(), Options{Command: "false"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ExitCode == 0 {
		t.Error("expected non-zero exit code")
	}
}

func TestRunExitCode(t *testing.T) {
	r, err := Run(t.Context(), Options{Command: "exit 42"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ExitCode != 42 {
		t.Errorf("expected exit code 42, got %d", r.ExitCode)
	}
}

func TestRunLabel(t *testing.T) {
	r, err := Run(t.Context(), Options{Command: "echo hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Label != "echo hello" {
		t.Errorf("expected label 'echo hello', got %q", r.Label)
	}
}

func TestRunCustomLabel(t *testing.T) {
	r, err := Run(t.Context(), Options{Command: "echo hello", Label: "my test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Label != "my test" {
		t.Errorf("expected label 'my test', got %q", r.Label)
	}
}

func TestDeriveLabel(t *testing.T) {
	tests := []struct {
		command string
		want    string
	}{
		{"echo hello", "echo hello"},
		{"/usr/bin/pytest -x", "pytest -x"},
		{"make build", "make build"},
		{"go test ./...", "go test ./..."},
		{"echo $HOME | grep -q nope", "echo $HOME | grep -q nope"},
		{"  go   vet\n  ./...  ", "go vet ./..."},
		{"pytest tests/unit/test_auth.py::test_login_fails -x", "pytest tests/unit/test_auth.py::test_..."},
		{"echo ✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓", "echo ✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓✓..."},
		{"", "unknown"},
	}
	for _, tt := range tests {
		got := deriveLabel(tt.command)
		if got != tt.want {
			t.Errorf("deriveLabel(%q) = %q, want %q", tt.command, got, tt.want)
		}
	}
}

func TestRunArgsLabel(t *testing.T) {
	r, err := Run(t.Context(), Options{Args: []string{"sh", "-c", "exit 0"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Label != "sh -c exit 0" {
		t.Errorf("expected label 'sh -c exit 0', got %q", r.Label)
	}
}
