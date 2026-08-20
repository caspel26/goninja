<p align="center">
  <img src="docs/logo/goninja-logo-full.png" alt="goninja" width="720">
</p>

<p align="center">
  <a href="https://github.com/caspel26/goninja/actions/workflows/go.yml"><img src="https://github.com/caspel26/goninja/actions/workflows/go.yml/badge.svg" alt="Build Status"></a>
  <a href="https://sonarcloud.io/summary/new_code?id=caspel26_goninja"><img src="https://sonarcloud.io/api/project_badges/measure?project=caspel26_goninja&metric=alert_status" alt="SonarQube Quality Gate"></a>
  <img src="https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/caspel26/goninja/main/coverage-badge.json" alt="Coverage">
  <a href="LICENSE"><img src="https://img.shields.io/github/license/caspel26/goninja" alt="License"></a>
  <img src="https://img.shields.io/badge/go-1.25%2B-00ADD8?logo=go&logoColor=white" alt="Go version">
  <img src="https://img.shields.io/badge/status-pre--alpha-orange" alt="Status: pre-alpha">
</p>

<p align="center">
  <b>Generate typed, validated REST APIs from annotated Go structs — no reflection, no runtime magic.</b>
</p>

<p align="center">
  <a href="https://caspel26.github.io/goninja/docs/getting-started/">Full documentation →</a>
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

<p align="center">
<code>🔴&nbsp;🟡&nbsp;🟢&nbsp;&nbsp;<b>zsh</b></code>

```console
$ go install github.com/caspel26/goninja/cmd/goninja@latest
$ goninja generate
  ✓ parsed models (3 found)
  ✓ wrote internal/api/book_generated.go
  ✓ wrote internal/api/author_generated.go
```
</p>

That writes typed schemas, handlers, database queries, and an OpenAPI
fragment for the model — plain Go under `internal/api`, readable and
debuggable, nothing reflected at runtime. Wiring it into a server is a few
lines, no framework of its own to learn:

```go
// main.go
package main

import (
    "net/http"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"

    "github.com/caspel26/goninja"
    "github.com/caspel26/goninja/docsui"
    "myapp/internal/api"
    "myapp/models"
)

func main() {
    db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    db.AutoMigrate(&models.Author{}, &models.Book{}) // goninja doesn't generate migrations

    mux := http.NewServeMux()
    app := goninja.NewAPI("Bookstore API", "0.1.0")

    app.Mount(mux,
        api.NewAuthorResource(db),
        api.NewBookResource(db),
    )
    app.MountDocs(mux, "/docs", docsui.SwaggerUI())

    http.ListenAndServe(":8080", mux)
}
```

