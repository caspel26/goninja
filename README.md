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
fragment for the model — plain Go under `internal/api`, readable and
debuggable, nothing reflected at runtime. Wiring it into a server is a few
lines, no framework of its own to learn:

```go
// main.go
func main() {
    db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    db.AutoMigrate(&models.Author{}, &models.Book{}) // goninja doesn't generate migrations

    mux := http.NewServeMux()
    doc := goninja.NewAPI("Bookstore API", "0.1.0")

    goninja.Mount(mux, doc,
        api.NewAuthorResource(db),
        api.NewBookResource(db),
    )
    goninja.MountDocs(mux, doc, "/docs", goninja.SwaggerUI())

    http.ListenAndServe(":8080", mux)
}
```

That's a full `net/http` server: `GET/POST /books`, `GET/PUT/DELETE
/books/{id}`, filtering, pagination, validation, and `/docs` — all from the
struct at the top. `goninja.Mount` just does `Register(mux)` +
`doc.Add(...)` for every resource passed to it, so you're never required to
go through it either.

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

Hooks and per-method overriding (below) are built; auth is still designed
but not yet built.

### Example: create → retrieve → filter

Real output from a running `examples/prototype` server — nothing edited:

```console
$ curl -s -X POST localhost:8080/authors \
    -d '{"name":"Ursula K. Le Guin","bio":"Author of the Earthsea series."}'
{"id":"76f431cf-5418-448a-a0bb-fd5b7853b27d","name":"Ursula K. Le Guin","bio":"Author of the Earthsea series."}

$ curl -s -X POST localhost:8080/books \
    -d '{"title":"The Left Hand of Darkness","author_id":"76f431cf-5418-448a-a0bb-fd5b7853b27d","price":14.99,"published":true}'
{"id":"0aac15c6-df8d-4696-af75-f625794b7f3a","title":"The Left Hand of Darkness",
 "author_id":"76f431cf-5418-448a-a0bb-fd5b7853b27d",
 "author":{"id":"76f431cf-5418-448a-a0bb-fd5b7853b27d","name":"Ursula K. Le Guin","bio":"Author of the Earthsea series."},
 "price":14.99,"published":true,"created_at":"2026-08-19T15:19:34+02:00"}
```

`Create`/`Retrieve` preload `Author` automatically and nest it as its own
`Retrieve` shape — no separate round trip. `List` never preloads, by
design, so it stays lean:

```console
$ curl -s 'localhost:8080/books?published=true&price_min=10&order=-created_at&limit=20'
{"items":[
  {"id":"0aac15c6-df8d-4696-af75-f625794b7f3a","title":"The Left Hand of Darkness","author_id":"76f431cf-5418-448a-a0bb-fd5b7853b27d","price":14.99,"published":true,"created_at":"2026-08-19T15:19:34+02:00"},
  {"id":"46a4b864-0d95-4e60-8db0-ce94beff83b0","title":"Book 2","author_id":"a9132a58-0531-4d98-9906-1776da5036f4","price":10,"published":true,"created_at":"2026-08-19T12:34:12+02:00"}
],"total":2,"limit":20,"offset":0}
```

And a `validate`-tag failure never reaches the database:

```console
$ curl -s -X POST localhost:8080/books -d '{"author_id":"76f431cf-...","price":5}'   # no title
{"code":"VALIDATION_FAILED","errors":{"title":"required"}}
```

### Generated docs UI

One call — `goninja.MountDocs(mux, doc, "/docs", ui)` — serves the merged
OpenAPI document as JSON plus a rendered UI, both fully embedded (no
external CDN). `ui` is an interface, not a hardcoded renderer, so swapping
one line swaps the whole UI:

<table>
<tr>
<td width="50%" align="center">

**Swagger UI**<br><sub>`goninja.SwaggerUI()` — the default</sub>

<img src="docs/screenshots/swagger-ui.png" alt="Swagger UI listing every route, grouped by model" width="100%">

</td>
<td width="50%" align="center">

**ReDoc**<br><sub>`goninja.ReDoc()` — a drop-in swap</sub>

<img src="docs/screenshots/redoc.png" alt="ReDoc three-pane layout with sidebar nav and response samples" width="100%">

</td>
</tr>
</table>

Every operation expands into its request/response schema, complete with
example values generated straight from the model's fields — no hand-written
OpenAPI, ever:

<p align="center">
<img src="docs/screenshots/swagger-ui-operation.png" alt="Swagger UI showing an expanded POST /books operation with its request body schema and response example" width="720">
<br>
<sub>An expanded <code>POST /books</code> operation in Swagger UI</sub>
</p>

### Extending a resource: hooks and overrides

Nothing generated is meant to be edited by hand — you extend a resource by
embedding it in your own type. Go has no dynamic dispatch through
embedding, so a wrapper's overrides only take effect once you point the
embedded resource's `Self()` at it:

```go
// resources.go — your file, never touched by the generator.
type authorWithAudit struct {
    *api.AuthorResource
}

// BeforeCreateHook: runs inside the same transaction as Create, and an
// error here rolls the whole request back — nothing gets written.
func (r *authorWithAudit) BeforeCreate(ctx context.Context, in *api.AuthorCreate) error {
    if in.Name == "" {
        return goninja.ValidationError{Fields: map[string]string{"name": "required"}}
    }
    return nil
}

// AfterCreateHook: runs once the row exists, still inside that transaction.
func (r *authorWithAudit) AfterCreate(ctx context.Context, out *api.AuthorRetrieve) error {
    log.Printf("author created: %s", out.ID)
    return nil
}

func NewAuthorWithAudit(db *gorm.DB) *authorWithAudit {
    inner := api.NewAuthorResource(db)
    w := &authorWithAudit{AuthorResource: inner}
    inner.SetSelf(w) // wires the wrapper's hooks/overrides into the generated handlers
    return w
}
```

`BeforeCreateHook`/`AfterCreateHook`/`BeforeUpdateHook`/`BeforeDeleteHook`
are plain optional interfaces (`goninja` root package) — implement only the
ones you need. The same `SetSelf` wiring also makes an overridden
`Retrieve`/`List`/`Update`/`Delete` method take effect, e.g. to add
caching in front of the generated query without forking it.

## Contributing

The project is pre-alpha; the implementation plan is the source of truth
for scope and sequencing. Open an issue before a large PR.

## License

[MIT](LICENSE)
