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
# ✓ pytest (1.8s)

# On failure, shows output
hush "pytest -x"
# ✗ pytest (0.4s)
#   FAILED tests/test_auth.py::test_login - assert 200 == 401

# Custom label
hush --label "unit tests" "pytest tests/unit"

# Filter output on failure
hush --tail 30 "npm test"          # last 30 lines
hush --grep "error|FAIL" "make"    # lines matching pattern
hush --head 20 "cargo build"       # first 20 lines

# Batch mode
hush batch "ruff check ." "ty check src/" "pytest -x"
# ✓ ruff (0.3s)
# ✓ ty (0.4s)
# ✓ 2/2 checks passed (2.4s)

# Continue on failure
hush batch --continue "ruff check ." "false" "pytest -x"
```

## Config File (optional)

For one-off commands, flags are enough. A config file is useful when you have multiple tools to run and want to bake in the right filters for each — so agents can just call `hush lint` or `hush all` without repeating flags every time.

Create `.hush.yaml` in your project root:

```yaml
checks:
  lint:
    cmd: ruff check .
  types:
    cmd: ty check src/
    grep: "error:"
  test:
    cmd: pytest -x
    tail: 40
```

Then run named checks:

```bash
hush lint          # run a single check
hush all           # run all checks in order
```

## Flags

| Flag | Description |
|------|-------------|
| `--label` | Custom label for the summary line |
| `--tail N` | Show only last N lines on failure |
| `--head N` | Show only first N lines on failure |
| `--grep PATTERN` | Filter output to matching lines |
| `--no-time` | Suppress duration |
| `--color` | Force colored output |
| `--no-color` | Disable colored output |

## Token Savings

The whole point of hush is saving context tokens when coding agents run shell commands. Here's how it stacks up across popular test runners:

**Passing tests** (hush output: single summary line)

| Test Runner | Raw Output | With hush | Saved |
|---|---|---|---|
| pytest (47 tests) | ~1,028 tokens | ~4 tokens | **99%** |
| Jest (89 tests + coverage) | ~886 tokens | ~5 tokens | **99%** |
| go test (14 packages) | ~196 tokens | ~6 tokens | **96%** |
| cargo test (51 tests) | ~683 tokens | ~5 tokens | **99%** |

**Failing tests** (hush output: summary + filtered error context)

| Test Runner | Raw Output | With hush | Saved |
|---|---|---|---|
| pytest (1 failure) | ~1,262 tokens | ~35 tokens | **97%** |
| Jest (1 failure) | ~1,000 tokens | ~63 tokens | **93%** |
| go test (1 pkg failure) | ~295 tokens | ~75 tokens | **74%** |

<details>
<summary><strong>Before and after — pytest with 47 passing tests</strong></summary>

```
# Without hush (50 lines, ~1,028 tokens)
============================= test session starts ==============================
platform linux -- Python 3.12.1, pytest-8.0.2, pluggy-1.4.0
...
tests/test_auth.py::test_login_success PASSED                            [  2%]
tests/test_auth.py::test_login_invalid_password PASSED                   [  4%]
  ... (47 more lines)
=============================== 47 passed in 3.42s ================================

# With hush (1 line, ~4 tokens)
✓ pytest (3.4s)
```

</details>

Run the full benchmark yourself:

```bash
bash benchmark/run.sh
```

## License

[MIT](LICENSE)
