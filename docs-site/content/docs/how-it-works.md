---
title: How It Works
weight: 2
---

goninja has two halves that never run at the same time: a generator you invoke
from the command line, and a small runtime package the generated code calls
into. Understanding which half does what explains most of the design.

## The pipeline

{{% steps %}}

### Parse

`goninja generate` scans the `-models` directory with `go/parser` and
`go/ast` — it reads source text, it does not build or load your package. Any
struct with at least one `goninja`-tagged field becomes a model. Structs with
no tagged fields are skipped silently.

### Build an IR

Each model becomes an intermediate representation: the model's name, its ID
type, and one entry per tagged field carrying the field's Go type, JSON name,
`validate` tag, whether it is a relation, and which verbs it opted into.

Because the IR is decoupled from both parsing and templating, the same
description drives the handlers, the queries and the OpenAPI fragment. That is
why the generated document cannot disagree with the generated code.

### Render and format

The IR is rendered through `text/template`, passed through `go/format`, and
written as `<model>_generated.go` — one file per model. Formatting happens
before the write, so generated output is always gofmt-clean.

### Compile

The result is ordinary Go. Your normal `go build` type-checks the generated
handlers against your models. A tag that names a field wrongly, or a relation
type goninja cannot handle, fails here — not at runtime.

{{% /steps %}}

## What runs at request time

Only plain Go. A generated handler decodes JSON into a typed struct, validates
it against its `validate` tags, builds a GORM query, maps the row into a typed
response struct, and writes JSON. There is no route table built by reflection,
no schema resolved from a type at first request, no `any`-typed model registry.

This is the trade goninja makes. A reflection-driven CRUD library can adapt to
a type it has never seen, at the cost of doing that work per process (or per
request) and of failing at runtime when a type does not fit. goninja needs a
regeneration step whenever a model changes, and in exchange the failure mode
moves to compile time and the request path stays boring.

{{< callout type="info" >}}
Regenerating is the step people forget. Run `goninja generate -watch` during
development and it happens on every save.
{{< /callout >}}

## The two halves

| | Generator | Runtime |
|---|---|---|
| Package | `internal/codegen`, driven by `cmd/goninja` | `goninja` plus `openapi`, `docsui`, `id` |
| Runs | when you invoke the CLI | on every request |
| Sees your models | as source text | as Go types, already compiled |
| Failure shows up as | a generate or build error | an HTTP error response |

The runtime is deliberately small. It holds the pieces that would otherwise be
duplicated into every generated file: `BaseResource` (database access,
transactions, config, the `Self()` dispatch point), the error types and their
mapping to HTTP responses, validation, pagination, and the auth contract.
Generated code calls these directly, so there is no per-model runtime
scaffolding to deduplicate.

## Two views per model, by design

Every model produces four output types, and the split between two of them is a
guarantee rather than a default:

- `<Model>List` is the collection view. It never preloads a relation.
- `<Model>Retrieve` is the detail view. It preloads every relation field it
  carries.

A list endpoint therefore cannot issue one query per row, because the code that
would do it is never generated. If you want a relation in the detail view but
not a join, tag it `byid` and the field becomes a bare ID.

`<Model>Create` and `<Model>Update` are the request bodies, and they are the
only types that carry `validate` tags — validation applies to input only.

All four are separate structs from your GORM model. The model is never reused
as a response type, so a column cannot appear in a response merely because it
exists on the struct.

## Extending without editing generated files

Generated files carry a `DO NOT EDIT` header and are meant to be regenerated
freely. Every extension point works from the outside:

- **Hooks** — implement an interface like `BeforeCreateHook` on a wrapper type
  to run logic inside the operation's transaction.
- **Method overrides** — each resource exposes a `<Model>Ops` interface; wrap
  the resource and replace a single method.
- **Config** — implement `Configurer` to change the mount path or restrict
  which routes are registered.
- **Actions** — attach non-CRUD endpoints with `goninja.Actions` at
  construction, or `SetActions` afterward.

All of these resolve through `SetSelf`/`Self()`, because Go has no dynamic
dispatch through embedding: a generated method has to ask "has someone wrapped
me?" explicitly. See [Hooks & Overrides](../guides/hooks-and-overrides).

## Next

{{< cards >}}
  {{< card link="../reference/tags" title="Struct Tags" icon="tag" subtitle="The full tag vocabulary." >}}
  {{< card link="../reference/generated-code" title="Generated Code" icon="document-text" subtitle="What a generated file contains." >}}
{{< /cards >}}
