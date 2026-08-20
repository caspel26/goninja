---
title: Validation
weight: 2
---

Input validation is driven by the standard `validate` struct tag from
[go-playground/validator](https://github.com/go-playground/validator). goninja
copies that tag onto the generated request types and calls the validator before
anything touches the database.

## How it is wired

A `validate` tag on a model field is copied verbatim onto the matching
`<Model>Create` and `<Model>Update` field:

```go {filename="models/book.go"}
type Book struct {
    ID       string  `gorm:"primaryKey;type:uuid" goninja:"list,retrieve"`
    Title    string  `goninja:"list,retrieve,create,update" validate:"required,max=200"`
    AuthorID string  `goninja:"list,retrieve,create,update" validate:"required,uuid4"`
    Price    float64 `goninja:"list,retrieve,create,update" validate:"min=0"`
    Stock    int     `goninja:"list,retrieve,create,update"`
}
```

becomes:

```go {filename="internal/api/book_generated.go"}
type BookCreate struct {
    Title    string  `json:"title" validate:"required,max=200"`
    AuthorID string  `json:"author_id" validate:"required,uuid4"`
    Price    float64 `json:"price" validate:"min=0"`
    Stock    int     `json:"stock"`
}
```

`Stock` carries no `validate` tag because the model field had none. The
generated `Create` and `Update` methods call `goninja.Validate(in)` as their
first step, before opening a transaction or issuing a query.

{{< callout type="info" >}}
`validate` tags are only copied onto `Create` and `Update`. The `List` and
`Retrieve` types never carry them, because validation applies to input, not to
what you return.
{{< /callout >}}

## The failure response

`goninja.Validate` converts the validator's errors into a
`goninja.ValidationError` keyed by **JSON field name**, not Go field name. The
value is the name of the tag that failed.

A `POST /books` with an empty title and a negative price returns 422:

```json
{
  "code": "VALIDATION_FAILED",
  "errors": {
    "title": "required",
    "price": "min"
  }
}
```

Every failing field is reported in one response. See
[Errors & Responses](../errors) for the full status mapping.

## Validating by hand

`Validate` is exported, so custom action handlers can use the same rules and
produce the same response shape:

```go
func publishHandler(r *api.BookResource) http.HandlerFunc {
    return func(w http.ResponseWriter, req *http.Request) {
        var in PublishRequest
        if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
            goninja.Respond(w, r.ErrorMapper(), goninja.BadRequest{Detail: "invalid body"})
            return
        }
        if err := goninja.Validate(in); err != nil {
            goninja.Respond(w, r.ErrorMapper(), err)
            return
        }
        // ...
    }
}
```

```go
func Validate(v any) error
```

It returns `nil` on success, a `ValidationError` when tags fail, and passes any
other error through unchanged.

## Custom tags

Built-in tags (`required`, `max`, `email`, `uuid4`, `oneof`, …) work out of the
box. To add your own, register it once at startup — before serving traffic —
with `goninja.RegisterValidation`:

```go
func RegisterValidation(tag string, fn validator.Func) error
```

```go {filename="main.go"}
import "github.com/go-playground/validator/v10"

func main() {
    err := goninja.RegisterValidation("isbn", func(fl validator.FieldLevel) bool {
        return isValidISBN(fl.Field().String())
    })
    if err != nil {
        log.Fatal(err)
    }

    // ... mount resources
}
```

Then use it like any other tag:

```go {filename="models/book.go"}
type Book struct {
    ISBN string `goninja:"list,retrieve,create,update" validate:"required,isbn"`
}
```

It forwards to the shared validator instance, so every generated `Create` and
`Update` recognises the tag immediately. There is nothing to wire per resource.

{{< callout type="warning" >}}
Register custom tags before the server starts handling requests. The validator
instance is shared, and registering while requests are in flight is not safe.
{{< /callout >}}

## Rules validation cannot express

Tag-based validation sees one value at a time, with no database and no request
context. Anything needing a query, another field, or the caller's identity
belongs in a hook, which runs inside the operation's transaction and can abort
it by returning an error:

```go
func (r *bookWithChecks) BeforeCreate(ctx context.Context, in *api.BookCreate) error {
    var n int64
    if err := r.DB(ctx).Model(&models.Book{}).
        Where("title = ? AND author_id = ?", in.Title, in.AuthorID).
        Count(&n).Error; err != nil {
        return err
    }
    if n > 0 {
        return goninja.ValidationError{Fields: map[string]string{
            "title": "already exists for this author",
        }}
    }
    return nil
}
```

Returning a `ValidationError` from a hook produces the same 422 body as a tag
failure. See [Hooks & Overrides](../hooks-and-overrides) and
[Transactions](../transactions).

## Next

{{< cards >}}
  {{< card link="../errors" title="Errors & Responses" icon="exclamation-circle" subtitle="Status codes and custom error mapping." >}}
  {{< card link="../hooks-and-overrides" title="Hooks & Overrides" icon="puzzle" subtitle="Validation that needs the database." >}}
{{< /cards >}}
