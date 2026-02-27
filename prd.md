# hush — Context-Efficient Command Runner

## Problem

Coding agents (Claude Code, Cursor, Copilot) waste significant context window tokens on verbose tool output. A passing test suite might dump 200+ lines that all say "everything is fine." A failing build buries the actual error in pages of noise. Developers end up either manually filtering output for the agent or losing productive session time to unnecessary context compaction.

There's no standard, tool-agnostic way to say "run this command, tell me if it passed, and only show me the details if it didn't."

## Solution

`hush` is a lightweight CLI that wraps any shell command. On success, it prints a single summary line. On failure, it dumps output — optionally filtered to just the actionable bits. It preserves exit codes, strips noise, and is designed to sit between dev tools and coding agents.

## Core Behaviour

```
$ hush "pytest -x"
✓ pytest (1.8s)

$ hush "pytest -x"
✗ pytest (0.4s)
  FAILED tests/test_auth.py::test_login - assert 200 == 401
```

Success = one line. Failure = the error and nothing else.

## Features

### Wrap any command
```
hush "make build"
hush "npm test"
hush "cargo clippy"
```
Runs the command in a subshell, captures stdout and stderr. Prints `✓ <label>` on exit code 0, `✗ <label>` + output on non-zero. Always passes through the original exit code.

### Labels
```
hush --label "unit tests" "pytest tests/unit"
```
Custom label for the summary line. Defaults to the command name.

### Output filtering on failure
```
hush --tail 30 "npm test"          # last 30 lines only
hush --grep "error|FAIL" "make"    # lines matching pattern
hush --head 20 "cargo build"       # first 20 lines (useful for first-error)
```
Multiple filters can be combined. Helps with tools that don't have built-in quiet modes.

### Timing
```
hush --time "cargo build"
✓ cargo build (4.2s)
```
Appends wall-clock duration. On by default, `--no-time` to disable.

### Batch mode
```
hush batch "ruff check ." "mypy src/" "pytest -x --agent-output"
✓ ruff (0.3s)
✓ mypy (2.1s)
✗ pytest (0.9s)
  FAILED tests/test_api.py::test_create - assert 201 == 400
```
Runs commands sequentially. Stops on first failure by default. `--continue` to run all regardless. Single summary if everything passes: `✓ 3/3 checks passed (3.3s)`.

### Config file
```yaml
# .hush.yaml
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
```
hush all              # run all checks in order
hush lint             # run a named check
hush test types       # run specific checks
```

### Agent mode
```
hush --agent "pytest -x"
```
Extra aggressive: strips ANSI escape codes, collapses multi-line tracebacks to one-liners, removes timestamps and progress bars. Designed for piping into LLM context.

### ANSI stripping
Enabled by default when stdout is not a TTY (i.e., when captured by an agent). `--color` to force colors, `--no-color` to force plain text.

## Non-goals

- Not a task runner (no dependency graph, no parallelism, no watch mode)
- Not a CI system (no caching, no artifacts, no reporting)
- Not tool-specific (no built-in knowledge of pytest, eslint, etc.)

## Implementation notes

- Single binary, no runtime dependencies. Written in Go.
- Should work on macOS, Linux, and WSL. Windows is nice-to-have.
- Config file is optional — zero-config by default.
- Installable via `brew`, `go install`.

## Success criteria

A developer can add `hush` to their coding agent workflow in under 2 minutes and immediately see fewer tokens wasted on passing checks. A `.hush.yaml` with 3-5 checks should be the only setup needed for most projects.
