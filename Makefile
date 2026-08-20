.PHONY: build test vet fmt cover bench bench-baseline bench-check bench-profile generate-prototype run-prototype docs docs-serve

build:
	go build ./...

test:
	go test ./...

# Runs the benchmark suite (examples/prototype/benchmark_test.go): base
# serialization, filter-clause building, and automatic-Preload cost.
# -run=^$ skips the package's regular tests during a bench invocation.
bench:
	go test ./examples/prototype/... -bench=. -benchmem -run='^$$'

# Overwrites the committed regression baseline (scripts/testdata/bench-baseline.txt)
# with a fresh run. Only do this deliberately — after confirming a change in
# the numbers is expected — and commit the result; `make bench-check` is what
# compares against it on every PR.
bench-baseline:
	go test ./examples/prototype/... -bench=. -benchmem -run='^$$' -count=10 > scripts/testdata/bench-baseline.txt
	@echo "wrote scripts/testdata/bench-baseline.txt - review and commit it"

# Fails if a benchmark regressed beyond THRESHOLD_PCT (default 25) against the
# baseline; see scripts/bench-regression.sh for how it tells a real
# regression from run-to-run noise.
bench-check:
	./scripts/bench-regression.sh

# Profiles the benchmark suite to find what's actually worth optimizing, not
# just whether it regressed. Writes pprof files under profiles/ (gitignored,
# via *.out/*.test) and prints a quick top-10 for CPU and allocated memory.
# Follow up with the interactive view, e.g.:
#   go tool pprof -http=:8080 profiles/bench.test profiles/cpu.out
#   go tool pprof -http=:8080 -alloc_space profiles/bench.test profiles/mem.out
bench-profile:
	@mkdir -p profiles
	go test ./examples/prototype -bench=. -benchmem -run='^$$' \
		-cpuprofile=profiles/cpu.out -memprofile=profiles/mem.out \
		-o profiles/bench.test
	@echo
	@echo "--- CPU: top 10 by flat time ---"
	go tool pprof -top -nodecount=10 profiles/bench.test profiles/cpu.out
	@echo
	@echo "--- Memory: top 10 by allocated space ---"
	go tool pprof -top -nodecount=10 -alloc_space profiles/bench.test profiles/mem.out
	@echo
	@echo "Interactive flamegraph/callgraph:"
	@echo "  go tool pprof -http=:8080 profiles/bench.test profiles/cpu.out"

vet:
	go vet ./...

fmt:
	gofmt -l .

# Coverage for the goninja root package + internal/codegen (Fase 7's
# coverage item, see CLAUDE.md/goninja-implementation-plan.md) — pass a
# threshold to fail below it, matching CI: make cover THRESHOLD=70
cover:
	./scripts/coverage.sh $(THRESHOLD)

# Regenerates coverage-badge.json (the README coverage badge) from the last
# `make cover` run — commit it when coverage moves meaningfully.
cover-badge: cover
	./scripts/coverage_badge.sh

generate-prototype:
	go run ./cmd/goninja generate \
		-models ./examples/prototype/models \
		-out ./examples/prototype/internal/api \
		-package api \
		-models-import github.com/caspel26/goninja/examples/prototype/models

run-prototype: generate-prototype
	cd examples/prototype && go run .

# Builds every documented version into ./public: the working tree as /dev/,
# each release tag as /vX.Y/, and the newest release at the root. This is what
# CI publishes; BASE_URL=http://localhost:1313 for a local preview.
docs:
	./scripts/build-docs.sh

# Live preview of the working tree only, without the version subdirectories.
docs-serve:
	cd docs-site && hugo serve --buildDrafts
