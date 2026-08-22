---
title: Custom Actions
weight: 9
---
`Action` + `BaseResource.SetActions` let you declare extra endpoints as
data — a `Name`, whether it's `Detail` (mounted under
`<base>/{id}/<UrlPath>`) or
collection-level (`<base>/<UrlPath>`), an HTTP `Method`, and a `Handler` —
instead of writing route-mounting code by hand. `Register(mux)` mounts
every action declared via `SetActions` automatically, after the generated
CRUD routes, wrapped through `ProtectAction` — the same auth/middleware
chain `Protect` gives the generated CRUD routes, but consulting the
action's own `Auth` field first (see [Authentication](../auth)) before
falling back to `Action.Name` converted to `goninja.Route(a.Name)`, the
same key `ResourceConfig.Auth`/`RouteList`/`RouteCreate`/etc. use; a
`Summary` on the `Action` gets it documented in `OpenAPI()` too, no extra
step. Unlike hooks and method overrides, this needs no wrapper type or
`SetSelf` — an `Action` already carries its own `http.HandlerFunc`, so
there's nothing to dispatch per request; call `SetActions` right after
constructing the resource.

Keep the handler logic in its own file, next to the model it operates
on — a function taking the already-built resource and returning its
`[]Action` — and call `SetActions` explicitly in `main.go` alongside the
rest of your wiring, rather than hiding it behind a custom constructor.
That way `main.go` stays the one place you can see everything that's
actually mounted:

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

// BookActions returns the custom actions to declare on r via
// goninja.Actions (main.go) or SetActions — auth may be nil, see
// "Auth on an Action" below.
func BookActions(r *api.BookResource, auth *goninja.RouteAuth) []goninja.Action {
    return []goninja.Action{
        {
            Name:    "publish",
            Detail:  true,
            Method:  http.MethodPost,
            UrlPath: "publish",
            Handler: publishHandler(r),
            Summary: "Publish a book",
            Responses: map[string]openapi.Response{"200": {Description: "OK"}},
            Auth:    auth,
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

    bookAPI := api.NewBookResource(db, goninja.Actions(handlers.BookActions, actionAuth))

    app.Mount(mux, api.NewTaskResource(db), api.NewAuthorResource(db), bookAPI)
}
```

Every generated `New<Model>Resource` accepts optional `<Model>Option`s
(e.g. `BookOption`), applied in order — a resource with no actions (like
`api.NewTaskResource(db)` here) just calls it with none, unchanged.
`goninja.Actions` builds one of these `Option`s from an action-builder
function, so actions attach right at construction instead of a separate
`SetActions` step afterward — nothing about what's mounted is hidden,
unlike wrapping construction *and* `SetActions` inside a resource-specific
constructor of your own (which would trade that visibility away):

```go
func Actions[R interface{ SetActions(...Action) }, A any](build func(R, A) []Action, arg A) func(R)
```

where `handlers.BookActions` is `func(r *api.BookResource, auth *goninja.RouteAuth) []goninja.Action`
— `arg`'s type isn't fixed to `*RouteAuth`, it's whatever `build`'s second
parameter needs. A build func that genuinely needs nothing beyond `r`
doesn't need `goninja.Actions` at all: `r := api.NewBookResource(db);
r.SetActions(handlers.BookActions(r)...)` is already just as short.
`examples/prototype/main.go` uses `goninja.Actions` for both `Task` and
`Book`, threading the same `*goninja.RouteAuth` (nil when
`PROTOTYPE_API_KEY` is unset) into both.

`goninja.Respond`/`goninja.RespondJSON` inside the handler are the same
helpers the generated handlers use for the error and success paths,
respectively — `Respond` maps an error through an `ErrorMapper` and
writes it as JSON; `RespondJSON` writes any value with a given status.
Leave an `Action`'s `Summary` empty to mount it without documenting it,
which is fine for an internal-only route. A runnable version of this
(flattened into `package main` rather than a separate `handlers`
package) lives in `examples/prototype/bookpublish.go`, wired into
`main.go`.

## Auth on an Action

Set `Auth` right on the `Action` instead of separately re-listing
`Route("publish")` in `Config.DefaultAuth.Routes` or `ResourceConfig.Auth`
somewhere else — one less place for the two to drift apart as actions get
added over time. Extending `BookActions` above with a second action that
needs *different* auth than `publish` (rather than the same `*RouteAuth`
passed straight through):

```go
func BookActions(r *api.BookResource, agentAuth goninja.Authenticator) []goninja.Action {
    protected := &goninja.RouteAuth{Auth: []goninja.Authenticator{agentAuth}}
    return []goninja.Action{
        {
            Name: "publish", Detail: true, Method: http.MethodPost, UrlPath: "publish",
            Handler: publishHandler(r), Summary: "Publish a book",
            Responses: map[string]openapi.Response{"200": {Description: "OK"}},
            Auth: protected, // requires agentAuth, regardless of DefaultAuth.Routes
        },
        {
            Name: "preview", Detail: true, Method: http.MethodGet, UrlPath: "preview",
            Handler: previewHandler(r), Summary: "Preview an unpublished book",
            Responses: map[string]openapi.Response{"200": {Description: "OK"}},
            Auth: &goninja.RouteAuth{Public: true}, // explicitly open, not just unlisted
        },
    }
}
```

`publish` and `preview` share the same `*RouteAuth` pointer where they
share the same policy — build it once, reuse it, rather than repeating the
`Authenticator` list per action. Leave `Auth` `nil` on an action that
should just follow whatever `ResourceConfig.Auth`/`Config.DefaultAuth`
already say for `Route("name")` — the lookup this replaces, still
available as the fallback.

If `Config.StrictAuth` is on (see [Authentication](../auth)), forgetting
`Auth` entirely on a new action — leaving it neither protected nor
explicitly `Public` anywhere — is a startup panic naming that action,
instead of a route nobody actually decided about serving traffic
unprotected.

### More than one variable policy: bundle arg in a struct

`goninja.Actions(build, arg)` threads exactly one `arg` value into `build`
— above, that's enough because only one policy (`agentAuth`) varies from
outside; `preview`'s `Public` is fixed in the function body. When *two or
more* actions on the same resource each need their own externally-supplied
policy (an agent-only action alongside an admin-only one, say), `arg`
becomes a small struct bundling them instead of a single `*RouteAuth`:

```go
type bookActionAuth struct {
    Publish, Archive *goninja.RouteAuth
}

func BookActions(r *api.BookResource, auth bookActionAuth) []goninja.Action {
    return []goninja.Action{
        {Name: "publish", ..., Auth: auth.Publish},
        {Name: "archive", ..., Auth: auth.Archive}, // a different policy than publish
    }
}
```

```go
bookAPI := api.NewBookResource(db, goninja.Actions(BookActions, bookActionAuth{
    Publish: agentAuth,
    Archive: adminAuth,
}))
```

`arg` is still exactly one value — a struct is just what that value looks
like once more than one independently-varying policy needs to reach
`build`.
