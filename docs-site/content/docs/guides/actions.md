---
title: Custom Actions
weight: 9
---
`Action` + `BaseResource.SetActions` is goninja's equivalent of
django-ninja-aio-crud's `@action`: declare extra endpoints as data — a
`Name`, whether it's `Detail` (mounted under `<base>/{id}/<UrlPath>`) or
collection-level (`<base>/<UrlPath>`), an HTTP `Method`, and a `Handler` —
instead of writing route-mounting code by hand. `Register(mux)` mounts
every action declared via `SetActions` automatically, after the generated
CRUD routes, wrapped through the same `Protect` auth/middleware chain
those get (`Action.Name`, converted to `goninja.Route(a.Name)`, is what
`ResourceConfig.Auth` keys against, same as
`RouteList`/`RouteCreate`/etc.); a `Summary` on the `Action` gets it
documented in `OpenAPI()` too, no extra step. Unlike hooks and method
overrides, this needs no wrapper type or `SetSelf` — an `Action` already
carries its own `http.HandlerFunc`, so there's nothing to dispatch per
request; call `SetActions` right after constructing the resource.

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
respectively — `Respond` maps an error through an `ErrorMapper` and
writes it as JSON; `RespondJSON` writes any value with a given status.
Leave an `Action`'s `Summary` empty to mount it without documenting it,
which is fine for an internal-only route. A runnable version of this
(flattened into `package main` rather than a separate `handlers`
package) lives in `examples/prototype/bookpublish.go`, wired into
`main.go`.
