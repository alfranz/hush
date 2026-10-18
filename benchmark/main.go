// Command benchmark measures how many tokens hush saves on real test runners.
//
// It builds one container image per test runner (benchmark/projects), runs
// each test suite without and with hush, and counts the tokens of both
// outputs with a Claude model through OpenRouter. It writes a table and all
// outputs to benchmark/results.
//
// Run it from the repository root:
//
//	OPENROUTER_API_KEY=... go run ./benchmark
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// runnerSpec is one test runner and the project it runs.
type runnerSpec struct {
	name    string // also the directory in benchmark/projects
	title   string
	command string
}

var runners = []runnerSpec{
	{"python", "pytest", "pytest"},
	{"node", "Jest", "npm test"},
	{"go", "go test", "go test ./..."},
	{"rust", "cargo test", "cargo test"},
}

// scenario is one way to run a test suite.
type scenario struct {
	name      string
	fail      bool     // set BENCH_FAIL=1, which breaks exactly one test
	hushFlags []string // hush flags before the command
}

var scenarios = []scenario{
	{"pass", false, nil},
	{"fail", true, nil},
	{"fail, --tail 20", true, []string{"--tail", "20"}},
}

// result is one measured row.
type result struct {
	runner     runnerSpec
	scenario   scenario
	rawTokens  int
	hushTokens int
	hushCmd    string
}

func main() {
	model := flag.String("model", "anthropic/claude-opus-5.5", "OpenRouter model whose tokenizer counts the tokens")
	outDir := flag.String("out", "benchmark/results", "directory for results.md and the outputs")
	flag.Parse()

	if err := run(*model, *outDir); err != nil {
		fmt.Fprintln(os.Stderr, "benchmark:", err)
		os.Exit(1)
	}
}

func run(model, outDir string) error {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return errors.New("OPENROUTER_API_KEY is not set")
	}
	if _, err := os.Stat("benchmark/projects"); err != nil {
		return errors.New("run from the repository root: go run ./benchmark")
	}
	cli := containerCLI()

	tmp, err := os.MkdirTemp("", "hush-bench")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	hushBin, err := filepath.Abs(filepath.Join(tmp, "hush"))
	if err != nil {
		return err
	}

	fmt.Println("Building hush for Linux ...")
	build := exec.Command("go", "build", "-o", hushBin, "./cmd/hush")
	build.Env = append(os.Environ(), "GOOS=linux", "GOARCH="+runtime.GOARCH, "CGO_ENABLED=0")
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("build hush: %v\n%s", err, out)
	}

	fmt.Println("Building images ...")
	if err := buildImages(cli); err != nil {
		return err
	}

	counter := &openRouterCounter{apiKey: apiKey, model: model}
	if err := os.MkdirAll(filepath.Join(outDir, "output"), 0o755); err != nil {
		return err
	}

	var results []result
	for _, r := range runners {
		for _, s := range scenarios {
			fmt.Printf("Running %s (%s) ...\n", r.title, s.name)
			res, err := measure(cli, hushBin, counter, outDir, r, s)
			if err != nil {
				return fmt.Errorf("%s (%s): %w", r.title, s.name, err)
			}
			results = append(results, res)
		}
	}

	table := resultsTable(results)
	fmt.Println()
	fmt.Print(table)
	report := resultsReport(table, model, counter.overhead)
	return os.WriteFile(filepath.Join(outDir, "results.md"), []byte(report), 0o644)
}

// measure runs one runner and scenario without and with hush, checks the
// exit codes, saves both outputs and counts their tokens.
func measure(cli, hushBin string, counter *openRouterCounter, outDir string, r runnerSpec, s scenario) (result, error) {
	image := "hush-bench-" + r.name
	var env []string
	if s.fail {
		env = []string{"BENCH_FAIL=1"}
	}

	raw, rawCode := runContainer(cli, image, hushBin, env, "sh", "-c", r.command+" 2>&1")
	hushArgs := append(append([]string{"hush"}, s.hushFlags...), r.command)
	hushOut, hushCode := runContainer(cli, image, hushBin, env, hushArgs...)

	if (rawCode != 0) != s.fail {
		return result{}, fmt.Errorf("raw run exited %d, want failure=%v:\n%s", rawCode, s.fail, raw)
	}
	if hushCode != rawCode {
		return result{}, fmt.Errorf("hush exited %d, raw run exited %d:\n%s", hushCode, rawCode, hushOut)
	}

	base := filepath.Join(outDir, "output", r.name+"-"+slug(s.name))
	if err := os.WriteFile(base+"-raw.txt", []byte(raw), 0o644); err != nil {
		return result{}, err
	}
	if err := os.WriteFile(base+"-hush.txt", []byte(hushOut), 0o644); err != nil {
		return result{}, err
	}

	rawTokens, err := counter.count(raw)
	if err != nil {
		return result{}, err
	}
	hushTokens, err := counter.count(hushOut)
	if err != nil {
		return result{}, err
	}
	return result{
		runner:     r,
		scenario:   s,
		rawTokens:  rawTokens,
		hushTokens: hushTokens,
		hushCmd:    shellJoin(hushArgs),
	}, nil
}

func buildImages(cli string) error {
	var wg sync.WaitGroup
	errs := make([]error, len(runners))
	for i, r := range runners {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dir := filepath.Join("benchmark", "projects", r.name)
			out, err := exec.Command(cli, "build", "-t", "hush-bench-"+r.name, dir).CombinedOutput()
			if err != nil {
				errs[i] = fmt.Errorf("build image %s: %v\n%s", r.name, err, out)
			}
		}()
	}
	wg.Wait()
	return errors.Join(errs...)
}

