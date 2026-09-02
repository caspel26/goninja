---
title: Documentation
next: /docs/getting-started
cascade:
  type: docs
---

Choose the path that matches what you are trying to do. If this is your first
visit, start with the five-minute setup; if you already have a resource, jump
straight to the guide or reference you need.

{{< callout type="warning" >}}
goninja is pre-alpha. The API described here is implemented and tested, but it
may change without notice and there is no compatibility guarantee yet.
{{< /callout >}}

## New to goninja

{{< cards >}}
  {{< card link="getting-started" title="Build your first API" icon="play" subtitle="Install the CLI and go from an empty Go module to a working resource." >}}
  {{< card link="how-it-works" title="Understand the generator" icon="cog" subtitle="Follow a struct from parser to generated handler and see what runs per request." >}}
  {{< card link="examples/bookstore" title="Explore a complete project" icon="collection" subtitle="Read a small bookstore API with relations, filters, validation, and docs." >}}
{{< /cards >}}

## Build an API

Shape your resources, queries, relations, and routes.

{{< cards >}}
  {{< card link="guides/querying" title="Filtering, Ordering & Pagination" icon="filter" >}}
  {{< card link="guides/relations" title="Relations" icon="link" >}}
  {{< card link="guides/validation" title="Validation" icon="check-circle" >}}
  {{< card link="guides/routing" title="Paths & Route Config" icon="map" >}}
  {{< card link="guides/router-adapters" title="Router Adapters" icon="switch-horizontal" >}}
  {{< card link="guides/openapi" title="OpenAPI & Docs UI" icon="book-open" >}}
{{< /cards >}}

## Secure and customize

Control access, errors, lifecycle behavior, and non-CRUD endpoints.

{{< cards >}}
  {{< card link="guides/auth" title="Authentication" icon="lock-closed" >}}
  {{< card link="guides/errors" title="Errors & Responses" icon="exclamation-circle" >}}
  {{< card link="guides/actions" title="Custom Actions" icon="lightning-bolt" >}}
  {{< card link="guides/transactions" title="Transactions" icon="database" >}}
  {{< card link="guides/hooks-and-overrides" title="Hooks & Overrides" icon="puzzle" >}}
  {{< card link="guides/best-practices" title="Best Practices" icon="sparkles" >}}
{{< /cards >}}

## Verify and operate

Keep generated resources correct and performance changes measurable.

{{< cards >}}
  {{< card link="guides/testing" title="Testing" icon="beaker" >}}
  {{< card link="guides/benchmarks" title="Performance & Benchmarks" icon="chart-bar" >}}
{{< /cards >}}

## Reference

{{< cards >}}
  {{< card link="reference/tags" title="Struct Tags" icon="tag" subtitle="Every verb and modifier the goninja tag accepts." >}}
  {{< card link="reference/generated-code" title="Generated Code" icon="document-text" subtitle="What lands in each generated file." >}}
  {{< card link="reference/cli" title="CLI" icon="terminal" subtitle="Flags, defaults and watch mode." >}}
  {{< card link="reference/runtime" title="Runtime API" icon="cube" subtitle="The types and functions generated code depends on." >}}
{{< /cards >}}

## Project

{{< cards >}}
  {{< card link="comparison" title="Comparison" icon="scale" subtitle="How goninja differs from Huma, gocrud and gorest." >}}
  {{< card link="changelog" title="Changelog" icon="tag" subtitle="Release notes, upgrades, and breaking changes." >}}
{{< /cards >}}
