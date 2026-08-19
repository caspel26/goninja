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
struct at the top. `goninja.NewAPI` is the app's entry point; `api.Mount`
just does `Register(mux)` + merges each resource's OpenAPI fragment for
every resource passed to it, and `api.MountDocs` serves a rendered UI over
the result — both are thin wrappers over the standalone `openapi`/`docsui`
packages, so you're never required to go through them either.

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
belongs-to and has-many relations on retrieve, `validate`-tag-driven input
validation with per-field 422 responses, `filter`-tag-driven filtering with
limit/offset pagination and ordering behind a `{items, total, limit,
offset}` envelope)
for any number of models, verified end to end against Postgres. A model's
ID field can be `int64` (DB auto-increment) or `string` (a UUID goninja
generates itself) — `examples/prototype`'s models use UUID IDs. Every
generated resource also emits an OpenAPI 3.0 fragment from the same
annotations, groupable under custom tags per resource; `goninja.NewAPI`
merges every resource mounted onto it and `api.MountDocs` serves the result
as JSON plus a docs UI — Swagger UI or ReDoc ship built in, both fully
embedded with no external CDN, and the `DocsUI` interface it takes isn't
hardcoded to either.

Try it (needs a running Postgres):

```console
$ export PROTOTYPE_DSN="host=localhost user=$(whoami) dbname=goninja_prototype sslmode=disable"
$ make generate-prototype   # writes examples/prototype/internal/api
$ make run-prototype        # serves /tasks, /authors, /books on :8080
$ curl "localhost:8080/books?published=true&price_min=10&order=-created_at&limit=20"
$ open http://localhost:8080/docs   # Swagger UI over the merged OpenAPI doc
```

Hooks, per-method overriding, custom path/route config, a global default
auth policy + middleware, a per-field choice between nesting a relation
and exposing just its ID (all below), and end-to-end resource testing via
`goninjatest` are built. The runtime is the root `goninja` package
(`BaseResource`, error types, hooks, auth, config, pagination, validation)
plus four focused subpackages — `openapi` (standalone OpenAPI 3.0 types),
`docsui` (pluggable docs UI, depends on `openapi`), `id` (UUID helper), and
`goninjatest` (in-memory SQLite + httptest helpers for testing your own
resources) — each with its own test suite (`make cover` for coverage
across all of them plus `internal/codegen`, enforced at 70% in CI).

### Generated docs UI

One call — `app.MountDocs(mux, "/docs", ui)` — serves the merged OpenAPI
document as JSON plus a rendered UI, both fully embedded (no external CDN).
`ui` is an interface, not a hardcoded renderer, so swapping one line swaps
the whole UI:

<table>
<tr>
<td width="50%" align="center">

**Swagger UI**<br><sub>`docsui.SwaggerUI()` — the default</sub>

<img src="docs/screenshots/swagger-ui.png" alt="Swagger UI listing every route grouped by model, including the custom POST /books/{id}/publish action alongside the generated Book CRUD routes" width="100%">

</td>
<td width="50%" align="center">

**ReDoc**<br><sub>`docsui.ReDoc()` — a drop-in swap</sub>

<img src="docs/screenshots/redoc.png" alt="ReDoc three-pane layout with sidebar nav, showing the custom Publish a book action documented next to the generated Book routes" width="100%">

</td>
</tr>
</table>

