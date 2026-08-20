---
title: Struct Tags
weight: 1
---

goninja reads a single struct tag key, `goninja`, off each field of a model
struct (`tag.Get("goninja")`). The tag value is a comma-separated list of
verbs and, for relation fields, one modifier.

## Verbs and modifier

| Name | Applies to | Effect |
|---|---|---|
| `list` | any field | Include the field in `<Model>List` (the collection view). Also makes the field orderable and contributes to the whitelist of orderable columns. |
| `retrieve` | any field | Include the field in `<Model>Retrieve` (the detail view). A relation field tagged `retrieve` is preloaded. |
| `create` | any field | Include the field in `<Model>Create` (the POST request body). |
| `update` | any field | Include the field in `<Model>Update` (the PUT request body). |
| `filter` | any field | Generate a filter field on `<Model>Filters`. Numeric fields also get `Min`/`Max` range filters. |
| `byid` | relation fields only | Expose the related model's ID instead of nesting the full related type. See [Relations](../../guides/relations). |

These five verbs and one modifier are the complete set. There are no others.

Values are trimmed, so `goninja:"list, retrieve"` and `goninja:"list,retrieve"`
are equivalent. Matching is exact string equality, which makes it
case-sensitive: `List` is not recognized and is silently ignored, same as an
unrecognized string.

## Untagged fields and structs

A field with no `goninja` tag, or an empty one (`goninja:""`), is skipped
entirely — it appears in none of the generated output types.

A struct with no `goninja`-tagged fields at all is silently skipped: no
model is generated for it, and the generator reports no error. If a model
you expect to see generated is missing, check that at least one field
carries a `goninja` tag.

## Why list and retrieve are separate

`list` and `retrieve` produce two independent output types by design, not
as an implementation shortcut. `<Model>List` never preloads relations,
because a collection endpoint that eagerly loads every relation on every
row is a standing N+1 (or at best a wide join) query risk. `<Model>Retrieve`
is the full detail view: it preloads every relation field tagged
`retrieve`. Keep this split in mind when deciding which verbs to put on a
relation field — tagging it `list` does not cause it to be preloaded or
even included on the list type in nested form; only `retrieve` does that.

{{< callout type="info" >}}
A relation field tagged `retrieve` is preloaded with
`q.Preload("<FieldName>")` — unless it also carries `byid`, in which case
preloading is skipped entirely and only the related ID is read off the
model's own foreign key column.
{{< /callout >}}

## Validation

The separate `validate:"..."` struct tag (from `go-playground/validator`)
is copied verbatim onto the matching `<Model>Create` and `<Model>Update`
fields only:

```go {filename="internal/codegen/templates/model.go.tmpl"}
json:"{{.JSONName}}"{{if .ValidateTag}} validate:"{{.ValidateTag}}"{{end}}
```

`<Model>List` and `<Model>Retrieve` fields never carry a `validate` tag,
because validation only applies to input. See
[Validation](../../guides/validation) for how `goninja.Validate` uses it.

## The ID field

A model must have a field literally named `ID`. The generator derives the
model's ID Go type from that field:

```go
func (m Model) IDGoType() string {
	for _, f := range m.Fields {
		if f.Name == "ID" {
			return f.GoType
		}
	}
	return "int64"
}
```

- `int64` IDs are parsed out of the URL path with `strconv.ParseInt`.
- Any other type is taken from the path as-is (a string) and treated as a
  UUID primary key — `Create` fills it in with `id.NewUUID()` when the
  incoming value is empty.

In practice the two supported ID types are `int64` and `string`.

{{< callout type="warning" >}}
If a model has no field named `ID`, `IDGoType()` silently falls back to
`int64` and the generated code will not compile against that model. If the
ID field's type is neither `int64` nor `string`, the generated code also
will not compile, since the path value handed to `Retrieve`/`Update`/`Delete`
is always a string.
{{< /callout >}}

## Complete example

```go {filename="models/book.go"}
package models

import "time"

type Book struct {
	ID        string    `goninja:"list,retrieve" gorm:"primaryKey"`
	Title     string    `goninja:"list,retrieve,create,update,filter" validate:"required,max=200"`
	AuthorID  string    `goninja:"list,retrieve,create,update,filter" validate:"required,uuid4"`
	Author    Author    `goninja:"retrieve,byid"`
	Price     float64   `goninja:"list,retrieve,create,update,filter" validate:"min=0"`
	Published bool      `goninja:"list,retrieve,create,update"`
	CreatedAt time.Time `goninja:"list,retrieve"`
}
```

Here, `Author` is a belongs-to relation tagged `retrieve,byid`: the
generated `BookRetrieve` type exposes `author_id` (typed after `Author`'s
own ID type) instead of a nested `AuthorRetrieve`, and the query skips
`Preload("Author")` entirely. Drop `byid` to get the fully nested relation
instead — see [Relations](../../guides/relations) for both cases in detail.
