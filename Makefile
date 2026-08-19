.PHONY: build test vet fmt cover generate-prototype run-prototype

build:
	go build ./...

test:
	go test ./...

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
