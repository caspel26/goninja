---
layout: hextra-home
title: goninja
description: Generate typed, validated REST APIs from annotated Go structs — no reflection, no runtime magic.
---

{{< hextra/hero-badge link="https://github.com/caspel26/goninja/releases" >}}
  <div class="hx:w-2 hx:h-2 hx:rounded-full hx:bg-primary-400"></div>
  <span>Pre-alpha · built in the open</span>
  {{< icon name="arrow-circle-right" attributes="height=14" >}}
{{< /hextra/hero-badge >}}

<div class="hx:mt-6 hx:mb-4">
{{< hextra/hero-headline >}}
  Typed REST APIs,&nbsp;<br class="hx:sm:block hx:hidden" />generated from your structs
{{< /hextra/hero-headline >}}
</div>

<div class="hx:mb-8">
{{< hextra/hero-subtitle >}}
  Annotate a Go struct, run one command, and get real&nbsp;`.go` files: handlers,
  queries, validation, filters, pagination and OpenAPI. Nothing is reflected at
  request time — if a tag is wrong, the build fails, not production.
{{< /hextra/hero-subtitle >}}
</div>

<div class="gn-hero-actions">
{{< hextra/hero-button text="Get started" link="docs/getting-started" >}}
<span class="gn-button-secondary">{{< hextra/hero-button text="Browse examples" link="docs/examples" >}}</span>
</div>

<figure class="gn-demo">
  <div class="gn-demo-window" role="img" aria-label="A Book model with goninja tags becomes a generated BookList type after running goninja generate.">
    <div class="gn-demo-titlebar">
      <span class="gn-demo-dots" aria-hidden="true"><i></i><i></i><i></i></span>
      <span>bookstore</span>
      <span class="gn-demo-status">generated code is yours</span>
    </div>
    <div class="gn-demo-panels">
      <section class="gn-demo-panel">
        <div class="gn-demo-tab"><span>MODEL</span> models/book.go</div>
        <pre><code><span class="gn-code-keyword">type</span> Book <span class="gn-code-keyword">struct</span> {
    ID    <span class="gn-code-type">string</span>  <span class="gn-code-tag">`goninja:"list,retrieve"`</span>
    Title <span class="gn-code-type">string</span>  <span class="gn-code-tag">`goninja:"list,retrieve,create"`</span>
    Price <span class="gn-code-type">float64</span> <span class="gn-code-tag">`goninja:"list,retrieve,filter"`</span>
}</code></pre>
      </section>
      <div class="gn-demo-generate" aria-hidden="true">
        <span>goninja generate</span>
        <b>→</b>
      </div>
      <section class="gn-demo-panel">
        <div class="gn-demo-tab"><span>GENERATED</span> api/book_generated.go</div>
        <pre><code><span class="gn-code-keyword">type</span> BookList <span class="gn-code-keyword">struct</span> {
    ID    <span class="gn-code-type">string</span>  <span class="gn-code-tag">`json:"id"`</span>
    Title <span class="gn-code-type">string</span>  <span class="gn-code-tag">`json:"title"`</span>
    Price <span class="gn-code-type">float64</span> <span class="gn-code-tag">`json:"price"`</span>
}</code></pre>
      </section>
    </div>
    <div class="gn-demo-output">
      <span><b>GET</b> /books</span>
      <span>typed handlers</span>
      <span>GORM queries</span>
      <span>OpenAPI</span>
    </div>
  </div>
  <figcaption>Tag the struct, run <code>goninja generate</code>, read the Go it wrote.</figcaption>
</figure>

