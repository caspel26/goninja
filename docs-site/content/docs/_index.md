---
title: Documentation
next: /docs/getting-started
cascade:
  type: docs
---

`goninja` turns an annotated Go struct into a complete REST resource: typed
request and response schemas, `net/http` handlers, GORM-backed queries,
validation, filtering, ordering, pagination and an OpenAPI fragment.

It does this by writing Go source. `goninja generate` reads your models and
emits one `<model>_generated.go` file per model — ordinary code you can read,
diff, and step through in a debugger. Nothing is resolved by reflection at
request time, so a mistake in a struct tag surfaces as a failed build rather
than a surprise in production.

{{< callout type="warning" >}}
goninja is pre-alpha. The API described here is implemented and tested, but it
may change without notice and there is no compatibility guarantee yet.
{{< /callout >}}

## Start here

{{< cards >}}
  {{< card link="getting-started" title="Getting Started" icon="play" subtitle="Install the CLI and go from an empty module to a running API." >}}
  {{< card link="how-it-works" title="How It Works" icon="cog" subtitle="The pipeline from struct tag to generated handler, and what runs at request time." >}}
{{< /cards >}}

## Guides

Task-oriented walkthroughs for each part of the framework.

{{< cards >}}
  {{< card link="guides/querying" title="Filtering, Ordering & Pagination" icon="filter" >}}
  {{< card link="guides/validation" title="Validation" icon="check-circle" >}}
  {{< card link="guides/relations" title="Relations" icon="link" >}}
  {{< card link="guides/errors" title="Errors & Responses" icon="exclamation-circle" >}}
  {{< card link="guides/transactions" title="Transactions" icon="database" >}}
  {{< card link="guides/hooks-and-overrides" title="Hooks & Overrides" icon="puzzle" >}}
  {{< card link="guides/routing" title="Paths & Route Config" icon="map" >}}
  {{< card link="guides/openapi" title="OpenAPI & Docs UI" icon="book-open" >}}
  {{< card link="guides/actions" title="Custom Actions" icon="lightning-bolt" >}}
  {{< card link="guides/auth" title="Authentication" icon="lock-closed" >}}
  {{< card link="guides/testing" title="Testing" icon="beaker" >}}
{{< /cards >}}

## Reference

{{< cards >}}
  {{< card link="reference/tags" title="Struct Tags" icon="tag" subtitle="Every verb and modifier the goninja tag accepts." >}}
  {{< card link="reference/generated-code" title="Generated Code" icon="document-text" subtitle="What lands in each generated file." >}}
  {{< card link="reference/cli" title="CLI" icon="terminal" subtitle="Flags, defaults and watch mode." >}}
  {{< card link="reference/runtime" title="Runtime API" icon="cube" subtitle="The types and functions generated code depends on." >}}
{{< /cards >}}

## Also

{{< cards >}}
  {{< card link="examples" title="Examples" icon="collection" subtitle="Complete, working projects." >}}
  {{< card link="comparison" title="Comparison" icon="scale" subtitle="How goninja differs from Huma, gocrud and gorest." >}}
{{< /cards >}}
