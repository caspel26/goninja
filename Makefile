.PHONY: build test vet fmt generate-prototype run-prototype

build:
	go build ./...

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l .

generate-prototype:
	go run ./cmd/goninja generate \
		-models ./examples/prototype/models \
		-out ./examples/prototype/internal/generated \
		-package generated \
		-models-import github.com/caspel26/goninja/examples/prototype/models

run-prototype: generate-prototype
	cd examples/prototype && go run .
