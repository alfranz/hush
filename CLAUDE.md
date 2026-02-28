# CLAUDE.md

## What is this?

`hush` is a context-efficient CLI command runner in Go. It wraps shell commands and prints a single ✓/✗ summary line, showing filtered output only on failure. Built for coding agents to reduce token waste.

## Commands

```bash
make build          # Build to bin/hush
make test           # Run tests with race detector (via hush)
make install        # go install
make integration-test  # Docker-based end-to-end tests (see below)
```

## Project Structure

- `cmd/hush/main.go` — Entry point
- `internal/runner/` — Command execution engine (sh -c, exit code extraction)
- `internal/filter/` — Output pipeline: ANSI stripping, grep, head/tail
- `internal/output/` — Formatted ✓/✗ printer (always plain text, no ANSI)
- `internal/config/` — .hush.yaml parsing via viper (checks + defaults)
- `internal/cli/` — Cobra commands: root (single cmd), batch, named checks

## Integration Tests

`tests/integration/` contains a Docker-based end-to-end suite that builds real language containers (Python/pytest, Node/node:test, Java/Maven) and runs hush against them. It covers every command and flag: root, batch, named checks, --label, --grep, --tail, --head, --continue.

**Do not run this on every change.** It is slow (builds Docker images, downloads Maven deps) and is not part of `make test`. Run it when:
- Changing CLI flags or adding new ones
- Touching filter, output, or runner internals
- Modifying config/named-check loading
- Before tagging a release

```bash
make integration-test
# or target a single suite for debugging:
go test -v -tags integration -run TestFlags ./tests/integration/...
```

## Key Decisions

- Combined stdout+stderr via `CombinedOutput()`
- `sh -c` execution for shell feature support
- Always plain text output — no ANSI codes, no color, no duration (optimized for agents)
- ANSI stripping always enabled (command output is cleaned)
- `os.Exit()` in RunE for exit code passthrough
- Module path: `github.com/alfranz/hush`
- Config supports `defaults` section for tail/head/grep/continue (precedence: CLI flags > per-check config > defaults)
