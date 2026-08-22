---
title: Best Practices
weight: 13
---
Judgment calls that come up in any real app built on goninja — where to put
custom code, when to reach for which mechanism, what to turn on and when.
Each one links back to the guide that covers the mechanism in full; this
page is about *when* to use it, not how it works.

## Project layout

The layout isn't borrowed from any particular router's conventions — it
falls out of two things goninja already commits to elsewhere:

1. **Generated and hand-written code are never the same file.** Every
   `<Model>Retrieve`/`<Model>Create`/handler goninja emits goes into
   `internal/api/`, always regenerable, never hand-edited — see
   [Generated Code](../../reference/generated-code). Anything you write —
   an `Action`, a hook, an `Authenticator`, an error mapper — has to live
   somewhere else by construction, or the next `goninja generate` would
   silently discard it.
2. **`main.go` is the one place you can see everything that's actually
   mounted** — no resource-specific constructor is allowed to hide a
   `SetActions`/`SetErrorMapper` call inside itself (see [Custom
   Actions](../actions) and [Errors & Responses](../errors) for the two
   places this rule already got made explicit). That only holds if
   `main.go` stays wiring-only — the moment handler logic creeps in
   there too, you lose the ability to read it top to bottom and know
   what's mounted.

Put those two together and a `handlers/` package — one file per model,
holding everything that model needs beyond generated CRUD — is the
natural consequence, not a style import from `gin`/`echo`. It happens to
look like how a hand-rolled `gin` app organizes a `handlers/`/
`controllers/` package, which is a coincidence of both landing on
"group hand-written code by the thing it operates on," not goninja
adopting gin's conventions:

```text
myapp/
├── go.mod
├── main.go              # wiring only — construct resources, mount, run
├── models/              # your annotated structs — source for `goninja generate`
│   ├── book.go
│   └── author.go
├── internal/
│   └── api/              # generated — never hand-edit
│       ├── book_generated.go
│       └── author_generated.go
└── handlers/             # your code: Actions, Authenticators, error mappers
    ├── book.go            # BookActions, publishHandler, bookErrorMapper
    └── auth.go             # your Authenticator implementations
```

This is the shape [Custom Actions](../actions) itself uses for its
`handlers/book.go` example. `examples/prototype` in the goninja repo
flattens it into `package main` instead (`bookpublish.go`,
`taskcomplete.go`, `auth.go`, `errors.go` all directly in the example's
root) — fine for a small demo where a separate package is more ceremony
than the app warrants, but the two rules above still hold even flattened:
each concern still gets its own file, and `main.go` still wires
everything visibly. Reach for the actual `handlers/` package once
there's more than a couple of these files.

## Attaching Actions: goninja.Actions vs SetActions

Both work identically — `goninja.Actions` builds a `<Model>Option` applied
at construction, `SetActions` is a separate call afterward. Reach for
`goninja.Actions` when you're passing something else the action-builder
needs anyway (an `Authenticator`, `Config.StrictAuth` policy) — it's one
line instead of two:

```go
bookAPI := api.NewBookResource(db, goninja.Actions(handlers.BookActions, actionAuth))
```

Use plain `SetActions` when there's nothing extra to thread through, or
when you're conditionally attaching actions after some setup that doesn't
fit neatly into the constructor call:

```go
bookAPI := api.NewBookResource(db)
bookAPI.SetActions(handlers.BookActions(bookAPI)...)
```

Neither is more "correct" — pick whichever reads better at the call site.
See [Custom Actions](../actions) for the full mechanism, including
bundling more than one varying policy into a struct `arg` when different
actions on the same resource need different auth.

## Where to put auth: Action.Auth vs Config.DefaultAuth.Routes

Default to `Action.Auth` for anything action-specific — it's declared
where the action is, so there's nothing to keep in sync elsewhere. Reach
for `Config.DefaultAuth.Routes`/`ResourceConfig.Auth` instead when a
policy is genuinely shared across many routes (all five CRUD routes on
every resource, say) and repeating an `Auth` reference on each would just
be noise. The two compose: `Action.Auth` always wins when set, so a
resource-wide default plus a handful of exceptions via `Action.Auth` is a
normal, expected combination — not a smell.

Turn on `Config.StrictAuth` once an app has more than two or three custom
actions. Below that, eyeballing whether each one is protected is
realistic; above it, it stops being realistic, and a startup panic naming
exactly what's unclassified is worth the (zero, since it's opt-in) cost.
See [Authentication](../auth).

## Error mapping: per-resource vs API.SetErrorMapper

Keep an error type — and its mapper — on the resource that produces it.
`alreadyPublishedError` only ever comes from `Book`'s `publish` action, so
it belongs on `BookResource.SetErrorMapper`, not the app-wide default —
registering it globally would mean every other resource's error handling
silently depends on a check for an error only one of them can ever
produce. Reach for `API.SetErrorMapper` only for an error type that's
genuinely cross-resource: a rate-limit error, a maintenance-mode error,
anything not tied to one model's own domain logic. See
[Errors & Responses](../errors).

## Indexing filter-tagged columns

`filter` doesn't imply "index this" — don't reflexively add `gorm:"index"`
to every field you tag `filter`. An index speeds up reads at a real cost
to writes and storage, and on a low-cardinality column (a three-value
`Status` enum, a `bool`) a single-column index often isn't even used by
the query planner. It earns its cost on:

- **Foreign-key columns** a relation `Preload`s through (`AuthorID`,
  `OrganizationID`) — consistently worth it, since these drive joins on
  every `Retrieve`, not just filtered `List` calls.
- **High-cardinality filter columns** (an email, a UUID, anything close to
  unique per row) — where a single-column index actually narrows the scan.

A `filter`+`order` combination the query planner can't satisfy from a
single-column index (e.g. `?status=open&order=-created_at`) needs a
composite index over exactly that column set to fix — and goninja's
generic filter/order system can't guess which combinations your app
actually uses, so it doesn't try to generate one for you. Measure with
`EXPLAIN`/`EXPLAIN ANALYZE` before adding an index, not on the assumption
that `filter` implies it should exist.
