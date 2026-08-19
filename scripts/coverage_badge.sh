#!/usr/bin/env bash
# Writes coverage-badge.json (shields.io "endpoint" schema) from the most
# recent scripts/coverage.sh run, for the README's coverage badge:
#   https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/caspel26/goninja/main/coverage-badge.json
#
# coverage-badge.json is a tracked file, not CI-generated: main requires a
# PR to merge and blocks direct pushes (see CLAUDE.md), so there's no branch
# protection-safe way for CI to commit an updated value back on every push
# without an external coverage service. Regenerate it locally
# (`make cover && ./scripts/coverage_badge.sh`) and commit it like any other
# generated file when coverage moves meaningfully.
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

profile="coverage.out"
if [[ ! -f "$profile" ]]; then
	echo "coverage.out not found — run scripts/coverage.sh first" >&2
	exit 1
fi

total="$(go tool cover -func="$profile" | awk '/^total:/ {print $3}' | tr -d '%')"

color="red"
if awk -v t="$total" 'BEGIN { exit !(t >= 80) }'; then
	color="brightgreen"
elif awk -v t="$total" 'BEGIN { exit !(t >= 70) }'; then
	color="green"
elif awk -v t="$total" 'BEGIN { exit !(t >= 50) }'; then
	color="yellow"
fi

cat >coverage-badge.json <<JSON
{
  "schemaVersion": 1,
  "label": "coverage",
  "message": "${total}%",
  "color": "${color}"
}
JSON

echo "wrote coverage-badge.json (${total}%, ${color})"
