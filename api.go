// API is goninja's application entry point: it accumulates every mounted
// resource's OpenAPI fragment into one merged document and mounts
// resources onto an *http.ServeMux. It owns that logic directly rather
// than wrapping a separate document type — openapi just supplies the wire
// types (Schema, PathItem, Spec, ...) it's built from. Docs are an
// optional layer on top (MountDocs), not the other way around.
package goninja

import (
	"net/http"

	"github.com/caspel26/goninja/docsui"
	"github.com/caspel26/goninja/openapi"
)

// Resource is what every generated <Model>Resource implements: it mounts
// its routes on a mux and describes itself as an OpenAPI fragment. Mount
// and MountWithConfig take a list of these so callers don't have to pair
// up a Register(mux) call with an Add call by hand for every resource.
type Resource interface {
	Register(mux *http.ServeMux)
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
	title, version string
	paths          map[string]*openapi.PathItem
	schemas        map[string]openapi.Schema
}

// NewAPI creates a new API document with the given title/version (OpenAPI's
// required top-level "info" fields).
func NewAPI(title, version string) *API {
	return &API{
		title:   title,
		version: version,
		paths:   map[string]*openapi.PathItem{},
		schemas: map[string]openapi.Schema{},
	}
}

// Add merges p's OpenAPI fragment (its paths and the schemas they
// reference) into a's document. Mount and MountWithConfig call this once
// per resource automatically; call it directly only when wiring a resource
// onto mux by hand instead of through them.
func (a *API) Add(p openapi.OpenAPIProvider) {
	paths, schemas := p.OpenAPI()
	for path, item := range paths {
		a.paths[path] = item
	}
	for name, schema := range schemas {
		a.schemas[name] = schema
	}
}

// Spec returns the merged OpenAPI document built from every resource
// mounted onto a so far.
func (a *API) Spec() openapi.Spec {
	return openapi.Spec{
		OpenAPI:    "3.0.3",
		Info:       openapi.Info{Title: a.title, Version: a.version},
		Paths:      a.paths,
		Components: openapi.Components{Schemas: a.schemas},
	}
}

// Mount registers every resource's routes on mux and merges their OpenAPI
// fragments into a, in one call. Call a resource's
// BaseResource.ExcludeFromDocs() before passing it here to keep its routes
// mounted but leave it out of the document; call its Register(mux)
// directly instead of going through Mount to skip documenting it entirely.
func (a *API) Mount(mux *http.ServeMux, resources ...Resource) {
	for _, r := range resources {
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
// policy or middleware to enforce.
func (a *API) MountWithConfig(mux *http.ServeMux, cfg Config, resources ...Resource) {
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
func (a *API) MountDocs(mux *http.ServeMux, path string, ui docsui.DocsUI) {
	docsui.MountDocs(mux, a, path, ui)
}
