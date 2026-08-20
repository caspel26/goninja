#!/usr/bin/env bash
#
# Fails if any benchmark in examples/prototype regresses beyond THRESHOLD_PCT
# against the committed baseline (scripts/testdata/bench-baseline.txt),
# using benchstat to tell a real regression from run-to-run noise.
#
# benchstat's own significance test (alpha=0.05 by default) already collapses
# a noisy, insignificant difference down to "~" — this script only adds a
# magnitude gate on top of that, so a statistically real but tiny delta
# (e.g. +2%) doesn't fail the build.
#
# Shared CI runners are noisy neighbors: expect more variance there than
# locally, hence the wider default threshold and higher -count below versus
# a quick manual `make bench`.
#
# Usage:
#   scripts/bench-regression.sh
#
# Environment:
#   THRESHOLD_PCT   fail if a benchmark's sec/op, B/op, or allocs/op grows by
#                    more than this many percent (default 25)
#   BENCH_COUNT     benchmark repetitions per run, passed to `go test -count`
#                    (default 10 - benchstat needs >= 6 for a confidence
#                    interval, more reduces noise further)

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BASELINE="$REPO_ROOT/scripts/testdata/bench-baseline.txt"
THRESHOLD_PCT="${THRESHOLD_PCT:-25}"
BENCH_COUNT="${BENCH_COUNT:-10}"
# Pinned pseudo-version (x/perf has no tagged releases). Requires Go >= 1.26,
# so `go run` may transparently download a newer toolchain the first time
# this runs (GOTOOLCHAIN=auto, the Go default) - that's expected, not a bug.
BENCHSTAT_PKG="golang.org/x/perf/cmd/benchstat@v0.0.0-20260819171926-ebcb4798430d"

if [[ ! -f "$BASELINE" ]]; then
  echo "bench-regression: no baseline at $BASELINE - run 'make bench-baseline' first" >&2
  exit 1
fi

cd "$REPO_ROOT"
CURRENT="$(mktemp)"
trap 'rm -f "$CURRENT"' EXIT

echo "bench-regression: running benchmarks (count=$BENCH_COUNT)..." >&2
go test ./examples/prototype/... -bench=. -benchmem -run=^$ -count="$BENCH_COUNT" >"$CURRENT"

echo "bench-regression: comparing against baseline (threshold ${THRESHOLD_PCT}%)..." >&2
TEXT="$(go run "$BENCHSTAT_PKG" "$BASELINE" "$CURRENT")"
CSV="$(go run "$BENCHSTAT_PKG" -format csv "$BASELINE" "$CURRENT")"
echo "$TEXT"

# In CI, also render the same comparison into the job summary (visible in the
# Actions run UI) so a reviewer can see the actual numbers without digging
# through raw logs - not just the final pass/fail.
if [[ -n "${GITHUB_STEP_SUMMARY:-}" ]]; then
  {
    echo "## Benchmark results"
    echo
    echo "baseline: \`scripts/testdata/bench-baseline.txt\` vs this run, threshold ${THRESHOLD_PCT}%"
    echo
    echo '```'
    echo "$TEXT"
    echo '```'
  } >>"$GITHUB_STEP_SUMMARY"
fi

fail=0
while IFS=, read -r name _ _ _ _ delta _; do
  # Only rows with a real +N%/-N% delta are candidates - this skips both
  # header lines, benchstat's own "~" (not statistically significant), and
  # the plain-text goos/goarch/pkg/cpu preamble go test prints before the
  # CSV tables.
  [[ "$delta" =~ ^[+-][0-9] ]] || continue
  [[ "$name" == "geomean" ]] && continue
  case "$delta" in
  -*) continue ;; # smaller/faster is never a regression
  esac
  pct="${delta#+}"
  pct="${pct%\%}"
  if awk "BEGIN{exit !($pct > $THRESHOLD_PCT)}"; then
    echo "bench-regression: $name regressed by $delta (threshold ${THRESHOLD_PCT}%)" >&2
    fail=1
  fi
done <<<"$CSV"

if [[ "$fail" -ne 0 ]]; then
  echo "bench-regression: one or more benchmarks regressed beyond ${THRESHOLD_PCT}%" >&2
  [[ -n "${GITHUB_STEP_SUMMARY:-}" ]] && echo "**Result: regression(s) beyond ${THRESHOLD_PCT}%**" >>"$GITHUB_STEP_SUMMARY"
  exit 1
fi

echo "bench-regression: no regressions beyond ${THRESHOLD_PCT}%" >&2
[[ -n "${GITHUB_STEP_SUMMARY:-}" ]] && echo "**Result: no regressions beyond ${THRESHOLD_PCT}%**" >>"$GITHUB_STEP_SUMMARY"
