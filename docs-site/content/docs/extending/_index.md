---
title: Extending a Resource
weight: 3
---

Nothing generated is meant to be edited by hand — every extension point
works by embedding a generated `<Model>Resource` in your own type. Go has
no dynamic dispatch through embedding, so a wrapper's overrides only take
effect once you point the embedded resource's `Self()` at it.

See the pages in this section for hooks, validation, routing, custom
actions, auth, relations, and testing.