Every operation expands into its request/response schema, complete with
example values generated straight from the model's fields — no hand-written
OpenAPI, ever. The example below is the custom `POST /books/{id}/publish`
action from [Adding custom routes beyond CRUD](#adding-custom-routes-beyond-crud) —
declared with a `Summary`, it's documented exactly like a generated route:

<p align="center">
<img src="docs/screenshots/swagger-ui-operation.png" alt="Swagger UI showing an expanded POST /books/{id}/publish operation, its id path parameter, and 200/404 responses" width="720">
<br>
<sub>An expanded <code>POST /books/{id}/publish</code> action in Swagger UI</sub>
</p>

### Extending a resource: hooks and overrides

Nothing generated is meant to be edited by hand — you extend a resource by
embedding it in your own type. Go has no dynamic dispatch through
embedding, so a wrapper's overrides only take effect once you point the
embedded resource's `Self()` at it:

```go
// handlers/author.go — your file, never touched by the generator.
package handlers

import (
    "context"
    "log"

    "gorm.io/gorm"

    "github.com/caspel26/goninja"
    "myapp/internal/api"
)

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

### Custom validation tags

`Create`/`Update` validation runs against `go-playground/validator`'s
built-in tags (`required`, `max`, `email`, ...) out of the box. Register
your own with `goninja.RegisterValidation` — a thin forward to the
library's own mechanism, called once at startup before serving traffic:

```go
goninja.RegisterValidation("isbn", func(fl validator.FieldLevel) bool {
    return isValidISBN(fl.Field().String()) // yours
})
```

```go
type Book struct {
    ISBN string `goninja:"create,update" validate:"required,isbn"`
}
```

Every generated `Create`/`Update`'s `Validate` call picks it up immediately
— there's nothing to wire per resource.

### Custom path and restricted routes

The same `SetSelf` wrapper can implement `goninja.Configurer` to override the
resource's mount path or drop routes it shouldn't expose — both
`Register(mux)` and the generated OpenAPI fragment pick this up:

```go
func (r *authorWithAudit) Config() goninja.ResourceConfig {
    return goninja.ResourceConfig{
        Path:   "/v1/authors", // default would be "/authors"
        Routes: []goninja.Route{goninja.RouteList, goninja.RouteRetrieve}, // no create/update/delete
    }
}
```

`Routes` is opt-in restriction, not a list you must spell out in full —
leave it unset (or nil) to keep every route. `ResourceConfig` also carries
a per-route `Auth map[goninja.Route]goninja.RouteAuth` override for the
global default auth below — key a route in the map to override it: give it
its own `Auth []goninja.Authenticator` to require different authenticators
than the default, leave `Auth` unset to require the default authenticators
on a route the default policy doesn't otherwise protect, or set
`Public: true` to punch a hole through the default for that one route.

### Adding custom routes beyond CRUD

`Action` + `BaseResource.SetActions` is goninja's equivalent of
django-ninja-aio-crud's `@action`: declare extra endpoints as data — a
`Name`, whether it's `Detail` (mounted under `<base>/{id}/<UrlPath>`) or
collection-level (`<base>/<UrlPath>`), an HTTP `Method`, and a `Handler` —
instead of writing route-mounting code by hand. `Register(mux)` mounts
every action declared via `SetActions` automatically, after the generated
CRUD routes, wrapped through the same `Protect` auth/middleware chain those
get (`Action.Name`, converted to `goninja.Route(a.Name)`, is what
`ResourceConfig.Auth` keys against, same as `RouteList`/`RouteCreate`/etc.);
a
`Summary` on the `Action` gets it documented in `OpenAPI()` too, no extra
step. Unlike hooks and method overrides, this needs no wrapper type or
`SetSelf` — an `Action` already carries its own `http.HandlerFunc`, so
there's nothing to dispatch per request; call `SetActions` right after
constructing the resource.

Keep the handler logic in its own file, next to the model it operates on —
a function taking the already-built resource and returning its `[]Action`
— and call `SetActions` explicitly in `main.go` alongside the rest of your
wiring, rather than hiding it behind a custom constructor. That way
`main.go` stays the one place you can see everything that's actually
mounted:

```go
// handlers/book.go — your file, never touched by the generator.
package handlers

import (
    "net/http"

    "github.com/caspel26/goninja"
    "github.com/caspel26/goninja/openapi"
    "myapp/internal/api"
    "myapp/models"
)

// BookActions returns the custom actions to declare on r via SetActions.
func BookActions(r *api.BookResource) []goninja.Action {
    return []goninja.Action{
        {
            Name:    "publish",
            Detail:  true,
            Method:  http.MethodPost,
            UrlPath: "publish",
            Handler: publishHandler(r),
            Summary: "Publish a book",
            Responses: map[string]openapi.Response{"200": {Description: "OK"}},
        },
    }
}

func publishHandler(r *api.BookResource) http.HandlerFunc {
    return func(w http.ResponseWriter, req *http.Request) {
        ctx := req.Context()
        id := req.PathValue("id")
        if err := r.DB(ctx).Model(&models.Book{}).Where("id = ?", id).
            Update("published", true).Error; err != nil {
            goninja.Respond(w, r.ErrorMapper(), err)
            return
        }
        out, err := r.Retrieve(ctx, id)
        if err != nil {
            goninja.Respond(w, r.ErrorMapper(), err)
            return
        }
        goninja.RespondJSON(w, http.StatusOK, out)
    }
}
```

```go
// main.go
func main() {
    // ... db setup, mux, app := goninja.NewAPI(...) ...

    bookAPI := api.NewBookResource(db)
    bookAPI.SetActions(handlers.BookActions(bookAPI)...)

    app.Mount(mux, api.NewTaskResource(db), api.NewAuthorResource(db), bookAPI)
}
```

`goninja.Respond`/`goninja.RespondJSON` inside the handler are the same
helpers the generated handlers use for the error and success paths,
respectively — `Respond` maps an error through an `ErrorMapper` and writes
it as JSON; `RespondJSON` writes any value with a given status. Leave an
`Action`'s `Summary` empty to mount it without documenting it, which is
fine for an internal-only route. A runnable version of this (flattened
into `package main` rather than a separate `handlers` package) lives in
`examples/prototype/bookpublish.go`, wired into `main.go`.

### Global auth and middleware

`api.MountWithConfig` is `api.Mount` plus a `Config`: a global default auth
policy and generic middleware (logging, CORS, ...) applied to every
resource passed to it.

Auth is an object, not a route-name string list — a `goninja.Authenticator`
you attach at the point of registration, mirroring how Django Ninja's
`AuthBase` works rather than DRF's `permission_classes`:

```go
type BearerAuth struct{}

func (BearerAuth) Name() string { return "bearer" }

func (BearerAuth) SecurityScheme() openapi.SecurityScheme {
    return openapi.SecurityScheme{Type: "http", Scheme: "bearer"}
}

func (BearerAuth) Authenticate(r *http.Request) (goninja.User, bool) {
    token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
    user, ok := lookupUser(token) // yours
    return user, ok
}

cfg := goninja.Config{
    DefaultAuth: goninja.AuthPolicy{
        Routes: []goninja.Route{goninja.RouteCreate, goninja.RouteUpdate, goninja.RouteDelete},
        Auth:   []goninja.Authenticator{BearerAuth{}},
    },
    Middleware: []func(http.Handler) http.Handler{LoggingMiddleware()},
}

app.MountWithConfig(mux, cfg,
    api.NewAuthorResource(db),
    api.NewBookResource(db),
)
```

`DefaultAuth.Routes` names the routes that require auth by default;
`DefaultAuth.Auth` is the list of `Authenticator`s tried, in order, against
a protected request — the first one whose `Authenticate` returns `ok` wins
and its `User` is attached to the context, and the request is rejected
with 401 only once every `Authenticator` has declined. `Authenticate`
itself never writes to the response, which is what makes trying several
`Authenticator`s in sequence safe. `Config.Middleware` wraps every route on
every resource unconditionally, public or not — for logging/CORS-style
concerns that aren't about identity. A resource that reads the
authenticated user retrieves it with `goninja.UserFromContext(ctx)`, the
`User` interface being just `ID() string`; goninja never constructs one
itself.

Because an `Authenticator` self-describes its own `SecurityScheme()`, the
same object that enforces auth at runtime is also the sole source of truth
for how it's documented — every protected route's generated `OpenAPI()`
carries a matching `Security` requirement, so enforcement and
documentation can never drift apart.

Override the default per resource or per route with
`ResourceConfig.Auth map[goninja.Route]goninja.RouteAuth` — see "Custom
path and restricted routes" above. Plain `api.Mount` still works exactly
as before — a resource it mounts gets a zero `Config`, so nothing is
protected and no middleware runs, unless you switch that resource to
`api.MountWithConfig`.

#### Built-in Authenticators

Writing `Name()`/`SecurityScheme()`/`Authenticate` by hand for a common
scheme is boilerplate goninja ships ready-made — each takes only the
`Verify` closure that turns a raw credential into a `User`:

```go
goninja.HTTPBearer{Verify: func(token string) (goninja.User, bool) { ... }}
goninja.HTTPBasic{Verify: func(username, password string) (goninja.User, bool) { ... }}
goninja.APIKeyHeader{Verify: func(key string) (goninja.User, bool) { ... }}   // default header: X-API-Key
goninja.CookieKey{Verify: func(value string) (goninja.User, bool) { ... }}    // default cookie: session
```

Each has an optional field to change what it reads (`HeaderName`,
`CookieName`) and an `AuthName` to rename its OpenAPI security scheme.
Anything more unusual — a custom header, combining multiple credentials, a
scheme not listed here — still just implements `Authenticator` directly, as
in the `BearerAuth` example above.

A runnable proof of this end to end lives in `examples/prototype`:
`auth.go`'s `newAPIKeyAuth` builds a `goninja.APIKeyHeader` and `main.go`
wires it in — set `PROTOTYPE_API_KEY` before starting the server and
create/update/delete on every resource require an `X-API-Key` header
matching it; leave it unset and the prototype stays fully public, for
frictionless local exploration.

### Relations: nested or by ID

A relation field is nested as the related model's own `Retrieve` type by
default — the full object, `Preload`ed automatically:

```go
type Book struct {
    ID       string `gorm:"primaryKey;type:uuid" goninja:"list,retrieve"`
    AuthorID string `goninja:"list,retrieve,create,update,filter"`
    Author   Author `goninja:"retrieve"` // nested as {"author": {...full Author Retrieve...}}
}
```

Add `byid` to skip that — the field exposes only the related ID instead,
and its `Preload` never runs:

```go
Author Author `goninja:"retrieve,byid"` // {"author_id": "..."} — no nesting, no preload
```

Useful when a caller only ever needs the reference, not the full related
object, and the extra join/preload would be wasted work.

### Testing your own resources

`goninjatest` (a separate module-level package, like `openapi`/`docsui`/`id`
— only test code needs to import it) gets a resource from zero to a real
HTTP request in a few lines: `NewDB` opens and `AutoMigrate`s an in-memory
SQLite database, `NewServer` mounts any number of resources on a fresh
`httptest.Server`. Both clean up after themselves via `t.Cleanup`.

```go
func TestBookResource_Create(t *testing.T) {
    db := goninjatest.NewDB(t, &models.Book{}, &models.Author{})
    srv := goninjatest.NewServer(t, api.NewBookResource(db))

    resp, err := http.Post(srv.URL+"/books", "application/json",
        strings.NewReader(`{"title":"Dune","author_id":"...","price":9.99}`))
    if err != nil || resp.StatusCode != http.StatusCreated {
        t.Fatalf("POST /books: err=%v status=%v", err, resp.StatusCode)
    }
}
```

No Postgres, no hand-rolled `httptest.NewServer(mux)` boilerplate, no
`gorm.Open` call to remember the driver for — generated resource code has
no Postgres-specific behavior, so the same in-memory SQLite connection
exercises it end to end.

## Contributing

The project is pre-alpha; the implementation plan is the source of truth
for scope and sequencing. Open an issue before a large PR.

## License

[MIT](LICENSE)
