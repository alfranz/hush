# CLAUDE.md

## What is this?

`hush` is a context-efficient CLI command runner in Go. It wraps shell commands and prints a single ✓/✗ summary line, showing filtered output only on failure. Built for coding agents to reduce token waste.

## Commands

```bash
make build          # Build to bin/hush
make test           # Run tests with race detector (via hush)
make install        # go install
make integration-test  # Docker-based end-to-end tests (see below)
go run ./benchmark     # Token benchmark on real test runners (see below)
```

## Project Structure

- `cmd/hush/main.go` — Entry point
- `internal/runner/` — Command execution engine (sh -c or direct exec, exit code extraction)
- `internal/filter/` — Output pipeline: ANSI stripping, grep, head/tail
- `internal/output/` — Formatted ✓/✗ printer (always plain text, no ANSI)
- `internal/config/` — .hush.yaml parsing via viper (checks + defaults)
- `internal/cli/` — Cobra commands: root (single cmd), batch, named checks
- `benchmark/` — Token benchmark: test projects in `projects/`, results in `results/`

## Integration Tests

`tests/integration/` contains a Docker-based end-to-end suite that builds real language containers (Python/pytest, Node/node:test, Java/Maven) and runs hush against them. It covers every command and flag: root, batch, named checks, --label, --grep, --tail, --head, --continue, --warn-pattern, --warn-tail. It uses Docker, or Podman with `HUSH_CONTAINER_CLI=podman`.

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

## Benchmark

`benchmark/` measures the tokens hush saves. It runs small real projects (pytest, Jest, go test, cargo test) in Docker, without and with hush. It counts tokens with a Claude model through OpenRouter. It writes `benchmark/results/results.md` and every output to `benchmark/results/output/`.

```bash
OPENROUTER_API_KEY=... go run ./benchmark
```

It costs a few cents per run and takes a few minutes the first time (image builds). Run it again when the output format changes (labels, summary lines), and update the token tables in `README.md` and `docs/index.html` from `results.md`.

## Release Process

Creating a GitHub release creates the tag. The tag starts `.github/workflows/release.yml`. goreleaser (`.goreleaser.yaml`) then builds the binaries, uploads them to the release, and keeps the release notes. It sets `main.version` through its default ldflags.

**Ask the user before you create the release.** It publishes the release and cannot be undone cleanly.

1. **Pick the version.** Use semantic versioning with a `v` prefix (`v0.3.0`). Before 1.0: a breaking change or a new feature is a minor bump. A fix only is a patch bump. Confirm the version with the user.
2. **Check main.**
   - All PRs for the release are merged. `git pull` on main.
   - CI on the latest main commit passes: `gh run list --branch main --limit 3`.
   - `make integration-test` passes on main (needs Docker, takes about 2 minutes).
   - If the docs site (`docs/index.html`, https://alfranz.github.io/hush/) changed, look at it in a browser.
3. **Write `notes.md`** in the repo root. Git ignores it. Base it on the PRs merged since the last tag: `git log --oneline --merges vPREV..origin/main`, then `gh pr view <n>` for each one. List breaking changes first, with a before/after example. Then list the changes with their PR numbers. Check every example against a fresh build. GitHub adds the list of merged PRs below the notes.
4. **Create the release** (after the user says yes):
   ```bash
   gh release create vX.Y.Z --target main --generate-notes --notes-file notes.md
   ```
5. **Verify.**
   - Wait for the Release workflow: `gh run list --workflow release.yml --limit 1`.
   - The release has 6 archives (`hush_X.Y.Z_{darwin,linux}_{amd64,arm64}.tar.gz`, `hush_X.Y.Z_windows_{amd64,arm64}.zip`) and `checksums.txt`: `gh release view vX.Y.Z --json assets --jq '.assets[].name'`.
   - A downloaded binary prints the right version:
     ```bash
     gh release download vX.Y.Z -p '*darwin_arm64.tar.gz' -D "$TMPDIR/hush-rel" && tar -xzf "$TMPDIR"/hush-rel/*.tar.gz -C "$TMPDIR/hush-rel" && "$TMPDIR/hush-rel/hush" --version
     # expect: hush version X.Y.Z
     ```

## Key Decisions

- Combined stdout+stderr via `CombinedOutput()`
- One argument runs with `sh -c` (shell features). Several arguments run the program directly, without a shell
- Root command reads hush flags only before the command (`SetInterspersed(false)`)
- Always plain text output — no ANSI codes, no color, no duration (optimized for agents)
- ANSI stripping always enabled (command output is cleaned)
- `os.Exit()` in RunE for exit code passthrough
- Module path: `github.com/alfranz/hush`
- Config supports `defaults` section for tail/head/grep/warn-pattern/warn-tail/continue (precedence: CLI flags > per-check config > defaults). `hush all` runs the checks in file order with the same settings as `hush <check>`