That's a full `net/http` server: `GET/POST /books`, `GET/PUT/DELETE
/books/{id}`, filtering, pagination, validation, and `/docs` — all from the
struct at the top. See the
[Getting Started guide](https://caspel26.github.io/goninja/docs/getting-started/)
for the full walkthrough, including `-watch` mode.

> **Status: pre-alpha.** The API above is the design target. What exists
> today is an early prototype — see [Status](#status).

---

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
  touching generated files. See
  [Extending a Resource](https://caspel26.github.io/goninja/docs/extending/)
  for hooks, custom actions, auth, and testing your own resources.

How does this compare to Huma, gocrud, or gorest? See the
[full comparison](https://caspel26.github.io/goninja/docs/comparison/) —
short version: those resolve your model at request time via generics and
reflection (or leave the handlers to you); goninja trades that runtime
flexibility for handlers, queries, and OpenAPI fragments that exist as
ordinary Go source before the binary is even built.

---

## Status

`goninja` is early and not yet usable for real projects. The current
engine lives under [`internal/codegen`](internal/codegen) and
[`examples/prototype`](examples/prototype): it parses `goninja`-tagged
struct fields and generates typed output types plus GORM-backed CRUD
(`net/http` handlers, transaction-aware queries, automatic preloading of
belongs-to and has-many relations on retrieve, `validate`-tag-driven input
validation with per-field 422 responses, `filter`-tag-driven filtering with
limit/offset pagination and ordering behind a `{items, total, limit,
offset}` envelope) for any number of models, verified end to end against
Postgres.

Try it (needs a running Postgres):

<p align="center">
<code>🔴&nbsp;🟡&nbsp;🟢&nbsp;&nbsp;<b>zsh</b></code>

```console
$ export PROTOTYPE_DSN="host=localhost user=$(whoami) dbname=goninja_prototype sslmode=disable"
$ make generate-prototype   # writes examples/prototype/internal/api
$ make run-prototype        # serves /tasks, /authors, /books on :8080
$ curl "localhost:8080/books?published=true&price_min=10&order=-created_at&limit=20"
$ open http://localhost:8080/docs   # Swagger UI over the merged OpenAPI doc
```
</p>

Hooks, per-method overriding, custom path/route config, a global default
auth policy + middleware, a per-field choice between nesting a relation and
exposing just its ID, and end-to-end resource testing via `goninjatest` are
all built — see the [docs](https://caspel26.github.io/goninja/docs/) for
each. The runtime is the root `goninja` package (`BaseResource`, error
types, hooks, auth, config, pagination, validation) plus four focused
subpackages — `openapi` (standalone OpenAPI 3.0 types), `docsui` (pluggable
docs UI, depends on `openapi`), `id` (UUID helper), and `goninjatest`
(in-memory SQLite + httptest helpers for testing your own resources) — each
with its own test suite (`make cover` for coverage across all of them plus
`internal/codegen`, enforced at 70% in CI).

### Generated docs UI

One call — `app.MountDocs(mux, "/docs", ui)` — serves the merged OpenAPI
document as JSON plus a rendered UI, both fully embedded (no external CDN).
`ui` is an interface, not a hardcoded renderer, so swapping one line swaps
the whole UI — `docsui.SwaggerUI()` (default) or `docsui.ReDoc()`:

<table>
<tr>
<td width="50%" align="center">

#### Swagger UI

<img src="docs/screenshots/swagger-ui.png" alt="Swagger UI listing every route grouped by model, including the custom POST /books/{id}/publish action alongside the generated Book CRUD routes" width="100%">

</td>
<td width="50%" align="center">

#### ReDoc

<img src="docs/screenshots/redoc.png" alt="ReDoc three-pane layout with sidebar nav, showing the custom Publish a book action documented next to the generated Book routes" width="100%">

</td>
</tr>
</table>

Every operation expands into its request/response schema, complete with
example values generated straight from the model's fields — no hand-written
OpenAPI, ever, including custom [actions](https://caspel26.github.io/goninja/docs/extending/actions/):

<p align="center">
<img src="docs/screenshots/swagger-ui-operation.png" alt="Swagger UI showing an expanded POST /books/{id}/publish operation, its id path parameter, and 200/404 responses" width="720">
<br>
<sub>An expanded <code>POST /books/{id}/publish</code> action in Swagger UI</sub>
</p>

---

## Documentation

The [full documentation site](https://caspel26.github.io/goninja/) covers:

- [Getting Started](https://caspel26.github.io/goninja/docs/getting-started/)
- [Comparison](https://caspel26.github.io/goninja/docs/comparison/) vs Huma, gocrud, gorest
- [CLI reference](https://caspel26.github.io/goninja/docs/cli/), including `-watch` mode
- [Extending a resource](https://caspel26.github.io/goninja/docs/extending/): hooks and overrides,
  custom validation tags, custom paths and restricted routes, custom
  actions beyond CRUD, global auth and middleware, relations
  (nested vs. by ID), and testing your own resources

---

## Contributing

The project is pre-alpha; the implementation plan is the source of truth
for scope and sequencing. Open an issue before a large PR.

## License

[MIT](LICENSE)
