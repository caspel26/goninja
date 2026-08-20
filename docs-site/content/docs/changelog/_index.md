---
title: Changelog
weight: 7
description: Every goninja release, what it added, and what it broke.
---

Every release gets its own page, so searching for a version — or for a feature
you remember landing in one — takes you straight to it.

The terse, machine-readable list lives in
[CHANGELOG.md](https://github.com/caspel26/goninja/blob/main/CHANGELOG.md);
these pages carry the context around it.

## Releases

| Version | Date | Notes |
| --- | --- | --- |
| [v0.2.0](v0.2.0/) | 2026-08-20 | Explicit errors, generator validation, versioned docs |
| [v0.1.0](v0.1.0/) | 2026-08-20 | First pre-alpha release |

## Versioning

goninja follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html),
with the usual pre-1.0 caveat: **while the major version is 0, a minor bump
may contain breaking changes.** Pin an exact version if that matters to you.

```shell
go get github.com/caspel26/goninja@v0.2.0
```

These pages document goninja as it currently stands, which may include changes
that are not in any release yet. Every released minor series also keeps a
frozen snapshot of its own documentation — use the version selector in the
navbar to reach it.

{{< cards >}}
  {{< card link="v0.2.0" title="v0.2.0" icon="tag" subtitle="Explicit errors, generator validation, versioned docs." >}}
  {{< card link="v0.1.0" title="v0.1.0" icon="tag" subtitle="First pre-alpha release — CRUD generation, filters, auth, OpenAPI." >}}
{{< /cards >}}