// runContainer runs args in a new container with hush mounted at
// /usr/local/bin/hush. It returns the combined output and the exit code.
func runContainer(cli, image, hushBin string, env []string, args ...string) (string, int) {
	cmdArgs := []string{"run", "--rm", "-v", hushBin + ":/usr/local/bin/hush:ro"}
	for _, e := range env {
		cmdArgs = append(cmdArgs, "-e", e)
	}
	cmdArgs = append(cmdArgs, image)
	cmdArgs = append(cmdArgs, args...)

	out, err := exec.Command(cli, cmdArgs...).CombinedOutput()
	code := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			code = exitErr.ExitCode()
		} else {
			code = -1
			out = append(out, []byte(err.Error())...)
		}
	}
	return string(out), code
}

// containerCLI picks docker or podman, like the integration tests.
func containerCLI() string {
	if cli := os.Getenv("HUSH_CONTAINER_CLI"); cli != "" {
		return cli
	}
	for _, cli := range []string{"docker", "podman"} {
		if _, err := exec.LookPath(cli); err == nil && exec.Command(cli, "info").Run() == nil {
			return cli
		}
	}
	return "docker"
}

// openRouterCounter counts tokens with a model's own tokenizer. OpenRouter
// has no endpoint that only counts tokens, so it sends the text as a prompt
// with max_tokens 1 and reads usage.prompt_tokens. That number includes a
// fixed overhead per message, which count subtracts.
type openRouterCounter struct {
	apiKey   string
	model    string
	overhead int // set on the first call to count
	measured bool
}

func (c *openRouterCounter) count(text string) (int, error) {
	if !c.measured {
		// "a" is one token, so the rest of its prompt_tokens is overhead.
		n, err := c.promptTokens("a")
		if err != nil {
			return 0, err
		}
		c.overhead = n - 1
		c.measured = true
	}
	if text == "" {
		return 0, nil
	}
	n, err := c.promptTokens(text)
	if err != nil {
		return 0, err
	}
	return n - c.overhead, nil
}

func (c *openRouterCounter) promptTokens(text string) (int, error) {
	body, err := json.Marshal(map[string]any{
		"model":      c.model,
		"max_tokens": 1,
		"messages":   []map[string]string{{"role": "user", "content": text}},
	})
	if err != nil {
		return 0, err
	}

	var lastErr error
	for attempt := range 3 {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}
		req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(body))
		if err != nil {
			return 0, err
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		data, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		var parsed struct {
			Usage struct {
				PromptTokens int `json:"prompt_tokens"`
			} `json:"usage"`
			Error *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(data, &parsed); err != nil {
			lastErr = fmt.Errorf("openrouter: HTTP %d: %s", resp.StatusCode, data)
			continue
		}
		if parsed.Error != nil {
			lastErr = fmt.Errorf("openrouter: HTTP %d: %s", resp.StatusCode, parsed.Error.Message)
			continue
		}
		if parsed.Usage.PromptTokens == 0 {
			lastErr = fmt.Errorf("openrouter: no prompt_tokens in response: %s", data)
			continue
		}
		return parsed.Usage.PromptTokens, nil
	}
	return 0, lastErr
}

func resultsTable(results []result) string {
	var b strings.Builder
	b.WriteString("| Runner | Scenario | hush command | Without hush | With hush | Saved |\n")
	b.WriteString("|---|---|---|---:|---:|---:|\n")
	for _, r := range results {
		saved := 0.0
		if r.rawTokens > 0 {
			saved = 100 * float64(r.rawTokens-r.hushTokens) / float64(r.rawTokens)
		}
		fmt.Fprintf(&b, "| %s | %s | `%s` | %d | %d | %.0f%% |\n",
			r.runner.title, r.scenario.name, r.hushCmd, r.rawTokens, r.hushTokens, saved)
	}
	return b.String()
}

func resultsReport(table, model string, overhead int) string {
	version, err := exec.Command("git", "describe", "--tags", "--always", "--dirty").Output()
	if err != nil {
		version = []byte("unknown")
	}
	return fmt.Sprintf(`# Benchmark results

Generated by `+"`go run ./benchmark`"+` on %s with hush %s.

%s
- **Without hush:** the test command with stdout and stderr combined, as an agent sees it.
- **With hush:** the output of the hush command in the table.
- **Scenarios:** "pass" runs all tests green. "fail" sets `+"`BENCH_FAIL=1`"+`, which breaks exactly one test.
- **Tokens:** counted with the tokenizer of %s through OpenRouter. The fixed overhead of %d tokens per message is subtracted.
- **Outputs:** every output is in [output/](output/).

The test projects are in [benchmark/projects](../projects). Timings in the outputs change between runs, so the counts can change by a few tokens.
`, time.Now().Format("2006-01-02"), strings.TrimSpace(string(version)), table, "`"+model+"`", overhead)
}

func slug(s string) string {
	s = strings.NewReplacer(",", "", "--", "", " ", "-").Replace(s)
	return strings.Trim(s, "-")
}

// shellJoin quotes args that contain spaces, for display.
func shellJoin(args []string) string {
	quoted := make([]string, len(args))
	for i, a := range args {
		if strings.ContainsAny(a, " \t") {
			a = `"` + a + `"`
		}
		quoted[i] = a
	}
	return strings.Join(quoted, " ")
}
