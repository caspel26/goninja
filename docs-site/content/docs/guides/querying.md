---
title: Filtering, Ordering & Pagination
weight: 1
---
Every generated `List` handler accepts filters, ordering, and pagination through query parameters. This page documents exactly what those parameters are and how the generated code interprets them, using a `Book` model as the running example.

## Filters

A model field tagged `filter` gets an exact-match pointer field on the generated `<Model>Filters` struct. If the field is numeric, it also gets `Min` and `Max` pointer fields for range queries.

Given a `Book` model with `AuthorID`, `Price`, and `Published` fields tagged `filter`, the generator produces:

```go {filename="internal/api/book_generated.go"}
type BookFilters struct {
	AuthorID  *string
	Price     *float64
	PriceMin  *float64
	PriceMax  *float64
	Published *bool
	Limit     int
	Offset    int
	Order     string
}
```

`parseBookFilters(req *http.Request)` reads these query parameters:

| Query parameter | Filter field | Type |
|---|---|---|
| `author_id` | `AuthorID` | string, exact match |
| `price` | `Price` | float64, exact match |
| `price_min` | `PriceMin` | float64, range |
| `price_max` | `PriceMax` | float64, range |
| `published` | `Published` | bool, exact match |
| `limit` | `Limit` | int, pagination |
| `offset` | `Offset` | int, pagination |
| `order` | `Order` | string, ordering |

Range filters use `_min`/`_max` suffixes, not `_gte`/`_lte`. Only the filters actually present in the query become `WHERE` clauses — an absent parameter is not treated as a zero value.

Booleans are parsed with `strconv.ParseBool`. An invalid value produces a `goninja.BadRequest` with one of these exact details:

| Query parameter | Bad-value detail |
|---|---|
| `price` | `invalid price` |
| `price_min` | `invalid price_min` |
| `price_max` | `invalid price_max` |
| `published` | `invalid published` |

See [Errors & Responses](../errors) for how `BadRequest` maps to an HTTP response.

## Ordering

Pass `order=<field>` to sort ascending, or `order=-<field>` to sort descending:

```
GET /books?order=-created_at
```

Only fields tagged `list` are orderable. The generator builds a package-level whitelist map (e.g. `bookOrderableColumns`) from JSON field name to database column, and `List` looks up the requested field there before ever touching `.Order()`:

```go
field, _ := strings.CutPrefix(order, "-")
if _, ok := bookOrderableColumns[field]; !ok {
    return f, goninja.BadRequest{Detail: "cannot order by \"" + field + "\""}
}
```

An unknown or misspelled `order` value is a **400**, raised while parsing the
query string rather than ignored in the query:

```json
{ "code": "BAD_REQUEST", "error": "cannot order by \"titel\"" }
```

The whitelist is also what makes ordering safe against SQL injection: nothing
outside it ever reaches the query builder, regardless of what the client sends.

{{< callout type="info" >}}
If a field isn't sorting the way you expect, check that it's tagged `list` in
the model. Fields without that tag never enter the whitelist, so `order`
requests referencing them are rejected.
{{< /callout >}}

## Pagination

Pagination is handled by `goninja.ParseLimitOffset`, shared across every model:

```go
const DefaultLimit = 20
const MaxLimit     = 100

func ParseLimitOffset(q url.Values) (limit, offset int, err error)
```

| Query parameter | Default | Parse errors | Out-of-range behavior |
|---|---|---|---|
| `limit` | 20 | `BadRequest{Detail: "invalid limit"}` on parse failure or negative value | values above 100 are silently clamped to 100 |
| `offset` | 0 | `BadRequest{Detail: "invalid offset"}` on parse failure or negative value | no upper bound |

Note the asymmetry: an over-limit `limit` is clamped without error, but a malformed or negative value for either parameter is a 400. There's no separate error for "limit too high" — it just gets capped.

`List` counts the total number of matching rows before applying `Limit`/`Offset`, so `total` in the response reflects the full filtered set, not just the returned page.

## Response envelope

List responses are wrapped in `goninja.ListEnvelope[T]`:

```go
type ListEnvelope[T any] struct {
	Items  []T   `json:"items"`
	Total  int64 `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}
```

These four fields are the entire envelope — there's no `has_more`, `page`, or `pages` field. Compute pagination state on the client from `total`, `limit`, and `offset` if you need it.

Example response body:

```json
{
  "items": [
    { "id": "b1e...", "title": "Domain-Driven Design", "price": 39.99, "published": true },
    { "id": "c2f...", "title": "Refactoring", "price": 44.50, "published": true }
  ],
  "total": 37,
  "limit": 20,
  "offset": 0
}
```

## Worked example

```shell
curl "https://api.example.com/books?published=true&price_min=20&price_max=50&order=-price&limit=10&offset=0"
```

This returns published books priced between 20 and 50, sorted by price descending, ten at a time. `total` in the response reflects the full count of published books in that price range, not just the ten returned.

Related: [Struct Tags Reference](../../reference/tags), [Errors & Responses](../errors).
