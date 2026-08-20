---
title: Examples
weight: 6
---

This section walks through a complete, runnable goninja project rather
than isolated snippets. The full source lives in this repository under
`examples/prototype` — three models, a real Postgres-backed server, a
custom action, and a test — and can be regenerated and run locally with
`make generate-prototype` / `make run-prototype` (see the Makefile at the
repo root; `run-prototype` needs `PROTOTYPE_DSN` set, e.g.
`export PROTOTYPE_DSN="host=localhost user=$(whoami) dbname=goninja_prototype sslmode=disable"`).

{{< cards >}}
{{< card link="bookstore" title="Bookstore API" icon="book-open" subtitle="Tasks, authors, and books wired end to end against Postgres, with a custom publish action and generated docs." >}}
{{< /cards >}}
