package mw

// ResourceConfig lets a resource wrapper customize how the generated
// Register(mux) and OpenAPI() methods mount and document a resource's
// routes, and how the (not yet built, Fase 6 item 5) global default auth
// applies to them — plan section 5.3. A resource picks this up by
// implementing Configurer on the value passed to SetSelf, the same
// dispatch hooks.go and method overrides use (plan section 5.10): a
// resource used directly, with no wrapper, gets every generated default.
type ResourceConfig struct {
	// Path overrides the resource's base route path (e.g. "/v1/books"),
	// with "/{id}" appended for the retrieve/update/delete routes. Empty
	// keeps the generated default (e.g. "/books").
	Path string

	// Routes restricts which of "list", "retrieve", "create", "update",
	// "delete" get mounted/documented. Empty means all of them — Routes is
	// an opt-in restriction, not an enable list you must spell out in full
	// just to keep every route.
	Routes []string

	// Auth is an additive-only override of the global default auth. See
	// AuthOverride.
	Auth AuthOverride
}

// AuthOverride is additive-only by design (plan section 5.3): per-resource
// config can only ever add protection relative to the future global
// default auth (Config.DefaultAuth, Fase 6 item 5, not yet built), never
// silently remove it. AlsoProtect names routes to protect beyond the
// default; Public names routes to make public, but only takes effect for a
// route explicitly listed there — a route missing from AlsoProtect is left
// exactly as the global default treats it, and a route missing from Public
// is never made public by omission.
type AuthOverride struct {
	AlsoProtect []string
	Public      []string
}

// Configurer is the optional interface a resource wrapper implements to
// customize its ResourceConfig, mirroring hooks.go's BeforeCreateHook etc:
// generated Register/OpenAPI methods check r.Self() for it and fall back
// to ResourceConfig{} (every generated default) when absent.
type Configurer interface {
	Config() ResourceConfig
}

// RouteEnabled reports whether route (one of "list", "retrieve", "create",
// "update", "delete") should be mounted/documented per cfg.Routes.
func (cfg ResourceConfig) RouteEnabled(route string) bool {
	if len(cfg.Routes) == 0 {
		return true
	}
	for _, r := range cfg.Routes {
		if r == route {
			return true
		}
	}
	return false
}

// PathOr returns cfg.Path if set, else def — the generated default path.
func (cfg ResourceConfig) PathOr(def string) string {
	if cfg.Path == "" {
		return def
	}
	return cfg.Path
}
