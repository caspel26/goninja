#!/usr/bin/env bash
# Runs the framework's test suite (root goninja package, internal/codegen,
# and the openapi/docsui/id/router/goninjatest subpackages split out of
# root — per CLAUDE.md/the implementation plan) with coverage
# instrumentation, prints the total percentage, and — if a threshold is
# passed as $1 — fails when total coverage is below it. Used by `make
# cover` locally and by CI (.github/workflows/go.yml) to enforce the agreed
# minimum. Deliberately excludes cmd/goninja (thin flag-parsing wrapper),
# examples/prototype (exercised manually against real Postgres, not unit
# tested — its own goninjatest-based test is coverage for goninjatest, not
# for the example; benchmark_test.go there is likewise never covered, since
# `go test` without -bench never runs a benchmark's body), and the
# adapters/* nested modules (each its own go.mod, outside this module's
# `go test` invocation entirely — see their own go.mod for how to test
# them, e.g. `cd adapters/gin && go test ./...`).
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

profile="coverage.out" # gitignored; scripts/coverage_badge.sh reads this back

go test . ./internal/codegen ./docsui ./id ./openapi ./router ./goninjatest -coverprofile="$profile" -covermode=atomic

total="$(go tool cover -func="$profile" | awk '/^total:/ {print $3}' | tr -d '%')"
echo "total coverage: ${total}%"

if [[ "${1:-}" != "" ]]; then
	threshold="$1"
	if awk -v t="$total" -v th="$threshold" 'BEGIN { exit !(t < th) }'; then
		echo "coverage ${total}% is below the required ${threshold}%" >&2
		exit 1
	fi
fi
