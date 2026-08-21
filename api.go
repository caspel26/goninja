// API is goninja's application entry point: it accumulates every mounted
// resource's OpenAPI fragment into one merged document and mounts
// resources onto a Router. It owns that logic directly rather than
// wrapping a separate document type — openapi just supplies the wire
// types (Schema, PathItem, Spec, ...) it's built from. Docs are an
// optional layer on top (MountDocs), not the other way around.
package goninja

import (
	"github.com/caspel26/goninja/docsui"
	"github.com/caspel26/goninja/openapi"
	"github.com/caspel26/goninja/router"
)

// Router is what a generated Register method mounts routes onto —
// see the router package. *http.ServeMux satisfies it as-is; adapters for
// gin, echo, and chi (github.com/caspel26/goninja/adapters/...) satisfy it
// too, so the same generated code mounts on any of them unchanged.
type Router = router.Router

// Resource is what every generated <Model>Resource implements: it mounts
// its routes on a Router and describes itself as an OpenAPI fragment. Mount
// and MountWithConfig take a list of these so callers don't have to pair
// up a Register(mux) call with an Add call by hand for every resource.
type Resource interface {
	Register(mux Router)
	openapi.OpenAPIProvider
}

// API accumulates every mounted resource's OpenAPI fragment into one
// merged document. Construct with NewAPI, then Mount/MountWithConfig your
// generated resources onto it, and optionally MountDocs to serve a
// rendered UI over the result:
//
//	api := goninja.NewAPI("Bookstore API", "0.1.0")
//	api.Mount(mux, taskResource, authorResource, bookResource)
//	api.MountDocs(mux, "/docs", docsui.SwaggerUI())
type API struct {
	title, version  string
	paths           map[string]*openapi.PathItem
	schemas         map[string]openapi.Schema
	securitySchemes map[string]openapi.SecurityScheme
	errorMapper     ErrorMapper
}

// NewAPI creates a new API document with the given title/version (OpenAPI's
// required top-level "info" fields).
func NewAPI(title, version string) *API {
	return &API{
		title:           title,
		version:         version,
		paths:           map[string]*openapi.PathItem{},
		schemas:         map[string]openapi.Schema{},
		securitySchemes: map[string]openapi.SecurityScheme{},
	}
}

// SetErrorMapper sets an app-wide default ErrorMapper from one or more
// ErrorMappings (build each with NewErrorMapping), applied by Mount and
// MountWithConfig to every resource that hasn't set its own via
// BaseResource.SetErrorMapper — one place to register per-error-type
// handling for the whole app, rather than building a Config by hand just
// to reach MountWithConfig.
// Takes ErrorMappings rather than a whole ErrorMapper so mappings composed
// from different files/resources merge into one list safely: a plain
// ErrorMapper has no way to say "I didn't recognize this error, try the
// next one" (DefaultErrorMapper, for instance, answers every error), so
// chaining whole ErrorMappers would let an earlier one silently swallow
// everything. An ErrorMapping's own Matches avoids that. Calling this more
// than once replaces the previous mappings, it doesn't add to them. A
// Config.DefaultErrorMapper passed explicitly to MountWithConfig still
// wins over this if both are set — same "more specific wins" rule as
// everything else Config touches.
func (a *API) SetErrorMapper(mappings ...ErrorMapping) {
	a.errorMapper = NewErrorMapper(mappings...)
}

// Add merges p's OpenAPI fragment (its paths, the schemas they reference,
// and the security schemes any protected path's Security refers to) into
// a's document. Mount and MountWithConfig call this once per resource
// automatically; call it directly only when wiring a resource onto mux by
// hand instead of through them.
func (a *API) Add(p openapi.OpenAPIProvider) {
	paths, schemas, securitySchemes := p.OpenAPI()
	for path, item := range paths {
		a.paths[path] = item
	}
	for name, schema := range schemas {
		a.schemas[name] = schema
	}
	for name, scheme := range securitySchemes {
		a.securitySchemes[name] = scheme
	}
}

// Spec returns the merged OpenAPI document built from every resource
// mounted onto a so far.
func (a *API) Spec() openapi.Spec {
	return openapi.Spec{
		OpenAPI:    "3.0.3",
		Info:       openapi.Info{Title: a.title, Version: a.version},
		Paths:      a.paths,
		Components: openapi.Components{Schemas: a.schemas, SecuritySchemes: a.securitySchemes},
	}
}

// Mount registers every resource's routes on mux and merges their OpenAPI
// fragments into a, in one call. If SetErrorMapper was called on a, every
// resource gets it via SetConfig — the only Config field a plain Mount
// applies; DefaultAuth/Middleware still need MountWithConfig. Call a
// resource's BaseResource.ExcludeFromDocs() before passing it here to keep
// its routes mounted but leave it out of the document; call its
// Register(mux) directly instead of going through Mount to skip
// documenting it entirely.
func (a *API) Mount(mux Router, resources ...Resource) {
	for _, r := range resources {
		if a.errorMapper != nil {
			if x, ok := r.(interface{ SetConfig(Config) }); ok {
				x.SetConfig(Config{DefaultErrorMapper: a.errorMapper})
			}
		}
		r.Register(mux)
		if x, ok := r.(interface{ DocsExcluded() bool }); ok && x.DocsExcluded() {
			continue
		}
		a.Add(r)
	}
}

// MountWithConfig is Mount plus a Config: it calls BaseResource.SetConfig
// on every resource before Register(mux), so cfg's DefaultAuth and
// Middleware take effect when each resource's generated Register builds
// its handlers. Use it instead of Mount once the app has a global auth
// policy or middleware to enforce. cfg.DefaultErrorMapper, if set, wins
// over a's own SetErrorMapper; otherwise a's applies here too.
func (a *API) MountWithConfig(mux Router, cfg Config, resources ...Resource) {
	if cfg.DefaultErrorMapper == nil {
		cfg.DefaultErrorMapper = a.errorMapper
	}
	for _, r := range resources {
		if x, ok := r.(interface{ SetConfig(Config) }); ok {
			x.SetConfig(cfg)
		}
		r.Register(mux)
		if x, ok := r.(interface{ DocsExcluded() bool }); ok && x.DocsExcluded() {
			continue
		}
		a.Add(r)
	}
}

// MountDocs serves a's merged OpenAPI document plus a rendered docs UI at
// path — see docsui.MountDocs, which this forwards to. ui selects the
// renderer (docsui.SwaggerUI(), docsui.ReDoc(), or your own
// docsui.DocsUI); nil defaults to Swagger UI.
func (a *API) MountDocs(mux Router, path string, ui docsui.DocsUI) {
	docsui.MountDocs(mux, a, path, ui)
}
