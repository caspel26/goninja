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
# Also writes a self-contained HTML report (reports/bench-report.html) for
# easier reading than raw logs - not committed, meant to be picked up as a
# CI artifact (see .github/workflows/bench.yml).
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
REPORT_DIR="$REPO_ROOT/reports"
REPORT_FILE="$REPORT_DIR/bench-report.html"
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
regressions=()
while IFS=, read -r name _ _ _ _ delta _; do
  # Only rows with a real +N%/-N% delta are candidates - this skips both
  # header lines, benchstat's own "~" (not statistically significant), and
  # the plain-text goos/goarch/pkg/cpu preamble go test prints before the
  # CSV tables.
  [[ "$delta" =~ ^[+-][0-9] ]] || continue
  [[ "$name" == "geomean" ]] && continue
  case "$delta" in
  -*) continue ;; # smaller/faster is never a regression
  *) ;;           # a "+" delta falls through to the threshold check below
  esac
  pct="${delta#+}"
  pct="${pct%\%}"
  if awk "BEGIN{exit !($pct > $THRESHOLD_PCT)}"; then
    echo "bench-regression: $name regressed by $delta (threshold ${THRESHOLD_PCT}%)" >&2
    regressions+=("$name regressed by $delta")
    fail=1
  fi
done <<<"$CSV"

mkdir -p "$REPORT_DIR"
if [[ "$fail" -ne 0 ]]; then
  status_label="FAIL"
  status_class="fail"
else
  status_label="PASS"
  status_class="pass"
fi

regressions_html="<p>No regressions beyond ${THRESHOLD_PCT}%.</p>"
if [[ "${#regressions[@]}" -gt 0 ]]; then
  regressions_html="<ul class=\"regressions\">"
  for r in "${regressions[@]}"; do
    regressions_html+="<li>$(printf '%s' "$r" | sed 's/&/\&amp;/g; s/</\&lt;/g; s/>/\&gt;/g')</li>"
  done
  regressions_html+="</ul>"
fi

escaped_text="$(printf '%s' "$TEXT" | sed 's/&/\&amp;/g; s/</\&lt;/g; s/>/\&gt;/g')"
generated_at="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

cat <<HTML >"$REPORT_FILE"
<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>goninja benchmark report</title>
<style>
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; margin: 2rem; background:#0d1117; color:#c9d1d9; }
  h1 { font-size: 1.4rem; }
  .status { display:inline-block; padding:.25rem .6rem; border-radius:.4rem; font-weight:600; }
  .pass { background:#238636; color:#fff; }
  .fail { background:#da3633; color:#fff; }
  pre { background:#161b22; padding:1rem; border-radius:.5rem; overflow-x:auto; border:1px solid #30363d; }
  .regressions li { color:#f85149; }
  .meta { color:#8b949e; font-size:.85rem; margin-bottom:1rem; }
</style>
</head>
<body>
<h1>goninja benchmark regression report</h1>
<p class="meta">Generated ${generated_at} &middot; threshold ${THRESHOLD_PCT}% &middot; count=${BENCH_COUNT}</p>
<p><span class="status ${status_class}">${status_label}</span></p>
${regressions_html}
<pre>${escaped_text}</pre>
</body>
</html>
HTML

echo "bench-regression: wrote $REPORT_FILE" >&2

if [[ "$fail" -ne 0 ]]; then
  echo "bench-regression: one or more benchmarks regressed beyond ${THRESHOLD_PCT}%" >&2
  [[ -n "${GITHUB_STEP_SUMMARY:-}" ]] && echo "**Result: regression(s) beyond ${THRESHOLD_PCT}%**" >>"$GITHUB_STEP_SUMMARY"
  exit 1
fi

echo "bench-regression: no regressions beyond ${THRESHOLD_PCT}%" >&2
[[ -n "${GITHUB_STEP_SUMMARY:-}" ]] && echo "**Result: no regressions beyond ${THRESHOLD_PCT}%**" >>"$GITHUB_STEP_SUMMARY"
