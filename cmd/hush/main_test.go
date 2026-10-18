package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// hushBin is the path to a hush binary built once for all tests in this file.
// runRoot calls os.Exit, so the CLI is tested end to end as a real process.
var hushBin string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "hush-test")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	hushBin = filepath.Join(dir, "hush")
	if out, err := exec.Command("go", "build", "-o", hushBin, ".").CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build hush: %v\n%s", err, out)
		os.Exit(1)
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// runHush runs hush in an empty directory, so no .hush.yaml is loaded.
func runHush(t *testing.T, args ...string) (string, int) {
	t.Helper()
	return runHushIn(t, t.TempDir(), args...)
}

// runHushIn runs hush in dir.
func runHushIn(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(hushBin, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	code := 0
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("run hush: %v", err)
		}
		code = exitErr.ExitCode()
	}
	return string(out), code
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	return line
}

func TestCommandForms(t *testing.T) {
	// "ls /nonexistent" exits 1 with BSD ls (macOS) and 2 with GNU ls (Linux).
	// hush must pass through whatever code ls returns.
	lsExit := exec.Command("ls", "/nonexistent").Run().(*exec.ExitError).ExitCode()

	tests := []struct {
		name      string
		args      []string
		wantLine  string
		wantCode  int
		wantInOut string
	}{
		// Bug 1: arguments after the first word must not be dropped.
		{"unquoted", []string{"ls", "/nonexistent"}, "✗ ls /nonexistent", lsExit, "/nonexistent"},
		{"unquoted with label", []string{"--label", "x", "ls", "/nonexistent"}, "✗ x", lsExit, "/nonexistent"},
		{"double dash", []string{"--", "ls", "/nonexistent"}, "✗ ls /nonexistent", lsExit, "/nonexistent"},
		{"quoted", []string{"ls /nonexistent"}, "✗ ls /nonexistent", lsExit, "/nonexistent"},
		{"batch quoted", []string{"batch", "ls /nonexistent"}, "✗ ls /nonexistent", lsExit, "/nonexistent"},

		// Bug 2: flags after the command belong to the command.
		{"command flag", []string{"ls", "-la", "/nonexistent"}, "✗ ls -la /nonexistent", lsExit, "/nonexistent"},
		{"sh -c", []string{"sh", "-c", "exit 3"}, "✗ sh -c exit 3", 3, ""},
		{"hush flags then double dash", []string{"--label", "t", "--", "sh", "-c", "exit 3"}, "✗ t", 3, ""},
		{"hush --help after command", []string{"sh", "-c", "exit 0", "--help"}, "✓ sh -c exit 0 --help", 0, ""},

		// Quoted form still runs through the shell.
		{"quoted pipe", []string{"echo hi | grep -q nope"}, "✗ echo hi | grep -q nope", 1, ""},
		{"quoted success", []string{"true"}, "✓ true", 0, ""},
		{"unquoted success", []string{"sh", "-c", "true"}, "✓ sh -c true", 0, ""},

		// Unquoted form does not go through a shell: metacharacters are plain arguments.
		{"unquoted no shell", []string{"echo", "a;", "exit", "1"}, "✓ echo a; exit 1", 0, ""},

		// Flags after a quoted command are no longer hush flags. Fail with a hint.
		{"flag after quoted command", []string{"sh -c 'exit 0'", "--label", "x"}, `hush: "sh -c 'exit 0'" looks like a quoted command, but more arguments follow it: ["--label" "x"]`, 1, "hush [flags]"},

		// A missing program fails like it does in a shell.
		{"not found", []string{"hush-no-such-program", "arg"}, "✗ hush-no-such-program arg", 127, "not found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, code := runHush(t, tt.args...)
			if got := firstLine(out); got != tt.wantLine {
				t.Errorf("first line = %q, want %q\nfull output:\n%s", got, tt.wantLine, out)
			}
			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d\nfull output:\n%s", code, tt.wantCode, out)
			}
			if tt.wantInOut != "" && !strings.Contains(out, tt.wantInOut) {
				t.Errorf("output does not contain %q\nfull output:\n%s", tt.wantInOut, out)
			}
		})
	}
}

