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

<div class="gn-release-list">
  <a class="gn-release gn-release-latest" href="v0.6.1/">
    <span class="gn-release-version">v0.6.1 <em>latest</em></span>
    <span class="gn-release-summary">GORM-aware generated SQL and published benchmark baselines</span>
    <time datetime="2026-09-02">2 Sep 2026</time>
    <span class="gn-release-arrow" aria-hidden="true">→</span>
  </a>
  <a class="gn-release" href="v0.6.0/">
    <span class="gn-release-version">v0.6.0</span>
    <span class="gn-release-summary">Leaner list queries and shared OpenAPI construction</span>
    <time datetime="2026-08-27">27 Aug 2026</time>
    <span class="gn-release-arrow" aria-hidden="true">→</span>
  </a>
  <a class="gn-release" href="v0.5.0/">
    <span class="gn-release-version">v0.5.0</span>
    <span class="gn-release-summary">Explicit action auth, strict auth, and time filters</span>
    <time datetime="2026-08-22">22 Aug 2026</time>
    <span class="gn-release-arrow" aria-hidden="true">→</span>
  </a>
  <a class="gn-release" href="v0.4.0/">
    <span class="gn-release-version">v0.4.0</span>
    <span class="gn-release-summary">Application-level error mapping</span>
    <time datetime="2026-08-21">21 Aug 2026</time>
    <span class="gn-release-arrow" aria-hidden="true">→</span>
  </a>
  <a class="gn-release" href="v0.3.1/">
    <span class="gn-release-version">v0.3.1</span>
    <span class="gn-release-summary">SpecProvider rename and CI cleanup</span>
    <time datetime="2026-08-20">20 Aug 2026</time>
    <span class="gn-release-arrow" aria-hidden="true">→</span>
  </a>
  <a class="gn-release" href="v0.3.0/">
    <span class="gn-release-version">v0.3.0</span>
    <span class="gn-release-summary">Gin, Echo, and Chi adapters with benchmark tooling</span>
    <time datetime="2026-08-20">20 Aug 2026</time>
    <span class="gn-release-arrow" aria-hidden="true">→</span>
  </a>
  <a class="gn-release" href="v0.2.0/">
    <span class="gn-release-version">v0.2.0</span>
    <span class="gn-release-summary">Explicit errors, generator validation, and versioned docs</span>
    <time datetime="2026-08-20">20 Aug 2026</time>
    <span class="gn-release-arrow" aria-hidden="true">→</span>
  </a>
  <a class="gn-release" href="v0.1.0/">
    <span class="gn-release-version">v0.1.0</span>
    <span class="gn-release-summary">First pre-alpha release</span>
    <time datetime="2026-08-20">20 Aug 2026</time>
    <span class="gn-release-arrow" aria-hidden="true">→</span>
  </a>
</div>

## Versioning

goninja follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html),
with the usual pre-1.0 caveat: **while the major version is 0, a minor bump
may contain breaking changes.** Pin an exact version if that matters to you.

```shell
go get github.com/caspel26/goninja@v0.6.1
```

These pages document goninja as it currently stands, which may include changes
that are not in any release yet. Every released minor series also keeps a
frozen snapshot of its own documentation — use the version selector in the
navbar to reach it.
