---
title: Router Adapters
weight: 12
---
`goninja.Resource.Register` mounts on anything satisfying `goninja.Router`
— a one-method interface `*http.ServeMux` already satisfies as-is:

```go
type Router interface {
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
}
```

Plain `net/http` needs nothing extra — `mux := http.NewServeMux()` +
`app.Mount(mux, resources...)` works exactly as before. For gin, echo, or
chi, wrap the router once with the matching adapter module and pass that
wrapped value everywhere a `Router` is expected — `Mount`, `MountWithConfig`,
`MountDocs`, and a resource's own `Register` all accept it identically.

## gin

```go
go get github.com/caspel26/goninja/adapters/gin
```

```go
engine := gin.New()

r := ginadapter.New(engine) // or ginadapter.New(engine.Group("/api/v1"))
app := goninja.NewAPI("Bookstore API", "0.1.0")
app.Mount(r, taskResource, authorResource, bookResource)
app.MountDocs(r, "/docs", docsui.SwaggerUI())

log.Fatal(engine.Run(":8080"))
```

## echo

```go
go get github.com/caspel26/goninja/adapters/echo
```

```go
e := echo.New()

r := echoadapter.New(e) // or echoadapter.New(e.Group("/api/v1"))
app := goninja.NewAPI("Bookstore API", "0.1.0")
app.Mount(r, taskResource, authorResource, bookResource)
app.MountDocs(r, "/docs", docsui.SwaggerUI())

log.Fatal(e.Start(":8080"))
```

## chi

```go
go get github.com/caspel26/goninja/adapters/chi
```

```go
mux := chi.NewRouter()

r := chiadapter.New(mux) // or a sub-router mounted via mux.Route(...)
app := goninja.NewAPI("Bookstore API", "0.1.0")
app.Mount(r, taskResource, authorResource, bookResource)
app.MountDocs(r, "/docs", docsui.SwaggerUI())

log.Fatal(http.ListenAndServe(":8080", mux))
```

## How it works

Every adapter translates a generated route's stdlib-style pattern (e.g.
`"GET /books/{id}"`) into that router's own path syntax (`:id` for gin and
echo; chi already matches stdlib's `{id}`), then binds the matched value
back onto the request via `(*http.Request).SetPathValue` before calling the
generated handler. The handler itself — `req.PathValue("id")` and all —
never changes; only route *registration* is router-specific. This also
means `BaseResource.Protect` (auth and middleware) behaves identically
under every router, since it only ever wraps `http.Handler`.

{{< callout type="info" >}}
Each adapter module (`adapters/gin`, `adapters/echo`, `adapters/chi`) is a
separate Go module with its own `go.mod` — so pulling in gin, echo, or chi
never becomes a dependency of a plain `net/http` project. Only `go get`
the one you actually use.
{{< /callout >}}

## Escape hatch

If a router-specific routing limitation gets in the way (gin's tree, for
example, is historically stricter about a param route sitting next to a
static sibling), you can always fall back to mounting a real
`*http.ServeMux` under your router instead — e.g. for gin,
`engine.NoRoute(gin.WrapH(mux))` — and drive that mux with plain
`app.Mount(mux, resources...)`. You lose that router's own route listing
for goninja's paths, but get byte-perfect stdlib routing semantics.

Related: [Paths & Route Config](../routing), [Authentication](../auth).
