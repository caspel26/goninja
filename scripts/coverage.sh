#!/usr/bin/env bash
# Runs the framework's test suite (root goninja package + internal/codegen —
# the two packages Fase 7's coverage item targets, per CLAUDE.md/the
# implementation plan) with coverage instrumentation, prints the total
# percentage, and — if a threshold is passed as $1 — fails when total
# coverage is below it. Used by `make cover` locally and by CI
# (.github/workflows/go.yml) to enforce the agreed minimum.
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

profile="coverage.out" # gitignored; scripts/coverage_badge.sh reads this back

go test . ./internal/codegen -coverprofile="$profile" -covermode=atomic

total="$(go tool cover -func="$profile" | awk '/^total:/ {print $3}' | tr -d '%')"
echo "total coverage: ${total}%"

if [ "${1:-}" != "" ]; then
	threshold="$1"
	if awk -v t="$total" -v th="$threshold" 'BEGIN { exit !(t < th) }'; then
		echo "coverage ${total}% is below the required ${threshold}%" >&2
		exit 1
	fi
fi
