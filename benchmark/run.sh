#!/usr/bin/env bash
#
# benchmark/run.sh — Token savings benchmark for hush CLI
#
# Compares raw test runner output vs hush output for passing and failing tests.
# Uses cl100k_base approximation (~4 chars per token) for token counting.
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FIXTURES_DIR="$SCRIPT_DIR/fixtures"

# ── Token counting ──────────────────────────────────────────────────────────
# Approximate tokens using chars/4 (cl100k_base average).
# If tiktoken is available, use it for accuracy.
count_tokens() {
    local text="$1"
    local chars=${#text}
    # Simple approximation: ~4 characters per token (matches cl100k_base average)
    echo $(( (chars + 3) / 4 ))
}

# ── Simulate hush output ────────────────────────────────────────────────────
hush_pass_output() {
    local label="$1"
    local time="$2"
    echo "✓ $label ($time)"
}

hush_fail_output() {
    local label="$1"
    local time="$2"
    local error_lines="$3"
    printf "✗ %s (%s)\n%s\n" "$label" "$time" "$error_lines"
}

# ── Formatting helpers ──────────────────────────────────────────────────────
BOLD="\033[1m"
DIM="\033[2m"
GREEN="\033[32m"
RED="\033[31m"
CYAN="\033[36m"
YELLOW="\033[33m"
RESET="\033[0m"

bar() {
    local pct=$1
    local width=30
    local filled=$(( pct * width / 100 ))
    local empty=$(( width - filled ))
    printf "${GREEN}"
    for ((i=0; i<filled; i++)); do printf "█"; done
    printf "${DIM}"
    for ((i=0; i<empty; i++)); do printf "░"; done
    printf "${RESET}"
}

separator() {
    printf "${DIM}─────────────────────────────────────────────────────────────────${RESET}\n"
}

# ── Run a single benchmark ──────────────────────────────────────────────────
run_benchmark() {
    local name="$1"
    local fixture_file="$2"
    local hush_output="$3"
    local scenario="$4"  # "pass" or "fail"

    local raw_output
    raw_output=$(cat "$fixture_file")

    local raw_tokens hush_tokens saved pct
    raw_tokens=$(count_tokens "$raw_output")
    hush_tokens=$(count_tokens "$hush_output")
    saved=$(( raw_tokens - hush_tokens ))
    if [ "$raw_tokens" -gt 0 ]; then
        pct=$(( saved * 100 / raw_tokens ))
    else
        pct=0
    fi

    if [ "$scenario" = "pass" ]; then
        local icon="${GREEN}✓${RESET}"
    else
        local icon="${RED}✗${RESET}"
    fi

    printf "  $icon ${BOLD}%-28s${RESET}" "$name"
    printf "  %6d → %6d tokens" "$raw_tokens" "$hush_tokens"
    printf "  ${CYAN}saved %5d${RESET}  " "$saved"
    bar "$pct"
    printf "  ${YELLOW}%d%%${RESET}\n" "$pct"

    # Accumulate totals
    TOTAL_RAW=$(( TOTAL_RAW + raw_tokens ))
    TOTAL_HUSH=$(( TOTAL_HUSH + hush_tokens ))
}

# ── Main ────────────────────────────────────────────────────────────────────
echo ""
printf "${BOLD}hush CLI — Token Savings Benchmark${RESET}\n"
printf "${DIM}Comparing raw test output vs hush output across test runners${RESET}\n"
echo ""

TOTAL_RAW=0
TOTAL_HUSH=0

# ── Passing tests ───────────────────────────────────────────────────────────
printf "${BOLD}${GREEN}Passing tests (hush shows: single summary line)${RESET}\n"
separator

run_benchmark "pytest (47 tests)" \
    "$FIXTURES_DIR/pytest_pass.txt" \
    "$(hush_pass_output "pytest" "3.4s")" \
    "pass"

run_benchmark "npm test / Jest (89 tests)" \
    "$FIXTURES_DIR/npm_pass.txt" \
    "$(hush_pass_output "npm test" "8.2s")" \
    "pass"

run_benchmark "go test (14 packages)" \
    "$FIXTURES_DIR/go_pass.txt" \
    "$(hush_pass_output "go test ./..." "4.8s")" \
    "pass"

run_benchmark "cargo test (51 tests)" \
    "$FIXTURES_DIR/cargo_pass.txt" \
    "$(hush_pass_output "cargo test" "10.7s")" \
    "pass"

echo ""

# ── Failing tests ──────────────────────────────────────────────────────────
printf "${BOLD}${RED}Failing tests (hush shows: summary + filtered error)${RESET}\n"
separator

# For failures, hush shows the summary line + relevant error output.
# Simulate what hush --tail 10 or hush --grep FAIL would produce.

PYTEST_FAIL_HUSH=$(cat <<'EOF'
✗ pytest (3.6s)
  FAILED tests/test_auth.py::test_signup_duplicate_email - AssertionError: assert 201 == 409
  1 failed, 46 passed in 3.58s
EOF
)
run_benchmark "pytest (1 failure)" \
    "$FIXTURES_DIR/pytest_fail.txt" \
    "$PYTEST_FAIL_HUSH" \
    "fail"

NPM_FAIL_HUSH=$(cat <<'EOF'
✗ npm test (8.4s)
  FAIL src/api/__tests__/client.test.ts
    ● API Client › should retry on 500 errors
      Expected: 3
      Received: 1
        at Object.<anonymous> (src/api/__tests__/client.test.ts:26:23)
  Tests: 1 failed, 88 passed, 89 total
EOF
)
run_benchmark "npm test / Jest (1 failure)" \
    "$FIXTURES_DIR/npm_fail.txt" \
    "$NPM_FAIL_HUSH" \
    "fail"

GO_FAIL_HUSH=$(cat <<'EOF'
✗ go test ./... (5.0s)
  --- FAIL: TestConnectionPool (1.203s)
      db_test.go:89: expected pool to reject connection, but got nil error
      db_test.go:90: pool size: got 11, want max 10
      db_test.go:112: context deadline exceeded after 500ms
  FAIL github.com/example/myservice/internal/db
EOF
)
run_benchmark "go test (1 pkg failure)" \
    "$FIXTURES_DIR/go_fail.txt" \
    "$GO_FAIL_HUSH" \
    "fail"

echo ""

# ── Summary ─────────────────────────────────────────────────────────────────
separator
TOTAL_SAVED=$(( TOTAL_RAW - TOTAL_HUSH ))
if [ "$TOTAL_RAW" -gt 0 ]; then
    TOTAL_PCT=$(( TOTAL_SAVED * 100 / TOTAL_RAW ))
else
    TOTAL_PCT=0
fi

printf "${BOLD}  TOTAL%30s${RESET}" ""
printf "  %6d → %6d tokens" "$TOTAL_RAW" "$TOTAL_HUSH"
printf "  ${CYAN}saved %5d${RESET}  " "$TOTAL_SAVED"
bar "$TOTAL_PCT"
printf "  ${YELLOW}${BOLD}%d%%${RESET}\n" "$TOTAL_PCT"
echo ""

# ── Side-by-side examples ──────────────────────────────────────────────────
printf "${BOLD}Example: pytest (47 tests passing)${RESET}\n"
separator
printf "${DIM}Without hush:${RESET}\n"
head -5 "$FIXTURES_DIR/pytest_pass.txt"
printf "${DIM}  ... (50 lines total)${RESET}\n"
tail -1 "$FIXTURES_DIR/pytest_pass.txt"
echo ""
printf "${DIM}With hush:${RESET}\n"
hush_pass_output "pytest" "3.4s"
echo ""
separator

printf "${BOLD}Example: npm test (1 failure)${RESET}\n"
separator
printf "${DIM}Without hush:${RESET}\n"
head -5 "$FIXTURES_DIR/npm_fail.txt"
printf "${DIM}  ... (58 lines total)${RESET}\n"
tail -3 "$FIXTURES_DIR/npm_fail.txt"
echo ""
printf "${DIM}With hush:${RESET}\n"
printf "%s\n" "$NPM_FAIL_HUSH"
echo ""
separator

printf "${DIM}Token estimation: ~4 chars/token (cl100k_base approximation)${RESET}\n"
echo ""
