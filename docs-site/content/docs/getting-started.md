---
title: Getting Started
weight: 1
---

## Install the CLI

```console
$ go install github.com/caspel26/goninja/cmd/goninja@latest
```

## Define a model

Annotate the fields you want exposed with a `goninja` struct tag —
`list`, `retrieve`, `create`, `update`, `filter` — plus an optional
`validate` tag for input validation:

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

## Generate

```console
$ goninja generate \
    -models ./models \
    -out ./internal/api \
    -package api \
    -models-import github.com/you/yourapp/models
```

Pass `-watch` to keep it running and regenerate automatically whenever a
`.go` file under `-models` changes, debounced so a single save triggers
one regeneration:

```console
$ goninja generate -models ./models -out ./internal/api \
    -package api -models-import github.com/you/yourapp/models -watch
```

This writes typed schemas, handlers, database queries, and an OpenAPI
fragment for the model — plain Go under `internal/api`, readable and
debuggable, nothing reflected at runtime.

## Wire it into a server

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
/books/{id}`, filtering, pagination, validation, and `/docs` — all from
the struct at the top. `goninja.NewAPI` is the app's entry point;
`api.Mount` just does `Register(mux)` plus merges each resource's OpenAPI
fragment for every resource passed to it, and `api.MountDocs` serves a
rendered UI over the result — both are thin wrappers over the standalone
`openapi`/`docsui` packages, so you're never required to go through them
either.

Next: [Extending a Resource](../extending/hooks-and-overrides) to add
your own logic without touching generated files.
