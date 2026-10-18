<div align="center">

# hush

**Context-efficient command runner for coding agents.**

Wraps any shell command and prints a single `✓`/`✗` summary line.\
On success, shows only the summary. On failure, shows filtered output. Preserves exit codes.

[![CI](https://github.com/alfranz/hush/actions/workflows/ci.yml/badge.svg)](https://github.com/alfranz/hush/actions/workflows/ci.yml)
[![Release](https://github.com/alfranz/hush/actions/workflows/release.yml/badge.svg)](https://github.com/alfranz/hush/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/alfranz/hush)](https://goreportcard.com/report/github.com/alfranz/hush)
[![GoDoc](https://pkg.go.dev/badge/github.com/alfranz/hush)](https://pkg.go.dev/github.com/alfranz/hush)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

</div>

---

## Install

```bash
go install github.com/alfranz/hush/cmd/hush@latest
```

This places the `hush` binary in your `$GOPATH/bin` (usually `~/go/bin`). Make sure that directory is in your `$PATH`:

```bash
# Check if it's already there
which hush

# If not, add to your shell profile (~/.zshrc, ~/.bashrc, etc.)
export PATH="$HOME/go/bin:$PATH"
```

## Usage

```bash
# Basic usage
hush "pytest -x"
# ✓ pytest -x

# On failure, shows full output (traceback, assertion details, etc.)
hush "pytest -x"
# ✗ pytest -x
#   =================================== FAILURES ===================================
#   __________________________________ test_login __________________________________
#
#       def test_login():
#           response = client.get("/auth/login")
#   >       assert response.status_code == 200
#   E       AssertionError: assert 401 == 200
#
#   tests/test_auth.py:10: AssertionError
#   =========================== short test summary info ============================
#   FAILED tests/test_auth.py::test_login - AssertionError: assert 401 == 200
#   ============================== 1 failed in 0.06s ===============================

# Without quotes: hush runs the program directly with its arguments
hush go test -race ./...
# ✓ go test -race ./...

# Custom label
hush --label "unit tests" "pytest tests/unit"

# Filter output on failure (use with care — see note below)
hush --tail 30 "npm test"          # last 30 lines
hush --grep "error|FAIL" "make"    # lines matching pattern
hush --head 20 "cargo build"       # first 20 lines

# Warning-aware success (exit code still 0)
hush --warn-pattern "warning TS[0-9]+" --warn-tail 5 "tsc --noEmit"
# ⚠ tsc --noEmit (3 warnings)
#   src/auth.ts:42:3 - warning TS6133: 'userId' is declared but its value is never read.
#   src/auth.ts:57:9 - warning TS6133: 'token' is declared but its value is never read.
#   src/db.ts:12:1 - warning TS6192: All imports in import declaration are unused.

# Batch mode
hush batch "ruff check ." "ty check src/" "pytest -x"
# ✓ ruff check .
# ✓ ty check src/
# ✓ pytest -x
# ✓ 3/3 checks passed

# Continue on failure
hush batch --continue "ruff check ." "false" "pytest -x"
```

### Two command forms

| Form | Example | How hush runs it |
|------|---------|------------------|
| One quoted argument | `hush "make check && make lint"` | With `sh -c`. Pipes, `&&`, `;`, globs and `$VARS` work. |
| Program and arguments | `hush make check` | Directly, with the arguments as given. No extra shell. |

For simple commands, both forms give the same result.

### Flag placement

Put hush flags **before** the command. Hush reads its own flags until the first argument that is not a flag. Everything after it belongs to the command. `time`, `nice` and `env` work the same way.

```bash
hush --tail 20 go test -race ./...    # --tail is for hush, -race is for go
hush --label t -- go test ./...       # -- ends the hush flags
hush "make check" --label check       # error: flags after the command
```

## Config File (optional)

For one-off commands, flags are enough. A config file is useful when you have multiple tools to run and want to bake in the right filters for each — so agents can just call `hush lint` or `hush all` without repeating flags every time.

Create `.hush.yaml` in your project root:

```yaml
defaults:
  continue: true

checks:
  lint:
    cmd: ruff check .
  types:
    cmd: ty check src/
    grep: "^error"
    warn-pattern: "^warning"
    warn-tail: 5
  test:
    cmd: pytest -x
    tail: 40
```

Then run named checks:

```bash
hush lint          # run a single check
hush all           # run all checks in order
```

Settings in `defaults` apply to all commands (root, batch, and named checks) unless overridden by per-check config or CLI flags. Precedence: **CLI flags > per-check config > defaults**.

## Flags

| Flag | Description |
|------|-------------|
| `--label` | Custom label for the summary line. The default label is the command, cut to 40 characters. Not supported with `batch` |
| `--tail N` | Show only last N lines on failure |
| `--head N` | Show only first N lines on failure |
| `--grep PATTERN` | Filter output to matching lines |
| `--warn-pattern REGEX` | On success, match warning lines and emit `⚠` with details |
| `--warn-tail N` | On warning-qualified success, show last N warning lines (default: 10) |
| `--continue` | Continue running after a failure (batch/all) |
| `--version`, `-v` | Print the hush version |

> **Note on `--grep` and test failures:** By default (no flags), hush prints the full command output on failure — including tracebacks, assertion diffs, and source context. This gives agents the most information to debug with. Use `--grep` and `--tail` primarily for **linters and build tools** that produce high-volume output. For **test runners** (pytest, Jest, go test), the unfiltered output is usually what the agent needs to fix the issue. A `--grep "FAIL"` on pytest output, for example, strips away the traceback and assertion details, leaving only the one-line summary.

## Token Savings

The whole point of hush is saving context tokens when coding agents run shell commands. The numbers below come from `go run ./benchmark`. It runs small real projects with each test runner in Docker, without and with hush, and counts tokens with Claude's tokenizer. See [benchmark/results](benchmark/results/results.md) for the method and every output.

**Passing tests** (hush output: one summary line)

| Test runner | Without hush | With hush | Saved |
|---|---:|---:|---:|
| pytest (39 tests) | 144 tokens | 6 tokens | **96%** |
| Jest via `npm test` (27 tests) | 127 tokens | 6 tokens | **95%** |
| go test (5 packages) | 98 tokens | 8 tokens | **92%** |
| cargo test (27 tests) | 604 tokens | 7 tokens | **99%** |

A passing run always costs hush about 6–8 tokens. Larger test suites print more, so they save more.

**Failing tests** (1 failing test)

| Test runner | Without hush | `hush` | `hush --tail 20` |
|---|---:|---:|---:|
| pytest | 391 tokens | 428 tokens (+9%) | 360 tokens (−8%) |
| Jest via `npm test` | 312 tokens | 341 tokens (+9%) | 254 tokens (−19%) |
| go test | 174 tokens | 191 tokens (+10%) | 191 tokens (+10%) |
| cargo test | 704 tokens | 765 tokens (+9%) | 273 tokens (−61%) |

On a failure, hush shows the full output by default, so the agent gets the whole traceback. This costs about 10% **more** than the raw output: hush adds the `✗` line and indents each output line by 2 spaces. Filters such as `--tail` save tokens on failures, but they can cut context the agent needs (see the note on `--grep` above).

<details>
<summary><strong>Before and after: pytest with 39 passing tests</strong></summary>

```
# Without hush (144 tokens)
============================= test session starts ==============================
platform linux -- Python 3.12.15, pytest-8.3.4, pluggy-1.7.0
rootdir: /app
collected 39 items

tests/test_cart.py .............                                         [ 33%]
tests/test_discounts.py .............                                    [ 66%]
tests/test_pricing.py .............                                      [100%]

============================== 39 passed in 0.03s ==============================

# With hush (6 tokens)
✓ pytest
```

</details>

Run the benchmark yourself (needs Docker and an [OpenRouter](https://openrouter.ai) API key, costs a few cents):

```bash
OPENROUTER_API_KEY=... go run ./benchmark
```

## License

[MIT](LICENSE)
