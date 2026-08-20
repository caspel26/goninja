---
title: Comparison
weight: 2
---

There's no shortage of Go REST tooling; here's honestly where goninja sits
relative to a few well-known ones, rather than a "why we're better than
everyone" table.

|  | [Huma](https://github.com/danielgtaylor/huma) | [ckoliber/gocrud](https://github.com/ckoliber/gocrud) | [nicolasbonnici/gorest](https://github.com/nicolasbonnici/gorest) | **goninja** |
|---|---|---|---|---|
| How CRUD gets built | You write it — Huma gives you typed `Input`/`Output` structs and builds OpenAPI from them via generics, but each operation is hand-registered | Runtime, via `gocrud.Register[Model]` — generics + reflection inspect your struct and wire up Huma operations at startup | Runtime by default (`crud.New[T](db)`); an optional `gorest-codegen` plugin can scaffold routes/DTOs from a DB schema | Static generation — `goninja generate` writes real `<model>_generated.go` files ahead of time, once, before you build |
| What's on disk after setup | Only what you wrote | Only what you wrote — no generated source to read | Only what you wrote, unless you opt into the codegen plugin | Your model plus a `.go` file per model you can open, read, and step through in a debugger |
| Where a typo shows up | Compile time (your handler) / OpenAPI mismatch caught by Huma | Runtime, when the reflection-driven registration hits it | Runtime for the generic path; compile time for codegen'd routes | Compile time — a bad `goninja` tag either doesn't parse or produces code that fails `go build` |
| Router | Router-agnostic (`net/http`, gin, fiber, chi, ...) via `huma.Context` | Whatever Huma supports, plus its own registration layer | Built on Fiber | Plain `net/http`, no router dependency |
| Maturity | Widely used, stable, Go 1.25+ | Small, newer project | Pre-1.0, documented as subject to breaking changes | Pre-alpha — this is the honest asterisk on goninja's own row |

The short version: Huma is the strongest choice if you want full control
over each handler with excellent OpenAPI support and don't mind writing
them by hand; the two CRUD generators get you further out of the box but
resolve your model at request time through generics and reflection;
goninja trades that runtime flexibility for handlers, queries, and
OpenAPI fragments that exist as ordinary Go source before the binary is
even built — nothing to introspect, nothing that can behave differently
under load than it did in a unit test. The cost is real too: goninja is
pre-alpha, has a narrower feature set than any of the three above, and
regenerating after a model change is an extra step `-watch` mode is meant
to make disappear rather than eliminate outright.
