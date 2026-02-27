# hush

Context-efficient command runner for coding agents.

Wraps any shell command and prints a single `✓`/`✗` summary line. On success, shows only the summary. On failure, shows filtered output. Preserves exit codes.

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

# Agent mode (strips ANSI, collapses tracebacks, removes noise)
hush --agent "pytest -x"

# Batch mode
hush batch "ruff check ." "mypy src/" "pytest -x"
# ✓ ruff (0.3s)
# ✓ mypy (2.1s)
# ✓ 2/2 checks passed (2.4s)

# Continue on failure
hush batch --continue "ruff check ." "false" "pytest -x"
```

## Config File

Create `.hush.yaml` in your project root:

```yaml
checks:
  lint:
    cmd: ruff check .
  types:
    cmd: mypy src/
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
| `--agent` | Agent mode: strip noise for LLM context |
| `--no-time` | Suppress duration |
| `--color` | Force colored output |
| `--no-color` | Disable colored output |
