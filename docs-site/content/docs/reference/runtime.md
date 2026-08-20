---
title: Runtime API
weight: 4
---

The symbols generated code depends on, and the ones you call yourself. Import
paths are shown per package.

## Packages

| Package | Import path | Role |
|---|---|---|
| root | `github.com/caspel26/goninja` | Everything mandatory: resources, errors, validation, pagination, auth, config |
| `openapi` | `github.com/caspel26/goninja/openapi` | OpenAPI 3.0 wire types. Standalone — depends on nothing else here |
| `docsui` | `github.com/caspel26/goninja/docsui` | Pluggable docs UI. Depends only on `openapi` |
| `id` | `github.com/caspel26/goninja/id` | `NewUUID()`, used when a model has a `string` ID |
| `goninjatest` | `github.com/caspel26/goninja/goninjatest` | Test-only helpers. Never import from production code |

The dependency graph is acyclic: `openapi` and `id` are standalone, `docsui`
depends on `openapi`, and the root package depends on `openapi` and `docsui`.

## Application

```go
func NewAPI(title, version string) *API

func (a *API) Mount(mux *http.ServeMux, resources ...Resource)
func (a *API) MountWithConfig(mux *http.ServeMux, cfg Config, resources ...Resource)
func (a *API) MountDocs(mux *http.ServeMux, path string, ui docsui.DocsUI)
func (a *API) Add(p openapi.OpenAPIProvider)
func (a *API) Spec() openapi.Spec
```

`Resource` is what every generated `<Model>Resource` satisfies:

```go
type Resource interface {
    Register(mux *http.ServeMux)
    openapi.OpenAPIProvider
}
```

Use `Mount` when there is no global policy to apply and `MountWithConfig` once
there is. `MountWithConfig` calls `SetConfig` on every resource before
registering it; a resource mounted with plain `Mount` has a zero `Config`, so
route protection is a no-op.

## BaseResource

Embedded by every generated resource. These are the methods you call or
override against.

```go
// Database access. Returns the transaction from ctx when inside one.
func (r *BaseResource) DB(ctx context.Context) *gorm.DB

// Wrapper dispatch. Required for hooks, method overrides and Configurer.
func (r *BaseResource) SetSelf(self any)
func (r *BaseResource) Self() any

// Error mapping.
func (r *BaseResource) SetErrorMapper(m ErrorMapper)
func (r *BaseResource) ErrorMapper() ErrorMapper

// Non-CRUD endpoints.
func (r *BaseResource) SetActions(actions ...Action)
func (r *BaseResource) Actions() []Action

// OpenAPI.
func (r *BaseResource) SetOpenAPITags(tags ...string)
func (r *BaseResource) OpenAPITags() []string
func (r *BaseResource) ExcludeFromDocs()
func (r *BaseResource) DocsExcluded() bool

// Applied by MountWithConfig; consulted by generated Register.
func (r *BaseResource) SetConfig(cfg Config)
func (r *BaseResource) Config() Config
func (r *BaseResource) Protect(route Route, cfg ResourceConfig, h http.HandlerFunc) http.HandlerFunc
func (r *BaseResource) SecurityFor(route Route, cfg ResourceConfig) []Authenticator
```

{{< callout type="warning" >}}
Always read the database through `r.DB(ctx)` rather than capturing a `*gorm.DB`.
Inside a transactional operation `DB(ctx)` returns the transaction; a captured
handle would silently write outside it.
{{< /callout >}}

## Transactions

```go
func InTransaction[T any](ctx context.Context, db *gorm.DB, fn func(ctx context.Context) (T, error)) (T, error)
func WithTx(ctx context.Context, tx *gorm.DB) context.Context
func TxFromContext(ctx context.Context) (*gorm.DB, bool)
```

Generated `create`, `update` and `delete` handlers run inside
`InTransaction`. `list` and `retrieve` do not. See
[Transactions](../../guides/transactions).

## Errors

```go
type NotFound struct {
    Resource string
    ID       any
}

type ValidationError struct {
    Fields map[string]string
}

type BadRequest struct {
    Detail string
}

type ErrorMapper interface {
    MapError(err error) (status int, body any)
}

type DefaultErrorMapper struct{}

func Respond(w http.ResponseWriter, mapper ErrorMapper, err error)
func RespondJSON(w http.ResponseWriter, status int, v any)
```

All three error types are struct values, constructed as literals. The default
mapping:

| Error | Status | Body |
|---|---|---|
| `NotFound` | 404 | `{"code":"NOT_FOUND","error":"…"}` |
| `ValidationError` | 422 | `{"code":"VALIDATION_FAILED","errors":{…}}` |
| `BadRequest` | 400 | `{"code":"BAD_REQUEST","error":"…"}` |
| anything else | 500 | `{"code":"INTERNAL","error":"internal error"}` |

Matching uses `errors.As`, so wrapping with `%w` still maps correctly. Full
detail in [Errors & Responses](../../guides/errors).

