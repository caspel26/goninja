---
title: Paths & Route Config
weight: 7
---
The same `SetSelf` wrapper can implement `goninja.Configurer` to override
the resource's mount path or drop routes it shouldn't expose — both
`Register(mux)` and the generated OpenAPI fragment pick this up:

```go
func (r *authorWithAudit) Config() goninja.ResourceConfig {
    return goninja.ResourceConfig{
        Path:   "/v1/authors", // default would be "/authors"
        Routes: []goninja.Route{goninja.RouteList, goninja.RouteRetrieve}, // no create/update/delete
    }
}
```

`Routes` is opt-in restriction, not a list you must spell out in full —
leave it unset (or nil) to keep every route. `ResourceConfig` also carries
a per-route `Auth map[goninja.Route]goninja.RouteAuth` override for the
[global default auth](../auth) — key a route in the map to override it:
give it its own `Auth []goninja.Authenticator` to require different
authenticators than the default, leave `Auth` unset to require the
default authenticators on a route the default policy doesn't otherwise
protect, or set `Public: true` to punch a hole through the default for
that one route.
