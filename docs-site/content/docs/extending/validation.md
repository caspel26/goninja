---
title: Custom Validation Tags
weight: 2
---

`Create`/`Update` validation runs against `go-playground/validator`'s
built-in tags (`required`, `max`, `email`, ...) out of the box. Register
your own with `goninja.RegisterValidation` — a thin forward to the
library's own mechanism, called once at startup before serving traffic:

```go
goninja.RegisterValidation("isbn", func(fl validator.FieldLevel) bool {
    return isValidISBN(fl.Field().String()) // yours
})
```

```go
type Book struct {
    ISBN string `goninja:"create,update" validate:"required,isbn"`
}
```

Every generated `Create`/`Update`'s `Validate` call picks it up
immediately — there's nothing to wire per resource.
