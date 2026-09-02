---
title: Guides
weight: 3
---

Task-oriented walkthroughs. Each page covers one part of the framework and
assumes you have already been through [Getting Started](../getting-started).

## Working with data

{{< cards >}}
  {{< card link="querying" title="Filtering, Ordering & Pagination" icon="filter" subtitle="The filter tag, range filters, the order parameter and the list envelope." >}}
  {{< card link="validation" title="Validation" icon="check-circle" subtitle="validate tags on input types, and registering your own." >}}
  {{< card link="relations" title="Relations" icon="link" subtitle="Belongs-to, has-many, and choosing between nesting and a bare ID." >}}
  {{< card link="transactions" title="Transactions" icon="database" subtitle="Which operations are transactional, and how to join one." >}}
{{< /cards >}}

## Shaping the API

{{< cards >}}
  {{< card link="errors" title="Errors & Responses" icon="exclamation-circle" subtitle="The error types, their status codes, and custom mapping." >}}
  {{< card link="routing" title="Paths & Route Config" icon="map" subtitle="Change a mount path or restrict which routes exist." >}}
  {{< card link="actions" title="Custom Actions" icon="lightning-bolt" subtitle="Non-CRUD endpoints on a resource." >}}
  {{< card link="openapi" title="OpenAPI & Docs UI" icon="book-open" subtitle="The merged document, Swagger UI and ReDoc." >}}
{{< /cards >}}

## Extending and hardening

{{< cards >}}
  {{< card link="hooks-and-overrides" title="Hooks & Overrides" icon="puzzle" subtitle="Run logic around an operation, or replace a generated method." >}}
  {{< card link="auth" title="Authentication" icon="lock-closed" subtitle="Authenticator objects, per-route policy and the built-in schemes." >}}
  {{< card link="testing" title="Testing" icon="beaker" subtitle="Drive a real resource over HTTP against in-memory SQLite." >}}
  {{< card link="benchmarks" title="Performance & Benchmarks" icon="chart-bar" subtitle="Reproducible request-path baselines and the CI regression gate." >}}
  {{< card link="router-adapters" title="Router Adapters" icon="switch-horizontal" subtitle="Mount generated resources on gin, echo, or chi instead of net/http." >}}
  {{< card link="best-practices" title="Best Practices" icon="badge-check" subtitle="Project layout and judgment calls for a real app, not just a demo." >}}
{{< /cards >}}
