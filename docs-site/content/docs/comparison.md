---
title: Comparison
weight: 2
---

There's no shortage of Go REST tooling. This page is an honest placement of
goninja next to three well-known alternatives — including where they win —
rather than a scoreboard.

## At a glance

| | [Huma](https://github.com/danielgtaylor/huma) | [gocrud](https://github.com/ckoliber/gocrud) | [gorest](https://github.com/nicolasbonnici/gorest) | **goninja** |
| --- | --- | --- | --- | --- |
| **CRUD handlers** | You write each one | Generated at startup | Generated at startup | Generated as source |
| **Model resolved** | Compile time | Startup, by reflection | Request time | Before `go build` |
| **Source on disk** | Yours only | Yours only | Yours only | Yours **+ one file per model** |
| **A bad field shows up** | Compile time | Runtime | Runtime | Compile time |
| **Router** | Any (`net/http`, gin, chi, fiber) | Huma's | Fiber only | `net/http`, gin, echo, chi |
| **OpenAPI** | Built in, excellent | From Huma | Partial | Built in, from the same IR |
| **Debuggable handlers** | Yes — you wrote them | No source to step into | No source to step into | Yes — plain generated Go |
| **Maturity** | Stable, widely used | Early | Pre-1.0, breaking changes | **Pre-alpha** |

## Reading the table

**[Huma](https://github.com/danielgtaylor/huma)** is the strongest choice if you
want full control of every handler. It gives you typed `Input`/`Output` structs
and derives an excellent OpenAPI document from them through generics, but each
operation is registered by hand — it deliberately doesn't generate CRUD for you.
Pick it when the endpoints are the interesting part.

**[gocrud](https://github.com/ckoliber/gocrud)** and
**[gorest](https://github.com/nicolasbonnici/gorest)** both get you a CRUD
surface with less code than goninja, because they resolve your model through
generics and reflection instead of writing files. That's a genuine ergonomic
win — there's no generate step and nothing to regenerate. The tradeoff is that
the handlers only exist at runtime: there's no source to read, diff in review,
or set a breakpoint inside, and a mistake in a struct surfaces when the code
runs rather than when it compiles.

**goninja** moves that work to generation time. `goninja generate` writes a
`<model>_generated.go` per model — handlers, queries, output types and an
OpenAPI fragment — as ordinary Go you commit. Nothing is reflected on the
request path, so what a unit test exercises is what runs under load, and a bad
`goninja` tag either fails to parse or produces code that fails `go build`.

## Where goninja costs you

Stated plainly, because the table above rounds in its favour:

- **It's pre-alpha.** The API can change without notice, and there's no
  compatibility guarantee yet.
- **The feature set is narrower** than any of the three above.
- **Generation is a step.** Change a model and you re-run the generator.
  [`-watch`](../reference/cli/) makes that automatic, but it doesn't remove the
  step — it just stops you having to remember it.
- **Generated files are in your repo.** Some teams consider that noise; goninja
  considers it the point.
- **GORM is assumed.** There's no adapter layer for other ORMs. Routing has
  one — gin, echo, and chi all mount the same generated code unchanged via
  [Router Adapters](../guides/router-adapters/) — but anything outside those
  four (fiber, gorilla, ...) still needs one written.

If reflection-driven CRUD already works for you, none of the above is worth
switching for. goninja is for the case where you'd rather read the code that
serves your API than trust that it exists.
