# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`goninja` is a pre-alpha, code-first Go framework for generating complete
CRUD REST APIs from annotated structs (routing, validation, serialization,
OpenAPI, filters, pagination). The full design — public API shape, phased
build plan, and rationale — lives in
[goninja-implementation-plan.md](goninja-implementation-plan.md); read it
before making architectural changes, since most decisions in this repo
trace back to it.

**Current phase: Phase 1 (code-generation engine)**, per section 6 of the
plan. Phase 0's decision gate passed (documented in the plan next to the
Phase 0 section) and its parser/IR/template/CLI became the base Phase 1
builds on, rather than being thrown away. What exists today is still
deliberately minimal — no ORM, a single tag vocabulary (`list`,
`retrieve`, `create`), in-memory storage — because ORM integration,
validation, filters/pagination, and OpenAPI are separate later phases.
Don't over-build beyond what the current phase calls for without checking
the plan.

## Commands

```sh
make build               # go build ./...
make test                # go test ./...
make vet                 # go vet ./...
make fmt                 # gofmt -l . (lists unformatted files)
make generate-prototype  # regenerate examples/prototype/internal/api from examples/prototype/models
make run-prototype       # generate-prototype, then run the example server on :8080
```

Run a single test: `go test ./internal/codegen/ -run TestParseModels -v`

The `goninja` CLI itself (`cmd/goninja`) is invoked directly for anything
outside the example, e.g.:

```sh
go run ./cmd/goninja generate \
  -models <dir> -out <dir> -package <name> \
  -models-import <full import path of the models package>
```

## Architecture

Two independent pieces:

1. **`internal/codegen`** — the generator engine, used by anything that
   calls `goninja generate`:
   - [parse.go](internal/codegen/parse.go): `go/parser`/`go/ast`-based
     scanner that reads a directory of Go source, finds struct types with
     at least one `goninja`-tagged field, and produces an IR (`ir.go`).
   - [ir.go](internal/codegen/ir.go): `Model`/`Field` — the intermediate
     representation decoupled from both parsing and templating, so it can
     later be reused for OpenAPI generation too (see plan section 6,
     Phase 1).
   - [generate.go](internal/codegen/generate.go) +
     [templates/](internal/codegen/templates/): renders the IR through
     `text/template`, formats with `go/format`, and writes one
     `<model>_generated.go` file per model plus a shared
     `runtime_generated.go` (helpers like `writeJSON`/`idFromPath` — kept
     in one shared file specifically to avoid duplicate-symbol errors when
     multiple models are generated into the same package).

2. **`cmd/goninja`** — the CLI (`goninja generate`), a thin flag-parsing
   wrapper around `internal/codegen`.

**`examples/prototype`** exercises the whole loop end to end and is the
concrete proof for both the Phase 0 decision gate and the Phase 1 exit
criterion ("works on a second, different model"):
- `models/task.go`, `models/author.go` — two annotated models, distinct in
  field count/types.
- `internal/api/` — **generated output, never hand-edit**; regenerate via
  `make generate-prototype`. Lives under `internal/` deliberately (Go's
  compiler-enforced non-importability from outside the module), since
  generated code is meant to be used only through the consuming app, not
  imported by others.
- `main.go` — a real `net/http` server wired to both generated resources.

When changing the generator's output shape (schemas, handler signatures,
route registration), the templates in
`internal/codegen/templates/*.tmpl` are the source of truth — edit those,
then run `make generate-prototype` and rebuild to verify the emitted code
still compiles and the example server still serves correctly, rather than
hand-editing anything under `examples/prototype/internal/api`.

### Key generator conventions

- Tag vocabulary is the field-level `goninja:"..."` struct tag (comma-
  separated verbs, e.g. `goninja:"list,retrieve,create"`); `Field.HasTag`
  in `ir.go` is the single place that interprets it.
- `list` and `retrieve` are separate schemas by design (see plan section
  5.1/5.5) — list stays lean, retrieve is the full detail view. This is
  preserved even in the Phase 0 prototype.
- Output schemas (`<Model>ListSchema`, `<Model>RetrieveSchema`,
  `<Model>CreateSchema`) are always separate Go structs from the model
  itself, never the model reused directly — this is intentional so a
  field can't leak into a response just because it exists on the struct.
- The generated package imports the models package by import path
  (`-models-import`), so generated code and model structs stay in
  separate packages (`ModelsImport`/`ModelsPkg` in `generate.go`).
