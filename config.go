// Config is the app-wide policy API.MountWithConfig (api.go) threads into
// every resource: a global default auth policy plus generic middleware.
// Each generated resource owns its own New<Model>Resource(db)/Register(mux),
// and resources are wired up via API.Mount/MountWithConfig rather than a
// central Register call, so Config rides along MountWithConfig and
// BaseResource.SetConfig — consistent with how ResourceConfig/Configurer
// (resource_config.go) and SetSelf (resource.go) extend a resource from
// outside without a central object owning it. This is what actually
// enforces AuthOverride (resource_config.go) and gives WithUser/
// UserFromContext (auth.go) a reason to exist beyond carrying a value
// through the request.
package goninja

import "net/http"

// AuthPolicy is a global default auth policy: which routes ("list",
// "retrieve", "create", "update", "delete") require auth by default, and
// the middleware chain that enforces it. That middleware is yours — it
// typically authenticates the request and calls goninja.WithUser (auth.go)
// before letting it through, and rejects it (401) otherwise.
type AuthPolicy struct {
	// Protected names routes that require auth by default, before any
	// per-resource AuthOverride (resource_config.go) is applied.
	Protected []string

	// Middleware is the chain that enforces Protected. Only wraps routes
	// this policy (combined with a resource's own AuthOverride) actually
	// protects — a public route never sees it.
	Middleware []func(http.Handler) http.Handler
}

// Config is the global policy API.MountWithConfig applies to every
// resource passed to it.
type Config struct {
	// DefaultAuth is the auth policy every resource is checked against,
	// unless a resource's own ResourceConfig.Auth (an AuthOverride) adds to
	// or explicitly punches a hole in it — additive-only, see AuthOverride.
	DefaultAuth AuthPolicy

	// Middleware wraps every route on every resource, regardless of auth
	// (logging, CORS, request IDs, ...) — it always runs, protected or
	// public, and runs outside DefaultAuth.Middleware.
	Middleware []func(http.Handler) http.Handler
}
