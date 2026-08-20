.PHONY: build test vet fmt cover bench generate-prototype run-prototype docs docs-serve

build:
	go build ./...

test:
	go test ./...

# Runs the benchmark suite (examples/prototype/benchmark_test.go): base
# serialization, filter-clause building, and automatic-Preload cost.
# -run=^$ skips the package's regular tests during a bench invocation.
bench:
	go test ./examples/prototype/... -bench=. -benchmem -run=^$

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
