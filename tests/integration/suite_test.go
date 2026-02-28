//go:build integration

package integration_test

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
)

const binaryPath = "suites/hush"

type imageSpec struct {
	tag      string
	file     string
	buildArg string // empty if no --build-arg needed
}

var images = []imageSpec{
	{"hush-python-pass", "suites/python/Dockerfile", "SCENARIO=pass"},
	{"hush-python-fail", "suites/python/Dockerfile", "SCENARIO=fail"},
	{"hush-node-pass", "suites/node/Dockerfile", "SCENARIO=pass"},
	{"hush-node-fail", "suites/node/Dockerfile", "SCENARIO=fail"},
	{"hush-java-pass", "suites/java/Dockerfile", "SCENARIO=pass"},
	{"hush-java-fail", "suites/java/Dockerfile", "SCENARIO=fail"},
	{"hush-named", "suites/named/Dockerfile", ""},
}

func TestMain(m *testing.M) {
	// Cross-compile hush for Linux
	build := exec.Command("go", "build", "-o", binaryPath, "../../cmd/hush")
	build.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64", "CGO_ENABLED=0")
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to build hush:", err)
		os.Exit(1)
	}
	defer os.Remove(binaryPath)

	// Build all Docker images in parallel
	type buildErr struct {
		tag string
		out []byte
		err error
	}
	var wg sync.WaitGroup
	errCh := make(chan buildErr, len(images))

	for _, img := range images {
		img := img
		wg.Add(1)
		go func() {
			defer wg.Done()
			args := []string{"build", "-f", img.file, "-t", img.tag}
			if img.buildArg != "" {
				args = append(args, "--build-arg", img.buildArg)
			}
			args = append(args, "suites/")
			out, err := exec.Command("docker", args...).CombinedOutput()
			if err != nil {
				errCh <- buildErr{img.tag, out, err}
			}
		}()
	}
	wg.Wait()
	close(errCh)

	failed := false
	for e := range errCh {
		fmt.Fprintf(os.Stderr, "docker build %s failed: %v\n%s\n", e.tag, e.err, e.out)
		failed = true
	}
	if failed {
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// run executes docker run --rm <image> <cmd...> and returns combined output and exit code.
func run(image string, cmd ...string) (string, int) {
	args := append([]string{"run", "--rm", image}, cmd...)
	out, err := exec.Command("docker", args...).CombinedOutput()
	exitCode := 0
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			exitCode = e.ExitCode()
		}
	}
	return string(out), exitCode
}

// assert checks exit code, an optional required substring, and an optional forbidden substring.
func assert(t *testing.T, out string, code, wantExit int, wantOut, wantNoOut string) {
	t.Helper()
	if code != wantExit {
		t.Errorf("exit %d, want %d\noutput:\n%s", code, wantExit, out)
	}
	if wantOut != "" && !strings.Contains(out, wantOut) {
		t.Errorf("output missing %q\noutput:\n%s", wantOut, out)
	}
	if wantNoOut != "" && strings.Contains(out, wantNoOut) {
		t.Errorf("output should not contain %q\noutput:\n%s", wantNoOut, out)
	}
}

// TestLanguageSuites verifies hush correctly wraps real language test runners.
func TestLanguageSuites(t *testing.T) {
	cases := []struct {
		name     string
		image    string
		cmd      []string
		wantExit int
		wantOut  string
	}{
		{"python/pass", "hush-python-pass", []string{"hush", "pytest tests/ -q"}, 0, "✓"},
		{"python/fail", "hush-python-fail", []string{"hush", "pytest tests/ -q"}, 1, "✗"},
		{"node/pass", "hush-node-pass", []string{"hush", "node --test"}, 0, "✓"},
		{"node/fail", "hush-node-fail", []string{"hush", "node --test"}, 1, "✗"},
		{"java/pass", "hush-java-pass", []string{"hush", "mvn test -q"}, 0, "✓"},
		{"java/fail", "hush-java-fail", []string{"hush", "mvn test -q"}, 1, "✗"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			out, code := run(tc.image, tc.cmd...)
			assert(t, out, code, tc.wantExit, tc.wantOut, "")
		})
	}
}

