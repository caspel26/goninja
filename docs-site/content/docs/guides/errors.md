---
title: Errors & Responses
weight: 4
---
goninja maps a small set of error types to HTTP responses through the `goninja.ErrorMapper` interface. Generated handlers, your own hooks, and your own custom actions all funnel errors through the same mapping, so the response shape stays consistent across a resource.

## The three error types

All three are plain struct values, constructed as literals — not pointers.

```go
type NotFound struct {
	Resource string
	ID       any
}

type ValidationError struct {
	Fields map[string]string
}

type BadRequest struct {
	Detail string
}
```

Use each for a distinct situation:

- **`NotFound`** — the requested row doesn't exist. `Resource` is the model name (lowercase, e.g. `"book"`), `ID` is whatever identifier was looked up.

  ```go
  return goninja.NotFound{Resource: "book", ID: id}
  ```

- **`ValidationError`** — input failed a semantic check that isn't already covered by a `validate` tag. `Fields` maps JSON field name to a short reason.

  ```go
  return goninja.ValidationError{Fields: map[string]string{"isbn": "already in use"}}
  ```

- **`BadRequest`** — the request itself is malformed independent of any model validation (bad query parameter, unparseable body). `Detail` is a short human-readable string.

  ```go
  return goninja.BadRequest{Detail: "invalid price_min"}
  ```

## Status and body mapping

`DefaultErrorMapper{}` matches errors with `errors.As` and maps them like this:

| Error type | Status | Body |
|---|---|---|
| `NotFound` | 404 | `{"code":"NOT_FOUND","error":"<resource> <id> not found"}` |
| `ValidationError` | 422 | `{"code":"VALIDATION_FAILED","errors":{"<field>":"<tag>"}}` |
| `BadRequest` | 400 | `{"code":"BAD_REQUEST","error":"<detail>"}` |
| anything else | 500 | `{"code":"INTERNAL","error":"internal error"}` |

The 500 case never leaks the underlying error message to the client — whatever the actual error says, the client only ever sees `"internal error"`.

## Wrapping is safe

Because the mapper matches with `errors.As`, wrapping one of these types in a standard Go error chain still maps correctly:

```go
return fmt.Errorf("loading book %s: %w", id, goninja.NotFound{Resource: "book", ID: id})
```

`Respond` still resolves this to a 404 with the same body as the unwrapped error. This means you can add context to an error on its way up the call stack without losing the mapping.

## Returning errors from hooks and actions

A `BeforeCreateHook`, `AfterCreateHook`, `BeforeUpdateHook`, or `BeforeDeleteHook` that returns one of these types gets the same mapping as an error from the generated CRUD logic itself, since both paths end up going through the same `Respond` call in the handler.

```go {filename="handlers/book.go"}
func (h *bookHooks) BeforeCreate(ctx context.Context, in *api.BookCreate) error {
	if isbnTaken(ctx, in.ISBN) {
		return goninja.ValidationError{Fields: map[string]string{"isbn": "already in use"}}
	}
	return nil
}
```

See [Hooks & Overrides](../hooks-and-overrides) and [Transactions](../transactions) for how a hook error interacts with the surrounding transaction — a hook error rolls back whatever the handler already did.

For a custom action's handler (see [Custom Actions](../actions)), call `goninja.Respond` yourself:

```go {filename="handlers/book.go"}
func publishBookHandler(r *api.BookResource) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		id := req.PathValue("id")
		if err := publish(req.Context(), r, id); err != nil {
			goninja.Respond(w, r.ErrorMapper(), err)
			return
		}
		goninja.RespondJSON(w, http.StatusOK, map[string]string{"status": "published"})
	}
}
```

`Respond` falls back to `DefaultErrorMapper{}` if the mapper passed in is `nil`, so this works even on a resource that never called `SetErrorMapper`.

## RespondJSON for success responses

`RespondJSON` isn't limited to error paths — it's the same helper for any custom action response:

```go
func RespondJSON(w http.ResponseWriter, status int, v any)
```

It sets `Content-Type: application/json`, writes the given status code, then JSON-encodes `v`. Use it any time a custom action needs to return something other than the standard CRUD response shapes.

## Custom error mappers

A resource can override its error mapping with `BaseResource.SetErrorMapper(m ErrorMapper)`. `BaseResource.ErrorMapper()` returns `DefaultErrorMapper{}` when nothing has been set, so an unconfigured resource behaves exactly like the default mapping above.

A realistic custom mapper wraps the default one and only special-cases an error type of your own:

```go {filename="handlers/book.go"}
type outOfStockError struct {
	BookID string
}

func (e outOfStockError) Error() string {
	return fmt.Sprintf("book %s is out of stock", e.BookID)
}

type bookErrorMapper struct {
	goninja.DefaultErrorMapper
}

func (m bookErrorMapper) MapError(err error) (int, any) {
	var oos outOfStockError
	if errors.As(err, &oos) {
		return http.StatusConflict, map[string]string{
			"code":  "OUT_OF_STOCK",
			"error": oos.Error(),
		}
	}
	return m.DefaultErrorMapper.MapError(err)
}
```

```go {filename="main.go"}
bookAPI := api.NewBookResource(db)
bookAPI.SetErrorMapper(bookErrorMapper{})
```

Every error that isn't `outOfStockError` falls through to `DefaultErrorMapper`'s own mapping unchanged, so `NotFound`/`ValidationError`/`BadRequest` still behave exactly as documented above on this resource.

Related: [Transactions](../transactions), [Hooks & Overrides](../hooks-and-overrides).
