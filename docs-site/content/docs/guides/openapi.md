---
title: OpenAPI & Docs UI
weight: 8
---
Every generated `<Model>Resource` implements `openapi.OpenAPIProvider`
via a generated `OpenAPI()` method, built from the same IR the rest of
the generated file is built from — request/response schemas, routes,
and auth requirements. Nothing here is hand-written OpenAPI, and nothing
here can drift from what the handlers actually do, because both are
generated from the same source of truth.

## How the document is assembled

`goninja.NewAPI(title, version)` creates an `API` — the framework's
application entry point. `app.Mount(mux, resources...)` calls
`Register(mux)` and `OpenAPI()` on each resource in turn: `Register`
wires the resource's routes onto the mux, and `OpenAPI()` returns three
things — the resource's paths, its schemas, and its security schemes —
which `API` merges into one document as each resource is added.

```go
mux := http.NewServeMux()
app := goninja.NewAPI("bookstore", "0.1.0")

app.Mount(mux,
    api.NewTaskResource(db),
    api.NewAuthorResource(db),
    api.NewBookResource(db),
)
```

`app.Spec()` returns the merged `openapi.Spec` at any point after
mounting. You won't normally call it directly — `MountDocs` (below)
serves it — but it's there if you need the raw document for something
else (writing it to a file for an external tool, for instance).

## Mounting the docs UI

One call serves the merged document as JSON plus a rendered UI, both
fully embedded:

```go
app.MountDocs(mux, "/docs", docsui.SwaggerUI())
```

`MountDocs` serves the UI at `path + "/"` and issues a 301 redirect from
the bare `path` to it. This isn't cosmetic: a `DocsUI`'s `Index` HTML
references its `Assets` by relative filename, and those only resolve
correctly against the trailing-slash URL — visiting `/docs` directly
(no trailing slash) would otherwise leave the browser resolving
`swagger-ui.css` against the wrong base path.

Assets are vendored under `docsui/swagger-ui/` and `docsui/redoc/` and
served via `go:embed` — there is no external CDN dependency, so the docs
UI works offline and isn't subject to a third party changing or removing
a hosted asset out from under you.

![Swagger UI listing every route grouped by model](/screenshots/swagger-ui.png)

## Swapping the UI

`ui` is the `docsui.DocsUI` interface, not a hardcoded renderer, so
switching renderers is a one-line change:

```go
app.MountDocs(mux, "/docs", docsui.ReDoc())
```

![ReDoc three-pane layout](/screenshots/redoc.png)

Both ship built in with their own vendored assets. Pick whichever fits
your team — Swagger UI's try-it-out console is often preferred for
day-to-day API exploration, ReDoc's three-pane layout for a cleaner
reference document.

## Custom actions are documented too

An `Action` with a non-empty `Summary` (see
[Adding Custom Routes Beyond CRUD](../actions)) gets a path entry in the
generated `OpenAPI()` output alongside the CRUD routes, tagged the same
way. `examples/prototype`'s `POST /books/{id}/publish` shows up in the
UI exactly like the generated routes around it, including its `id` path
parameter and declared `200`/`404` responses:

![Expanded POST /books/{id}/publish operation showing its id path parameter and 200/404 responses](/screenshots/swagger-ui-operation.png)

Leave `Summary` empty on an `Action` to mount it without documenting it
— useful for an internal-only route you don't want showing up in the
public spec.

## Implementing your own DocsUI

`docsui.DocsUI` is a two-method interface:

```go
type DocsUI interface {
    Index(specURL string) []byte
    Assets() (fs.FS, string)
}
```

`Index` returns the HTML page for the UI, given the URL the JSON spec is
served from; `Assets` returns a filesystem of static assets (CSS, JS)
plus the URL prefix they're served under. Implement it to wire in a
renderer other than Swagger UI or ReDoc — `SwaggerUI()`/`ReDoc()` are
convenience constructors, not the only supported option.

## Grouping routes

Every operation in a resource's fragment carries
`BaseResource.OpenAPITags()` as its OpenAPI `tags` — this is what
determines how a rendered UI groups a resource's routes together. It
defaults to the model name; override it before `Register`:

```go
bookAPI := api.NewBookResource(db)
bookAPI.SetOpenAPITags("Catalog")
```

## Excluding a resource from the document

To keep a resource's routes mounted while leaving it out of the merged
spec, call `ExcludeFromDocs()` before passing it to `Mount`:

```go
internalAPI := api.NewAuditResource(db)
internalAPI.ExcludeFromDocs()

app.Mount(mux, api.NewTaskResource(db), internalAPI)
```

`Mount` checks this via `BaseResource.DocsExcluded()`. To skip
documenting a resource entirely and bypass `Mount`'s bookkeeping
altogether, call its `Register(mux)` directly instead of going through
`Mount`.

## Security schemes come from the same objects that enforce auth

`openapi.Operation` carries a `Security` field, and the document's
`Components.SecuritySchemes` are populated from the same
`goninja.Authenticator` values passed to `Config.DefaultAuth.Auth` or a
resource's `ResourceConfig.Auth` overrides — not a separate,
hand-maintained description of your auth scheme. A generated `OpenAPI()`
resolves each route's required `Authenticator`s through the identical
logic `Protect` uses to enforce them at request time, so what the
document says is protected and what actually is protected cannot
diverge. See [Global Auth and Middleware](../auth) for how
`Authenticator`s are configured.

{{< callout type="info" >}}
goninja does not generate example values for request/response bodies.
Swagger UI synthesizes its own sample values from the schema types at
render time — that's the UI's behavior, not something goninja emits into
the document.
{{< /callout >}}