// TestFlags covers every root-command flag.
// Filter flags (--grep, --tail, --head) use predictable sh commands so assertions are exact.
func TestFlags(t *testing.T) {
	cases := []struct {
		name      string
		image     string
		cmd       []string
		wantExit  int
		wantOut   string
		wantNoOut string
	}{
		// --label: custom label appears on the summary line instead of the derived one
		{
			"label",
			"hush-python-pass",
			[]string{"hush", "--label", "mycheck", "echo ok"},
			0, "✓ mycheck", "",
		},
		// --no-time: (Xs) duration is omitted on success
		{
			"no-time/pass",
			"hush-python-pass",
			[]string{"hush", "--no-time", "echo ok"},
			0, "✓ echo", "(",
		},
		// --no-time: (Xs) duration is omitted on failure
		{
			"no-time/fail",
			"hush-python-pass",
			[]string{"hush", "--no-time", "exit 1"},
			1, "✗ exit", "(",
		},
		// --color: forces ANSI escape codes in hush's own summary line
		{
			"color",
			"hush-python-pass",
			[]string{"hush", "--color", "echo ok"},
			0, "\x1b[", "",
		},
		// Without --no-color: pytest's ANSI output passes through into the failure block
		{
			"raw-ansi",
			"hush-python-fail",
			[]string{"hush", "pytest tests/ -q --color=yes"},
			1, "\x1b[", "",
		},
		// --no-color: strips ANSI from command output and suppresses it from hush's line
		{
			"no-color",
			"hush-python-fail",
			[]string{"hush", "--no-color", "pytest tests/ -q --color=yes"},
			1, "✗", "\x1b[",
		},
		// --grep: failure block contains only lines matching the pattern
		{
			"grep",
			"hush-python-pass",
			[]string{"hush", "--grep", "ERROR", "printf 'INFO: ok\\nERROR: bad\\nINFO: fine\\n' && exit 1"},
			1, "ERROR: bad", "INFO:",
		},
		// --tail N: failure block shows only the last N lines
		{
			"tail",
			"hush-python-pass",
			[]string{"hush", "--tail", "2", "printf 'a\\nb\\nc\\nd\\ne\\n' && exit 1"},
			1, "  d\n  e", "  a",
		},
		// --head N: failure block shows only the first N lines
		{
			"head",
			"hush-python-pass",
			[]string{"hush", "--head", "2", "printf 'a\\nb\\nc\\nd\\ne\\n' && exit 1"},
			1, "  a\n  b", "  e",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			out, code := run(tc.image, tc.cmd...)
			assert(t, out, code, tc.wantExit, tc.wantOut, tc.wantNoOut)
		})
	}
}

// TestBatch covers the batch subcommand and its flags.
func TestBatch(t *testing.T) {
	cases := []struct {
		name      string
		image     string
		cmd       []string
		wantExit  int
		wantOut   string
		wantNoOut string
	}{
		// All pass: batch summary line is printed
		{
			"pass",
			"hush-python-pass",
			[]string{"hush", "batch", "echo one", "echo two"},
			0, "✓ 2/2 checks passed", "",
		},
		// First fails without --continue: stops early, no summary line
		{
			"fail-stop",
			"hush-python-pass",
			[]string{"hush", "batch", "exit 1", "echo two"},
			1, "✗ exit", "checks passed",
		},
		// --continue: runs all commands, summary reflects partial pass count
		{
			"continue",
			"hush-python-pass",
			[]string{"hush", "batch", "--continue", "exit 1", "echo two"},
			1, "✗ 1/2 checks passed", "",
		},
		// --no-time: batch summary line omits the (Xs) duration
		{
			"no-time",
			"hush-python-pass",
			[]string{"hush", "batch", "--no-time", "echo one", "echo two"},
			0, "✓ 2/2 checks passed", "(",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			out, code := run(tc.image, tc.cmd...)
			assert(t, out, code, tc.wantExit, tc.wantOut, tc.wantNoOut)
		})
	}
}

// TestNamedChecks covers named checks defined in .hush.yaml.
func TestNamedChecks(t *testing.T) {
	cases := []struct {
		name      string
		cmd       []string
		wantExit  int
		wantOut   string
		wantNoOut string
	}{
		// Named check with label: YAML label field appears on the summary line
		{"pass-with-label", []string{"hush", "passingtest"}, 0, "✓ mytest", ""},
		// Named check with grep: failure block filtered to FAILED lines only
		{"fail-with-grep", []string{"hush", "failingtest"}, 1, "FAILED", ""},
		// hush all: runs all checks, exits non-zero because one check fails
		{"all", []string{"hush", "all"}, 1, "✗", ""},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			out, code := run("hush-named", tc.cmd...)
			assert(t, out, code, tc.wantExit, tc.wantOut, tc.wantNoOut)
		})
	}
}
