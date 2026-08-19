# goninja

Code-first Go framework for generating complete CRUD REST APIs from
annotated structs: routing, input/output validation, serialization,
OpenAPI, filters, pagination.

Define your model once, annotate its fields, and generate the rest.

```go
// models/book.go
type Book struct {
    ID        int64     `gorm:"primaryKey" goninja:"list,retrieve"`
    Title     string    `gorm:"size:120;not null" goninja:"list,retrieve,create,update" validate:"required,max=120"`
    AuthorID  int64     `goninja:"list,retrieve,create,update,filter"`
    Price     float64   `goninja:"list,retrieve,create,update,filter" validate:"min=0"`
    Published bool      `goninja:"list,retrieve,create,update,filter"`
}
```

```console
$ goninja generate
```

This generates typed schemas, handlers, database queries, and an OpenAPI
fragment for the model, mounted onto a standard `net/http` mux — no
runtime reflection, no hidden magic: the generated code is plain Go you
can read, debug, and step through.

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
prototype lives under [`internal/codegen`](internal/codegen) and
[`examples/prototype`](examples/prototype): it parses `goninja`-tagged
struct fields and generates typed schemas, an in-memory store, and
`net/http` handlers for a single example model, end to end.

Try it:

```console
$ make generate-prototype   # writes examples/prototype/internal/generated
$ make run-prototype        # serves GET/POST /tasks, GET /tasks/{id} on :8080
```

Real database support, validation, OpenAPI, filters/pagination, auth, and
hooks are designed but not yet built. Follow progress in
[goninja-implementation-plan.md](goninja-implementation-plan.md).

## Contributing

The project is pre-alpha; the implementation plan is the source of truth
for scope and sequencing. Open an issue before a large PR.

## License

[MIT](LICENSE)