func TestBatch(t *testing.T) {
	t.Run("labels name each command", func(t *testing.T) {
		out, code := runHush(t, "batch", "--continue", "go version", "sh -c 'exit 2'")
		if !strings.Contains(out, "✓ go version\n") || !strings.Contains(out, "✗ sh -c 'exit 2'\n") {
			t.Errorf("labels do not name each command:\n%s", out)
		}
		if code != 2 {
			t.Errorf("exit code = %d, want 2", code)
		}
	})
	t.Run("label flag is rejected", func(t *testing.T) {
		out, code := runHush(t, "batch", "--label", "mine", "true", "false")
		if strings.Contains(out, "✓") || strings.Contains(out, "✗") {
			t.Errorf("batch ran with --label:\n%s", out)
		}
		if !strings.Contains(out, "--label") || code != 1 {
			t.Errorf("want an error about --label and exit 1, got exit %d:\n%s", code, out)
		}
	})
}

// TestNeverFalseSuccess guards bug 1 directly: a command that fails only when
// it gets all its arguments must never print ✓.
func TestNeverFalseSuccess(t *testing.T) {
	forms := [][]string{
		{"sh", "-c", "exit 7"},
		{"--", "sh", "-c", "exit 7"},
		{"--label", "x", "sh", "-c", "exit 7"},
		{"--tail", "1", "--", "sh", "-c", "exit 7"},
		{"sh -c 'exit 7'"},
	}
	for _, args := range forms {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			out, code := runHush(t, args...)
			if strings.HasPrefix(out, "✓") {
				t.Errorf("printed ✓ for a failing command:\n%s", out)
			}
			if code != 7 {
				t.Errorf("exit code = %d, want 7", code)
			}
		})
	}
}

func TestVersion(t *testing.T) {
	out, code := runHush(t, "--version")
	if code != 0 || !strings.HasPrefix(out, "hush version ") {
		t.Errorf("hush --version: exit %d, output %q", code, out)
	}

	// Release builds set the version with ldflags, as goreleaser does.
	bin := filepath.Join(t.TempDir(), "hush")
	build := exec.Command("go", "build", "-ldflags", "-X main.version=1.2.3", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build hush: %v\n%s", err, out)
	}
	got, err := exec.Command(bin, "--version").Output()
	if err != nil {
		t.Fatalf("hush --version: %v", err)
	}
	if string(got) != "hush version 1.2.3\n" {
		t.Errorf("hush --version = %q, want %q", got, "hush version 1.2.3\n")
	}
}

// TestAll checks that "hush all" uses the same settings as a single named
// check (CLI flags > per-check config > defaults) and runs the checks in the
// order of the config file.
func TestAll(t *testing.T) {
	dir := t.TempDir()
	config := `defaults:
  continue: true
  tail: 2
checks:
  zeta:
    cmd: printf 'z1\nz2\nz3\n'; exit 1
    label: Z
  alpha:
    cmd: printf 'E1\nok\nE2\nE3\n'; exit 1
    label: A
    grep: E
  mid:
    cmd: printf 'm1\nm2\nm3\nm4\n'; exit 1
    label: M
    tail: 3
  Warn:
    cmd: echo WARN w
    warn-pattern: WARN
`
	if err := os.WriteFile(filepath.Join(dir, ".hush.yaml"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		args []string
		want string
	}{
		{"config settings", []string{"all"}, `✗ Z
  z2
  z3
✗ A
  E2
  E3
✗ M
  m2
  m3
  m4
⚠ echo WARN w (1 warning)
  WARN w
✗ 1/4 checks passed
`},
		{"CLI flag wins", []string{"--tail", "1", "all"}, `✗ Z
  z3
✗ A
  E3
✗ M
  m4
⚠ echo WARN w (1 warning)
  WARN w
✗ 1/4 checks passed
`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run several times: a map-based order would differ between runs.
			for range 5 {
				out, code := runHushIn(t, dir, tt.args...)
				if out != tt.want {
					t.Fatalf("output:\n%s\nwant:\n%s", out, tt.want)
				}
				if code != 1 {
					t.Fatalf("exit code = %d, want 1", code)
				}
			}
		})
	}
}

// TestOutputIndent checks that a single command prints its output without an
// indent (each indented line costs about one token), and that batch indents
// each command's output to keep it apart from the summary lines.
func TestOutputIndent(t *testing.T) {
	out, _ := runHush(t, "printf 'a\\nb\\n'; exit 1")
	if want := "✗ printf 'a\\nb\\n'; exit 1\na\nb\n"; out != want {
		t.Errorf("single command:\ngot  %q\nwant %q", out, want)
	}

	out, _ = runHush(t, "batch", "--continue", "printf 'a\\n'; exit 1", "true")
	if want := "✗ printf 'a\\n'; exit 1\n  a\n✓ true\n✗ 1/2 checks passed\n"; out != want {
		t.Errorf("batch:\ngot  %q\nwant %q", out, want)
	}
}
