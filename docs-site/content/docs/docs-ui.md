---
title: Generated Docs UI
weight: 3
---

One call — `app.MountDocs(mux, "/docs", ui)` — serves the merged OpenAPI
document as JSON plus a rendered UI, both fully embedded (no external
CDN). `ui` is an interface, not a hardcoded renderer, so swapping one
line swaps the whole UI:

- `docsui.SwaggerUI()` — the default
- `docsui.ReDoc()` — a drop-in swap

Every operation expands into its request/response schema, complete with
example values generated straight from the model's fields — no
hand-written OpenAPI, ever. See the
[project README](https://github.com/caspel26/goninja#generated-docs-ui)
for screenshots of both UIs, including a custom
[action](../extending/actions) documented alongside generated CRUD
routes.
