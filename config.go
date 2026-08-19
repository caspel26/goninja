// Config is the app-wide policy API.MountWithConfig (api.go) threads into
// every resource: a global default auth policy plus generic middleware.
// Each generated resource owns its own New<Model>Resource(db)/Register(mux),
// and resources are wired up via API.Mount/MountWithConfig rather than a
// central Register call, so Config rides along MountWithConfig and
// BaseResource.SetConfig — consistent with how ResourceConfig/Configurer
// (resource_config.go) and SetSelf (resource.go) extend a resource from
// outside without a central object owning it. This is what actually
// enforces ResourceConfig.Auth (resource_config.go) and gives WithUser/
// UserFromContext (auth.go) a reason to exist beyond carrying a value
// through the request.
package goninja

import "net/http"

// AuthPolicy is a global default auth policy (plan section 5.15): which
// routes require auth by default, and the Authenticator(s) that enforce
// it, tried in order until one recognizes the request.
type AuthPolicy struct {
	// Routes names routes that require auth by default, before any
	// per-resource ResourceConfig.Auth override is applied.
	Routes []Route

	// Auth is the Authenticator(s) enforcing Routes, tried in order until
	// one returns ok=true (e.g. bearer-or-api-key). Only consulted for a
	// route this policy (combined with a resource's own ResourceConfig.Auth)
	// actually protects — a public route never sees it.
	Auth []Authenticator
}

// Config is the global policy API.MountWithConfig applies to every
// resource passed to it.
type Config struct {
	// DefaultAuth is the auth policy every resource is checked against,
	// unless a resource's own ResourceConfig.Auth entry for that route
	// overrides it — see RouteAuth.
	DefaultAuth AuthPolicy

	// Middleware wraps every route on every resource, regardless of auth
	// (logging, CORS, request IDs, ...) — it always runs, protected or
	// public, and runs outside whatever Authenticator(s) DefaultAuth applies.
	Middleware []func(http.Handler) http.Handler
}