{{< hextra/feature-grid cols="3" >}}

  {{< hextra/feature-card
    title="Code-first"
    icon="code"
    subtitle="The struct is the single source of truth. No schema files, no YAML, no DSL — a `goninja` tag on each field decides which operations expose it."
  >}}

  {{< hextra/feature-card
    title="Generated, not reflected"
    icon="lightning-bolt"
    subtitle="`goninja generate` writes ordinary Go you can read, diff and step through in a debugger. Zero reflection on the request path."
  >}}

  {{< hextra/feature-card
    title="Plain net/http"
    icon="server"
    subtitle="Routes mount on an `*http.ServeMux`. No custom router, no context type of its own, no framework lock-in."
  >}}

  {{< hextra/feature-card
    title="Safe by default"
    icon="shield-check"
    subtitle="Output types are always separate structs from your GORM model, so a field can never leak into a response just because it exists on the model."
  >}}

  {{< hextra/feature-card
    title="No N+1 by construction"
    icon="database"
    subtitle="`list` stays lean and never preloads. `retrieve` is the detail view and preloads every relation it carries. That split is a guarantee, not a default."
  >}}

  {{< hextra/feature-card
    title="OpenAPI included"
    icon="book-open"
    subtitle="Every resource emits an OpenAPI fragment built from the same IR as its handlers, merged into one document and served with Swagger UI or ReDoc."
  >}}

{{< /hextra/feature-grid >}}

<div class="hx:mt-16"></div>

{{< hextra/hero-section >}}
  From struct to running API
{{< /hextra/hero-section >}}

<div class="hx:mb-6 hx:text-base hx:text-gray-600 hx:dark:text-gray-400">
Three steps. The middle one is a command; the other two are files you write.
</div>

{{% steps %}}

### Annotate a model

Each `goninja` verb decides which operations expose that field. `validate` tags
apply to input only, so `Price` can never be written negative.

```go {filename="models/book.go"}
type Book struct {
    ID        string    `gorm:"primaryKey;type:uuid" goninja:"list,retrieve"`
    Title     string    `gorm:"size:200;not null" goninja:"list,retrieve,create,update" validate:"required,max=200"`
    AuthorID  string    `goninja:"list,retrieve,create,update,filter" validate:"required,uuid4"`
    Price     float64   `goninja:"list,retrieve,create,update,filter" validate:"min=0"`
    Published bool      `goninja:"list,retrieve,create,update,filter"`
    CreatedAt time.Time `goninja:"list,retrieve"`
    Author    Author    `goninja:"retrieve"`
}
```

### Generate

This writes one `<model>_generated.go` per model — output types, handlers,
queries and an OpenAPI fragment. Add `-watch` to regenerate on save.

```shell
goninja generate -models-import myapp/models
```

### Mount it

```go {filename="main.go"}
mux := http.NewServeMux()
app := goninja.NewAPI("Bookstore API", "0.1.0")

app.Mount(mux, api.NewAuthorResource(db), api.NewBookResource(db))
app.MountDocs(mux, "/docs", docsui.SwaggerUI())

http.ListenAndServe(":8080", mux)
```

That serves a full CRUD surface with filtering, ordering, pagination, validation
and live API docs:

```shell
curl "localhost:8080/books?published=true&price_min=10&order=-created_at&limit=20"
```

```json
{
  "items": [{ "id": "9f1c…", "title": "The Go Programming Language", "price": 34.99 }],
  "total": 128,
  "limit": 20,
  "offset": 0
}
```

{{% /steps %}}

<div class="hx:mt-14"></div>

{{< hextra/hero-section >}}
  Where to go next
{{< /hextra/hero-section >}}

{{< cards >}}
  {{< card link="docs/getting-started" title="Getting Started" icon="play" subtitle="Install the CLI and go from an empty module to a running API." >}}
  {{< card link="docs/reference/tags" title="Tag Reference" icon="tag" subtitle="Every verb and modifier the `goninja` struct tag accepts." >}}
  {{< card link="docs/guides/querying" title="Filtering & Pagination" icon="filter" subtitle="Exact-match and range filters, ordering, and the list envelope." >}}
  {{< card link="docs/guides/auth" title="Authentication" icon="lock-closed" subtitle="Authenticator objects, per-route policy, and the built-in schemes." >}}
  {{< card link="docs/examples" title="Examples" icon="collection" subtitle="Complete, working projects you can copy from." >}}
  {{< card link="docs/comparison" title="Comparison" icon="scale" subtitle="How goninja differs from Huma, gocrud and gorest." >}}
{{< /cards >}}
