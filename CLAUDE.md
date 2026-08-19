# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Workflow rule

**Never run `git commit` or `git push` without the user's explicit approval first** — ask before each one, even mid-task and even if earlier commits/pushes in the same session were approved. Staging changes (`git add`) and showing a diff/status is fine without asking.

## What this is

`goninja` is a pre-alpha, code-first Go framework for generating complete
CRUD REST APIs from annotated structs (routing, validation, serialization,
OpenAPI, filters, pagination). The full design — public API shape, phased
build plan, and rationale — lives in
[goninja-implementation-plan.md](goninja-implementation-plan.md); read it
before making architectural changes, since most decisions in this repo
trace back to it.

**Current phase: Phase 2 (GORM and queries)**, per section 6 of the plan.
Phase 0's decision gate passed and Phase 1's exit criterion (engine
generalizes to a second model) is met — both documented in the plan next
to their phase sections. Phase 2's exit criterion (CRUD end-to-end on real
Postgres, with automatic `Preload` on relation fields) is also met, proven
by `examples/prototype`. Still not built: validation, a pluggable
`ErrorMapper`, filters/pagination, OpenAPI, auth, hooks — those are
Phases 3-6. Don't over-build beyond what the current phase calls for
without checking the plan.

## Commands

```sh
make build               # go build ./...
make test                # go test ./...
make vet                 # go vet ./...
make fmt                 # gofmt -l . (lists unformatted files)
make generate-prototype  # regenerate examples/prototype/internal/api from examples/prototype/models
make run-prototype       # generate-prototype, then run the example server on :8080 (needs PROTOTYPE_DSN, see below)
```

Run a single test: `go test ./internal/codegen/ -run TestParseModels -v`

The `goninja` CLI itself (`cmd/goninja`) is invoked directly for anything
outside the example, e.g.:

```sh
go run ./cmd/goninja generate \
  -models <dir> -out <dir> -package <name> \
  -models-import <full import path of the models package>
```

`examples/prototype` needs a running Postgres and `PROTOTYPE_DSN` set,
e.g.:

```sh
export PROTOTYPE_DSN="host=localhost user=$(whoami) dbname=goninja_prototype sslmode=disable"
```

It calls `db.AutoMigrate` itself on startup — no separate migration step.

## Architecture

Three pieces:

1. **`internal/codegen`** — the generator engine, used by anything that
   calls `goninja generate`:
   - [parse.go](internal/codegen/parse.go): `go/parser`/`go/ast`-based
     scanner that reads a directory of Go source, finds struct types with
     at least one `goninja`-tagged field, and produces an IR (`ir.go`).
   - [ir.go](internal/codegen/ir.go): `Model`/`Field` — the intermediate
     representation decoupled from both parsing and templating, so it can
     later be reused for OpenAPI generation too (plan Phase 5). Also
     where `Field.IsRelation` lives — the heuristic (non-scalar Go type =
     relation) that drives automatic `Preload`.
   - [generate.go](internal/codegen/generate.go) +
     [templates/](internal/codegen/templates/): renders the IR through
     `text/template`, formats with `go/format`, and writes one
     `<model>_generated.go` file per model plus a shared
     `runtime_generated.go` (helpers like `writeJSON`/`idFromPath`/
     `mapError` — kept in one shared file specifically to avoid
     duplicate-symbol errors when multiple models are generated into the
     same package).

2. **root package `goninja`** (`resource.go`, `errors.go`) — the runtime
   support generated code depends on: `BaseResource` (embedded by every
   generated `<Model>Resource`), its transaction-aware `DB(ctx)`,
   `InTransaction`/`WithTx`/`TxFromContext`, and the `NotFound` error
   type. Generated code imports this package as
   `github.com/caspel26/goninja`.

3. **`cmd/goninja`** — the CLI (`goninja generate`), a thin flag-parsing
   wrapper around `internal/codegen`.

**`examples/prototype`** exercises the whole loop end to end against real
Postgres and is the concrete proof for the Phase 0-2 exit criteria:
- `models/task.go`, `models/author.go`, `models/book.go` — three models;
  `Book.Author` is a real GORM belongs-to relation (`AuthorID` + `Author`
  field, inferred by GORM's naming convention, no explicit `foreignKey`
  tag needed) proving automatic `Preload` on `Retrieve`.
- `internal/api/` — **generated output, never hand-edit**; regenerate via
  `make generate-prototype`. Lives under `internal/` deliberately (Go's
  compiler-enforced non-importability from outside the module), since
  generated code is meant to be used only through the consuming app, not
  imported by others.
- `main.go` — a real `net/http` server, opens Postgres via
  `PROTOTYPE_DSN`, `AutoMigrate`s all three models, registers all three
  generated resources.

When changing the generator's output shape (schemas, handler signatures,
route registration, query behavior), the templates in
`internal/codegen/templates/*.tmpl` are the source of truth — edit those,
then run `make generate-prototype` and rebuild to verify the emitted code
still compiles and the example server still serves correctly, rather than
hand-editing anything under `examples/prototype/internal/api`.

### Key generator conventions

- Tag vocabulary is the field-level `goninja:"..."` struct tag (comma-
  separated verbs: `list`, `retrieve`, `create`, `update`); `Field.HasTag`
  in `ir.go` is the single place that interprets it.
- `list` and `retrieve` are separate schemas by design (plan section
  5.1/5.5) — list stays lean and never preloads relations; retrieve is the
  full detail view and preloads every relation field it carries. This is
  the framework's guarantee against N+1, not an implementation detail —
  preserve it when touching `List`/`Retrieve` codegen.
- Output schemas (`<Model>ListSchema`, `<Model>RetrieveSchema`,
  `<Model>CreateSchema`, `<Model>UpdateSchema`) are always separate Go
  structs from the model itself, never the model reused directly — so a
  field can't leak into a response just because it exists on the struct.
  A relation field on `RetrieveSchema` is nested as the related model's
  own `RetrieveSchema` too, for the same reason (see `model.go.tmpl`).
- `Field.IsRelation` (in `ir.go`) currently assumes single-value (not
  slice/has-many) relations for schema nesting — a slice relation field
  would generate but its schema conversion wouldn't be correct yet.
- The generated package imports the models package by import path
  (`-models-import`), so generated code and model structs stay in
  separate packages (`ModelsImport`/`ModelsPkg` in `generate.go`).
- Every generated `<Model>Resource` assumes the model has a field
  literally named `ID` of type `int64` — hardcoded in the templates
  (`Retrieve(ctx, id int64)` etc.), not yet derived from the model's
  actual primary-key field/type.
