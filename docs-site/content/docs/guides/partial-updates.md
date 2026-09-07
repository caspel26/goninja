---
title: Partial Updates
weight: 8
---

Every generated resource exposes both a full replacement and a partial update:

| Method | Path | Input | Meaning |
|---|---|---|---|
| `PUT` | `/<models>/{id}` | `<Model>Update` | Replace every `update`-tagged field. |
| `PATCH` | `/<models>/{id}` | `<Model>Patch` | Change only properties present in the JSON body. |

`PATCH` is not a relaxed `PUT`. Its generated input fields are pointers, so
the server can distinguish a missing property from a deliberate zero value.

```go
type BookPatch struct {
	Title     *string  `json:"title" validate:"omitempty,required,max=200"`
	Price     *float64 `json:"price" validate:"omitempty,min=0"`
	Published *bool    `json:"published"`
}
```

For a book whose title is `"A Wizard of Earthsea"` and `published` is `true`:

```http
PATCH /books/7
Content-Type: application/json

{"published": false}
```

sets `published` to `false` and leaves the title unchanged. Likewise,
`{"price": 0}` writes zero rather than being mistaken for an absent value.

The existing `validate:"..."` rule is applied only when that PATCH property
is supplied. Validation, error mapping, transactions, route auth, strict auth,
and generated OpenAPI all apply to PATCH in the same way they do to PUT.

## Hooks and overrides

Implement `BeforePatchHook` on a resource wrapper to validate or adjust the
pointer-shaped input before it is written:

```go
func (r *bookResource) BeforePatch(ctx context.Context, in *api.BookPatch) error {
	if in.Title != nil {
		*in.Title = strings.TrimSpace(*in.Title)
	}
	return nil
}
```

`BeforePatchHook` runs inside the same transaction as `Patch`. To replace the
operation itself, implement `Patch(ctx, id, in)` on the wrapper; its method
set is separate from the existing CRUD override interface, so adding PATCH
does not disrupt existing `List`, `Retrieve`, `Create`, `Update`, or `Delete`
overrides.
