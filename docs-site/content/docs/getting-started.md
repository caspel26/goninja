---
title: Getting Started
weight: 1
---

This walks from an empty module to a running API with live docs. It assumes Go
1.25 or newer and a database GORM can reach — the examples use Postgres.

## Install the CLI

```shell
go install github.com/caspel26/goninja/cmd/goninja@latest
```

Add the runtime to your module:

```shell
go get github.com/caspel26/goninja
```

## Define a model

A model is a normal GORM struct. The `goninja` tag on each field decides which
operations expose it; a field with no tag never appears in any generated type.

```go {filename="models/author.go"}
package models

import "time"

type Author struct {
    ID        string    `gorm:"primaryKey;type:uuid" goninja:"list,retrieve"`
    Name      string    `gorm:"size:120;not null" goninja:"list,retrieve,create,update" validate:"required,max=120"`
    Country   string    `gorm:"size:2" goninja:"list,retrieve,create,update,filter" validate:"omitempty,len=2"`
    CreatedAt time.Time `goninja:"list,retrieve"`
}
```

```go {filename="models/book.go"}
package models

import "time"

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

`Author` on `Book` is a GORM belongs-to relation, inferred from the `AuthorID`
field by GORM's naming convention. Tagging it `retrieve` means the detail view
nests the full author and preloads it; the list view never does. See
[Struct Tags](../reference/tags) for the full vocabulary.

An ID field is required and must be named literally `ID`. A `string` ID is
treated as a UUID primary key and filled in on create; an `int64` ID is parsed
from the path with `strconv.ParseInt`.

## Generate

```shell
goninja generate -models-import myapp/models
```

`-models-import` is the only required flag — it is the import path of your
models package, written into the generated import block. Everything else has a
default: models are read from `./models`, code is written to `./internal/api`,
and the generated package is named `api`.

That writes one file per model:

{{< filetree/container >}}
  {{< filetree/folder name="internal" >}}
    {{< filetree/folder name="api" >}}
      {{< filetree/file name="author_generated.go" >}}
      {{< filetree/file name="book_generated.go" >}}
    {{< /filetree/folder >}}
  {{< /filetree/folder >}}
{{< /filetree/container >}}

Each file carries a `DO NOT EDIT` header. Commit it — it is part of your source
tree, and reviewing its diff is how you see the effect of a tag change.

{{< callout type="info" >}}
While developing, run `goninja generate -watch -models-import myapp/models` and
it regenerates on every save, debounced so one save is one regeneration.
{{< /callout >}}

## Mount it

```go {filename="main.go"}
package main

import (
    "log"
    "net/http"
    "os"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"

    "github.com/caspel26/goninja"
    "github.com/caspel26/goninja/docsui"

    "myapp/internal/api"
    "myapp/models"
)

func main() {
    db, err := gorm.Open(postgres.Open(os.Getenv("DSN")), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }

    // goninja does not generate migrations.
    if err := db.AutoMigrate(&models.Author{}, &models.Book{}); err != nil {
        log.Fatal(err)
    }

    mux := http.NewServeMux()
    app := goninja.NewAPI("Bookstore API", "0.1.0")

    app.Mount(mux,
        api.NewAuthorResource(db),
        api.NewBookResource(db),
    )
    app.MountDocs(mux, "/docs", docsui.SwaggerUI())

    log.Fatal(http.ListenAndServe(":8080", mux))
}
```

`Mount` registers each resource's routes on the mux and merges its OpenAPI
fragment into one document, which `MountDocs` then serves alongside a rendered
UI.

## What you get

| Method | Path | Description |
|---|---|---|
| `GET` | `/books` | list, with filters, ordering and pagination |
| `POST` | `/books` | create, validated against the `validate` tags |
| `GET` | `/books/{id}` | retrieve, with relations preloaded |
| `PUT` | `/books/{id}` | update |
| `DELETE` | `/books/{id}` | delete |
| `GET` | `/docs/` | Swagger UI over the merged OpenAPI document |

Same for `/authors`. Run it and try a query:

```shell
curl "localhost:8080/books?published=true&price_min=10&order=-created_at&limit=20"
```

```json
{
  "items": [
    {
      "id": "0f3d9a3e-6c1b-4a2f-9f77-2b1c8e5d4a10",
      "title": "The Go Programming Language",
      "author_id": "b21e5c74-9f0a-4c33-8f21-6de0a1b7c559",
      "price": 34.99,
      "published": true,
      "created_at": "2026-08-20T09:14:02Z"
    }
  ],
  "total": 128,
  "limit": 20,
  "offset": 0
}
```

List responses are wrapped in that envelope; retrieve, create and update return
the object directly. A validation failure returns 422 with a per-field body:

```json
{
  "code": "VALIDATION_FAILED",
  "errors": { "title": "required" }
}
```

## Next steps

{{< cards >}}
  {{< card link="../how-it-works" title="How It Works" icon="cog" subtitle="What the generator does and what runs per request." >}}
  {{< card link="../reference/tags" title="Struct Tags" icon="tag" subtitle="Every verb and modifier, in one table." >}}
  {{< card link="../guides/querying" title="Filtering & Pagination" icon="filter" subtitle="Filters, ranges, ordering, limits." >}}
  {{< card link="../guides/auth" title="Authentication" icon="lock-closed" subtitle="Protect routes with Authenticator objects." >}}
  {{< card link="../examples/bookstore" title="Full Example" icon="book-open" subtitle="The complete project, end to end." >}}
{{< /cards >}}
