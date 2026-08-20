---
title: Reference
weight: 4
---

Exhaustive lookup material: what the tags mean, what the generator emits, what
the CLI accepts, and what the runtime exposes.

{{< cards >}}
  {{< card link="tags" title="Struct Tags" icon="tag" subtitle="Every verb and modifier the goninja tag accepts, and the validate-tag rules." >}}
  {{< card link="generated-code" title="Generated Code" icon="document-text" subtitle="The declarations in a generated file and the OpenAPI type mapping." >}}
  {{< card link="cli" title="CLI" icon="terminal" subtitle="goninja generate flags, defaults and watch mode." >}}
  {{< card link="runtime" title="Runtime API" icon="cube" subtitle="The packages and symbols generated code depends on." >}}
{{< /cards >}}

## Godoc

Signature-level API reference — every exported type, function and method,
across all packages — is published automatically on **pkg.go.dev**:

{{< cards >}}
  {{< card link="https://pkg.go.dev/github.com/caspel26/goninja" title="pkg.go.dev/github.com/caspel26/goninja" icon="external-link" subtitle="Generated godoc for the root package and every subpackage." >}}
{{< /cards >}}

The pages here cover what godoc cannot: how the pieces fit together, which
struct tag drives which behaviour, and why a given default is what it is. Use
pkg.go.dev to look up a signature, and these pages to understand it.

Per-package entry points:

| Package | Reference |
| --- | --- |
| `goninja` | [godoc](https://pkg.go.dev/github.com/caspel26/goninja) · [Runtime API](runtime/) |
| `goninja/openapi` | [godoc](https://pkg.go.dev/github.com/caspel26/goninja/openapi) |
| `goninja/docsui` | [godoc](https://pkg.go.dev/github.com/caspel26/goninja/docsui) |
| `goninja/id` | [godoc](https://pkg.go.dev/github.com/caspel26/goninja/id) |
| `goninja/goninjatest` | [godoc](https://pkg.go.dev/github.com/caspel26/goninja/goninjatest) |
