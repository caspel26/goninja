---
title: Transactions
weight: 5
---
goninja wraps the write path of a generated resource in a database transaction. This page covers which operations are transactional, why that matters for how you write hooks and custom actions, and how to use the same mechanism yourself.

## Which operations are transactional

| Handler | Transactional |
|---|---|
| `list` | No |
| `retrieve` | No |
| `create` | Yes |
| `update` | Yes |
| `delete` | Yes |

The generated `create`, `update`, and `delete` handlers each wrap their work in `goninja.InTransaction`. `list` and `retrieve` are read-only and run directly against the resource's `*gorm.DB` with no transaction.

## Why `r.DB(ctx)` matters

`BaseResource` exposes:

```go
func (r *BaseResource) DB(ctx context.Context) *gorm.DB
```

`DB(ctx)` checks the context for a transaction first, and only falls back to the resource's own `*gorm.DB` if none is present:

```go
func WithTx(ctx context.Context, tx *gorm.DB) context.Context
func TxFromContext(ctx context.Context) (*gorm.DB, bool)
```

`InTransaction` puts the active transaction on the context via `WithTx` before invoking your function. That means any code — generated or your own — that calls `r.DB(ctx)` instead of holding onto a `*gorm.DB` captured earlier automatically participates in whatever transaction is already running, without needing to know one exists.

{{< callout type="warning" >}}
If a hook or custom action captures `db` from a closure instead of calling `r.DB(ctx)`, it silently runs outside the transaction — its writes commit immediately and won't roll back if a later step in the same request fails.
{{< /callout >}}

## Using `InTransaction` in a custom action

```go
func InTransaction[T any](ctx context.Context, db *gorm.DB, fn func(ctx context.Context) (T, error)) (T, error)
```

A custom action (see [Custom Actions](../actions)) that needs to make multiple related writes atomic can call this directly:

```go {filename="handlers/book.go"}
func publishBookHandler(r *api.BookResource) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		id := req.PathValue("id")
		ctx := req.Context()

		_, err := goninja.InTransaction(ctx, r.DB(ctx), func(ctx context.Context) (struct{}, error) {
			var book models.Book
			if err := r.DB(ctx).First(&book, "id = ?", id).Error; err != nil {
				return struct{}{}, goninja.NotFound{Resource: "book", ID: id}
			}
			book.Published = true
			if err := r.DB(ctx).Save(&book).Error; err != nil {
				return struct{}{}, err
			}
			return struct{}{}, notifySubscribers(ctx, book.ID)
		})
		if err != nil {
			goninja.Respond(w, r.ErrorMapper(), err)
			return
		}
		goninja.RespondJSON(w, http.StatusOK, map[string]string{"status": "published"})
	}
}
```

Note that the inner function calls `r.DB(ctx)` again rather than reusing the outer `r.DB(ctx)` value — both resolve to the same transaction, but calling it fresh at each site is what keeps the code correct if you refactor a piece of it out into a function that receives only `ctx`.

## Hooks run inside the transaction

The hook interfaces are exactly `BeforeCreateHook`, `AfterCreateHook`, `BeforeUpdateHook`, and `BeforeDeleteHook`. There is no `AfterUpdateHook` and no `AfterDeleteHook`.

Every hook that exists runs inside the same transaction as the operation it's attached to. Returning an error from a hook rolls back the entire transaction — not just the hook's own work.

This has a consequence worth calling out explicitly for `AfterCreateHook`: because it runs inside the same transaction as the insert, an error returned from `AfterCreate` rolls back the row that was just written. The row never existed as far as any other transaction is concerned, even though `Create` had already executed successfully at the point the hook ran.

```go {filename="handlers/book.go"}
func (h *bookHooks) AfterCreate(ctx context.Context, book *models.Book) error {
	if err := indexForSearch(ctx, book); err != nil {
		// this rolls back the Create that just happened
		return err
	}
	return nil
}
```

If a failure in `AfterCreate` shouldn't undo the insert — for example, a best-effort side effect like search indexing — don't return the error from the hook. Log it and return `nil`, or move the work out of the hook entirely and trigger it after the request completes.

Related: [Hooks & Overrides](../hooks-and-overrides), [Errors & Responses](../errors), [Custom Actions](../actions).
