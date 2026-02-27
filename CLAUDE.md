# CLAUDE.md

## What is this?

`hush` is a context-efficient CLI command runner in Go. It wraps shell commands and prints a single ✓/✗ summary line, showing filtered output only on failure. Built for coding agents to reduce token waste.

## Commands

```bash
make build          # Build to bin/hush
make test           # Run tests with race detector
make install        # go install
go test -race ./... # Run all tests
```

## Project Structure

- `cmd/hush/main.go` — Entry point
- `internal/runner/` — Command execution engine (sh -c, exit code extraction)
- `internal/filter/` — Output pipeline: ANSI stripping, agent mode, grep, head/tail
- `internal/output/` — Formatted ✓/✗ printer with TTY-aware color
- `internal/config/` — .hush.yaml parsing via viper
- `internal/cli/` — Cobra commands: root (single cmd), batch, named checks

## Key Decisions

- Combined stdout+stderr via `CombinedOutput()`
- `sh -c` execution for shell feature support
- Raw ANSI codes for color — no color library
- `os.Exit()` in RunE for exit code passthrough
- Module path: `github.com/alfranz/hush`
