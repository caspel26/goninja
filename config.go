// Config is the app-wide policy MountWithConfig threads into every
// resource: the global default auth and generic middleware the plan's
// section 5.2 originally sketched as a central goninja.New(Config)/
// api.Register(mux, resource) object. goninja's actual shape diverged from
// that sketch — each generated resource already owns its own
// New<Model>Resource(db)/Register(mux), and resources are wired up via
// Mount, not a central Register call — so Config instead rides along
// MountWithConfig (a sibling to Mount, openapi.go) and BaseResource.SetConfig,
// consistent with how ResourceConfig/Configurer (resource_config.go) and
// SetSelf (resource.go) already extend a resource from outside without a
// central object owning it.
//
// This is Fase 6 item 5, and the first thing that actually *enforces*
// AuthOverride (resource_config.go) and gives WithUser/UserFromContext
// (auth.go) a reason to exist beyond carrying a value through the request.
package goninja

import (
	"net/http"

	"github.com/caspel26/goninja/openapi"
)

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

// Config is the global policy MountWithConfig applies to every resource
// passed to it.
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

// MountWithConfig is Mount (openapi.go) plus a Config: it calls
// BaseResource.SetConfig(cfg) on every resource before Register(mux), so
// DefaultAuth and Middleware actually take effect when each resource's
// generated Register builds its handlers. Plain Mount leaves every
// resource's Config at its zero value — no global auth, no global
// middleware — so use MountWithConfig instead of Mount once an app has a
// global policy to enforce. Otherwise identical to Mount, including the
// doc == nil / ExcludeFromDocs behavior.
//
//	cfg := goninja.Config{
//	    DefaultAuth: goninja.AuthPolicy{
//	        Protected:  []string{"create", "update", "delete"},
//	        Middleware: []func(http.Handler) http.Handler{JWTMiddleware(secret)},
//	    },
//	    Middleware: []func(http.Handler) http.Handler{LoggingMiddleware()},
//	}
//	goninja.MountWithConfig(mux, doc, cfg, taskResource, authorResource)
func MountWithConfig(mux *http.ServeMux, doc *openapi.API, cfg Config, resources ...openapi.Resource) {
	for _, r := range resources {
		if x, ok := r.(interface{ SetConfig(Config) }); ok {
			x.SetConfig(cfg)
		}
		r.Register(mux)
		if doc == nil {
			continue
		}
		if x, ok := r.(interface{ DocsExcluded() bool }); ok && x.DocsExcluded() {
			continue
		}
		doc.Add(r)
	}
}
