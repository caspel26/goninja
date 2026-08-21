package goninja

// ResourceConfig lets a resource wrapper customize how the generated
// Register(mux) and OpenAPI() methods mount and document a resource's
// routes, and how the global default auth (Config.DefaultAuth) applies to
// them. A resource picks this up by implementing
// Configurer on the value passed to SetSelf, the same dispatch hooks.go and
// method overrides use: a resource used directly, with
// no wrapper, gets every generated default.
type ResourceConfig struct {
	// Path overrides the resource's base route path (e.g. "/v1/books"),
	// with "/{id}" appended for the retrieve/update/delete routes. Empty
	// keeps the generated default (e.g. "/books").
	Path string

	// Routes restricts which of RouteList/RouteRetrieve/RouteCreate/
	// RouteUpdate/RouteDelete get mounted/documented. Empty means all of
	// them — Routes is an opt-in restriction, not an enable list you must
	// spell out in full just to keep every route.
	Routes []Route

	// Auth overrides Config.DefaultAuth per route: a Route present as a key
	// here replaces how that route is authenticated; a Route absent from
	// this map is left exactly as Config.DefaultAuth treats it. See
	// RouteAuth.
	Auth map[Route]RouteAuth
}

// RouteAuth is one route's override entry in ResourceConfig.Auth. Its
// presence in the map is what makes it an override —
// there is no separate "enabled" flag to forget: Public=true and
// unset/empty Auth both mean "public", so a resource wrapper composing this
// map has exactly one way to express each outcome, not two that could
// disagree.
type RouteAuth struct {
	// Auth, if non-empty, replaces Config.DefaultAuth.Auth for this route —
	// tried in order until one Authenticator returns ok=true, same as the
	// global policy. Ignored when Public is true.
	Auth []Authenticator

	// Public, when true, makes this route require no auth at all,
	// regardless of what Config.DefaultAuth says — the explicit opt-out a
	// resource needs to punch a hole in a global default.
	Public bool
}

// Configurer is the optional interface a resource wrapper implements to
// customize its ResourceConfig, mirroring hooks.go's BeforeCreateHook etc:
// generated Register/OpenAPI methods check r.Self() for it and fall back
// to ResourceConfig{} (every generated default) when absent.
type Configurer interface {
	Config() ResourceConfig
}

// RouteEnabled reports whether route should be mounted/documented per
// cfg.Routes.
func (cfg ResourceConfig) RouteEnabled(route Route) bool {
	if len(cfg.Routes) == 0 {
		return true
	}
	return containsRoute(cfg.Routes, route)
}

// PathOr returns cfg.Path if set, else def — the generated default path.
func (cfg ResourceConfig) PathOr(def string) string {
	if cfg.Path == "" {
		return def
	}
	return cfg.Path
}
