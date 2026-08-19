<p align="center">
  <img src="docs/logo/goninja-logo-full.png" alt="goninja" width="720">
</p>

<p align="center">
  <a href="https://github.com/caspel26/goninja/actions/workflows/go.yml"><img src="https://github.com/caspel26/goninja/actions/workflows/go.yml/badge.svg" alt="Build Status"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/caspel26/goninja" alt="License"></a>
  <img src="https://img.shields.io/badge/go-1.25%2B-00ADD8?logo=go&logoColor=white" alt="Go version">
  <img src="https://img.shields.io/badge/status-pre--alpha-orange" alt="Status: pre-alpha">
</p>

<p align="center">
  <b>Generate typed, validated REST APIs from annotated Go structs — no reflection, no runtime magic.</b>
</p>

Code-first Go framework for generating complete CRUD REST APIs from
annotated structs: routing, input/output validation, serialization,
OpenAPI, filters, pagination.

Define your model once, annotate its fields, and generate the rest.

```go
// models/book.go
type Book struct {
    ID        string    `gorm:"primaryKey;type:uuid" goninja:"list,retrieve"`
    Title     string    `gorm:"size:120;not null" goninja:"list,retrieve,create,update" validate:"required,max=120"`
    AuthorID  string    `goninja:"list,retrieve,create,update,filter"`
    Price     float64   `goninja:"list,retrieve,create,update,filter" validate:"min=0"`
    Published bool      `goninja:"list,retrieve,create,update,filter"`
}
```

```console
$ goninja generate
```

This generates typed schemas, handlers, database queries, and an OpenAPI
fragment for the model, mounted onto a standard `net/http` mux — no
runtime reflection, no hidden magic: the generated code is plain Go you
can read, debug, and step through.

> **Status: pre-alpha.** The API above is the design target. What exists
> today is an early prototype — see [Status](#status).

## Why goninja

- **Code-first**: the struct is the single source of truth. No separate
  schema files, no config YAML.
- **Generated, not reflected**: `goninja generate` writes real `.go`
  files you commit. Errors show up at compile time, not in production.
- **Plain `net/http`**: no framework lock-in for routing.
- **Safe by default**: output schemas are always separate from your
  database model, so a sensitive field can't leak into a response just
  because it exists on the struct.
- **Built to be extended**: override any generated method, hook into
  create/update/delete, plug in your own auth middleware — all without
  touching generated files.

## Status

`goninja` is early and not yet usable for real projects. The current
engine lives under [`internal/codegen`](internal/codegen) and
[`examples/prototype`](examples/prototype): it parses `goninja`-tagged
struct fields and generates typed output types plus GORM-backed CRUD
(`net/http` handlers, transaction-aware queries, automatic preloading of
relations on retrieve, `validate`-tag-driven input validation with
per-field 422 responses, `filter`-tag-driven filtering with limit/offset
pagination and ordering behind a `{items, total, limit, offset}` envelope)
for any number of models, verified end to end against Postgres. A model's
ID field can be `int64` (DB auto-increment) or `string` (a UUID goninja
generates itself) — `examples/prototype`'s models use UUID IDs. Every
generated resource also emits an OpenAPI 3.0 fragment from the same
annotations, groupable under custom tags per resource; `goninja.MountDocs`
merges every registered resource's fragment and serves it as JSON plus a
docs UI — Swagger UI or ReDoc ship built in, both fully embedded with no
external CDN, and the `DocsUI` interface it takes isn't hardcoded to
either.

Try it (needs a running Postgres):

```console
$ export PROTOTYPE_DSN="host=localhost user=$(whoami) dbname=goninja_prototype sslmode=disable"
$ make generate-prototype   # writes examples/prototype/internal/api
$ make run-prototype        # serves /tasks, /authors, /books on :8080
$ curl "localhost:8080/books?published=true&price_min=10&order=-created_at&limit=20"
$ open http://localhost:8080/docs   # Swagger UI over the merged OpenAPI doc
```

Auth and hooks are designed but not yet built.

### Generated docs UI

`/docs` renders the merged OpenAPI document with either UI, both fully
embedded (no external CDN):

<table>
<tr>
<td width="50%">

**Swagger UI** (`goninja.SwaggerUI()`, the default)

<img src="docs/screenshots/swagger-ui.png" alt="Swagger UI listing every route, grouped by model">

</td>
<td width="50%">

**ReDoc** (`goninja.ReDoc()`, a drop-in swap)

<img src="docs/screenshots/redoc.png" alt="ReDoc three-pane layout with sidebar nav and response samples">

</td>
</tr>
</table>

Every operation is expandable into its request/response schema, complete
with example values generated from the model's fields:

<img src="docs/screenshots/swagger-ui-operation.png" alt="Swagger UI showing an expanded POST /books operation with its request body schema and response example" width="720">

## Contributing

The project is pre-alpha; the implementation plan is the source of truth
for scope and sequencing. Open an issue before a large PR.

## License

[MIT](LICENSE)