## Validation

```go
func Validate(v any) error
func RegisterValidation(tag string, fn validator.Func) error
```

`Validate` runs `go-playground/validator` against a value's `validate` tags and
returns a `ValidationError` keyed by JSON field name. Generated `Create` and
`Update` call it before touching the database.

`RegisterValidation` adds a custom tag to the shared validator. Call it once at
startup, before serving traffic. See [Validation](../../guides/validation).

## Pagination

```go
const DefaultLimit = 20
const MaxLimit     = 100

func ParseLimitOffset(q url.Values) (limit, offset int, err error)

type ListEnvelope[T any] struct {
    Items  []T   `json:"items"`
    Total  int64 `json:"total"`
    Limit  int   `json:"limit"`
    Offset int   `json:"offset"`
}
```

A `limit` above `MaxLimit` is clamped silently. A negative or non-numeric
`limit`/`offset` is a `BadRequest`. See
[Filtering, Ordering & Pagination](../../guides/querying).

## Hooks

```go
type BeforeCreateHook[TIn any] interface {
    BeforeCreate(ctx context.Context, in *TIn) error
}
type AfterCreateHook[TOut any] interface {
    AfterCreate(ctx context.Context, out *TOut) error
}
type BeforeUpdateHook[TIn any] interface {
    BeforeUpdate(ctx context.Context, in *TIn) error
}
type BeforeDeleteHook[TID any] interface {
    BeforeDelete(ctx context.Context, id TID) error
}
```

These four are the complete set — there is no `AfterUpdateHook` or
`AfterDeleteHook`. All run inside the operation's transaction, so returning an
error rolls the operation back. Hooks are found through `Self()`, so they only
fire on a wrapper that called `SetSelf`.

## Routes and configuration

```go
type Route string

const (
    RouteList     Route = "list"
    RouteRetrieve Route = "retrieve"
    RouteCreate   Route = "create"
    RouteUpdate   Route = "update"
    RouteDelete   Route = "delete"
)

type Config struct {
    DefaultAuth AuthPolicy
    Middleware  []func(http.Handler) http.Handler
}

type AuthPolicy struct {
    Routes []Route
    Auth   []Authenticator
}

type ResourceConfig struct {
    Path   string
    Routes []Route
    Auth   map[Route]RouteAuth
}

type RouteAuth struct {
    Auth   []Authenticator
    Public bool
}

type Configurer interface {
    ResourceConfig() ResourceConfig
}
```

`Config.Middleware` wraps every route unconditionally. `ResourceConfig.Routes`
is an opt-in restriction: empty means every route is registered. In
`ResourceConfig.Auth`, the presence of a `Route` key means that route is
overridden — `Public: true` removes protection, a non-empty `Auth` swaps in
different authenticators, and an empty `RouteAuth{}` opts the route into the
global default.

See [Paths & Route Config](../../guides/routing) and
[Authentication](../../guides/auth).

## Authentication

```go
type User interface {
    ID() string
}

func WithUser(ctx context.Context, u User) context.Context
func UserFromContext(ctx context.Context) (User, bool)

type Authenticator interface {
    Authenticate(r *http.Request) (User, bool)
    Name() string
    SecurityScheme() openapi.SecurityScheme
}
```

`Authenticate` returns an identity or declines without touching the response,
which is what lets several authenticators be tried in order. Built-in
implementations:

| Type | Credential | Default name |
|---|---|---|
| `HTTPBearer` | `Authorization: Bearer <token>` | `bearer` |
| `HTTPBasic` | RFC 7617 basic auth | `basic` |
| `APIKeyHeader` | a header, default `X-API-Key` | `apiKey` |
| `CookieKey` | a cookie, default `session` | `cookieAuth` |

Each takes a `Verify` closure receiving the already-extracted credential and
returning `(User, bool)`. All four decline when `Verify` is nil or the
credential is absent.

## Actions

```go
type Action struct {
    Name      string
    Detail    bool
    Method    string
    UrlPath   string
    Handler   http.HandlerFunc
    Summary   string
    Responses map[string]openapi.Response
}
```

`Detail: true` mounts under `<base>/{id}/<UrlPath>`, otherwise
`<base>/<UrlPath>`. `Name` is the key `ResourceConfig.Auth` targets, converted
to a `Route` where needed. A non-empty `Summary` is what puts the action in the
OpenAPI document. See [Custom Actions](../../guides/actions).

## Testing helpers

```go
func NewDB(t *testing.T, models ...any) *gorm.DB
func NewServer(t *testing.T, resources ...goninja.Resource) *httptest.Server
```

From `goninjatest`. `NewDB` opens and migrates in-memory SQLite; `NewServer`
mounts resources on an `httptest.Server`, registering routes only — no OpenAPI
merging or auth wiring. Both clean up via `t.Cleanup`. See
[Testing](../../guides/testing).
