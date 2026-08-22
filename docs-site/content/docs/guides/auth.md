---
title: Authentication
weight: 10
---
`api.MountWithConfig` is `api.Mount` plus a `Config`: a global default
auth policy and generic middleware (logging, CORS, ...) applied to every
resource passed to it.

Auth is an object, not a route-name string list — a
`goninja.Authenticator` you attach at the point of registration,
mirroring how Django Ninja's `AuthBase` works rather than DRF's
`permission_classes`:

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
`DefaultAuth.Auth` is the list of `Authenticator`s tried, in order,
against a protected request — the first one whose `Authenticate` returns
`ok` wins and its `User` is attached to the context, and the request is
rejected with 401 only once every `Authenticator` has declined.
`Authenticate` itself never writes to the response, which is what makes
trying several `Authenticator`s in sequence safe. `Config.Middleware`
wraps every route on every resource unconditionally, public or not — for
logging/CORS-style concerns that aren't about identity. A resource that
reads the authenticated user retrieves it with
`goninja.UserFromContext(ctx)`, the `User` interface being just `ID()
string`; goninja never constructs one itself.

Because an `Authenticator` self-describes its own `SecurityScheme()`, the
same object that enforces auth at runtime is also the sole source of
truth for how it's documented — every protected route's generated
`OpenAPI()` carries a matching `Security` requirement, so enforcement and
documentation can never drift apart.

Override the default per resource or per route with
`ResourceConfig.Auth map[goninja.Route]goninja.RouteAuth` — see
[Custom Path and Restricted Routes](../routing). Plain `api.Mount` still
works exactly as before — a resource it mounts gets a zero `Config`, so
nothing is protected and no middleware runs, unless you switch that
resource to `api.MountWithConfig`.

## Auth on a custom Action

`ResourceConfig.Auth` targets a route by name (`Route(action.Name)`) — a
new [Action](../actions) is easy to add and forget to also list there,
leaving it silently public. `Action.Auth` decides an action's auth right
where the action is declared instead:

```go
protected := &goninja.RouteAuth{Auth: []goninja.Authenticator{BearerAuth{}}}

goninja.Action{
    Name: "publish", Method: http.MethodPost, UrlPath: "publish",
    Handler: publishBookHandler(r),
    Auth:    protected, // same RouteAuth shape as ResourceConfig.Auth's entries
}
```

`Public`/`Auth` on the `RouteAuth` work exactly as they do for
`ResourceConfig.Auth`. Leave `Auth` `nil` to fall back to the name-based
lookup instead — useful when several actions already share the app-wide
default and repeating it on each isn't worth avoiding the lookup.

## Catching a route nobody classified: Config.StrictAuth

Both mechanisms above are opt-in: a route (CRUD or Action) that's never
mentioned in `DefaultAuth.Routes`, `ResourceConfig.Auth`, or `Action.Auth`
is silently public by default — nobody decided that, it just fell through.
`Config.StrictAuth: true` turns that into a startup panic instead, naming
every route with no explicit decision:

```go
cfg := goninja.Config{
    StrictAuth: true,
    DefaultAuth: goninja.AuthPolicy{ /* ... */ },
}
app.MountWithConfig(mux, cfg, resources...) // panics if anything's unclassified
```

A route deliberately left public still needs one explicit line —
`RouteAuth{Public: true}` (in `ResourceConfig.Auth` or `Action.Auth`) or
naming it in `DefaultAuth.Routes` — `StrictAuth` doesn't require
protecting everything, only *deciding* about everything. `false` (the
default) changes nothing: existing apps see no behavior difference unless
they opt in.

## Built-in Authenticators

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
Anything more unusual — a custom header, combining multiple credentials,
a scheme not listed here — still just implements `Authenticator`
directly, as in the `BearerAuth` example above.

A runnable proof of this end to end lives in `examples/prototype`:
`auth.go`'s `newAPIKeyAuth` builds a `goninja.APIKeyHeader` and `main.go`
wires it in — set `PROTOTYPE_API_KEY` before starting the server and
create/update/delete on every resource require an `X-API-Key` header
matching it; leave it unset and the prototype stays fully public, for
frictionless local exploration.
