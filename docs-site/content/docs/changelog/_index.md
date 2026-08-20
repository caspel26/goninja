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
| [v0.1.0](v0.1.0/) | 2026-08-20 | First pre-alpha release |

## Versioning

goninja follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html),
with the usual pre-1.0 caveat: **while the major version is 0, a minor bump
may contain breaking changes.** Pin an exact version if that matters to you.

```shell
go get github.com/caspel26/goninja@v0.1.0
```

Documentation is published per minor series. The version selector in the
navbar switches between them, and `/dev/` tracks the `main` branch — which may
document features that are not in any release yet.

{{< cards >}}
  {{< card link="v0.1.0" title="v0.1.0" icon="tag" subtitle="First pre-alpha release — CRUD generation, filters, auth, OpenAPI." >}}
{{< /cards >}}
