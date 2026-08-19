package goninja

import "net/http"

// OpenAPI 3.0 types generated code builds fragments of (plan section
// 5.10/Fase 5). Deliberately a small subset of the spec — just enough to
// describe what goninja actually generates (object schemas, path items with
// get/post/put/delete, query/path parameters, JSON request/response
// bodies) — not a general-purpose OpenAPI library. Property/type shape
// mirrors the JSON spec directly so Spec marshals to a valid document with
// no custom MarshalJSON needed.

// Schema is an OpenAPI schema object, or a $ref to one in
// Components.Schemas. Ref is mutually exclusive with the rest of the
// fields in a valid document, matching how the spec itself works.
type Schema struct {
	Ref        string            `json:"$ref,omitempty"`
	Type       string            `json:"type,omitempty"`
	Format     string            `json:"format,omitempty"`
	Items      *Schema           `json:"items,omitempty"`
	Properties map[string]Schema `json:"properties,omitempty"`
	Required   []string          `json:"required,omitempty"`
}

// Parameter is an OpenAPI path or query parameter.
type Parameter struct {
	Name     string `json:"name"`
	In       string `json:"in"`
	Required bool   `json:"required,omitempty"`
	Schema   Schema `json:"schema"`
}

// MediaType is the value of a Content map, keyed by MIME type (in
// practice always "application/json" for goninja-generated resources).
type MediaType struct {
	Schema Schema `json:"schema"`
}

// RequestBody is an OpenAPI request body object.
type RequestBody struct {
	Required bool                 `json:"required,omitempty"`
	Content  map[string]MediaType `json:"content"`
}

// Response is an OpenAPI response object, keyed by status code in
// Operation.Responses.
type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

// Operation is a single HTTP method entry (get/post/put/delete) on a
// PathItem. Tags groups the operation under one or more named sections in
// a rendered UI (Swagger UI's sidebar, ReDoc's nav, etc) — every operation
// generated for a resource carries the same Tags, set via
// BaseResource.SetOpenAPITags and defaulting to the model name.
type Operation struct {
	Summary     string              `json:"summary,omitempty"`
	Tags        []string            `json:"tags,omitempty"`
	Parameters  []Parameter         `json:"parameters,omitempty"`
	RequestBody *RequestBody        `json:"requestBody,omitempty"`
	Responses   map[string]Response `json:"responses"`
}

// PathItem is the set of operations mounted on one URL path.
type PathItem struct {
	Get    *Operation `json:"get,omitempty"`
	Post   *Operation `json:"post,omitempty"`
	Put    *Operation `json:"put,omitempty"`
	Delete *Operation `json:"delete,omitempty"`
}

// Info is the OpenAPI document's top-level "info" object.
type Info struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

// Components holds the document's reusable schema definitions, referenced
// from operations via Schema.Ref ("#/components/schemas/<name>").
type Components struct {
	Schemas map[string]Schema `json:"schemas,omitempty"`
}

// Spec is a full OpenAPI 3.0 document, as returned by API.Spec and
// marshaled straight to JSON by MountDocs.
type Spec struct {
	OpenAPI    string               `json:"openapi"`
	Info       Info                 `json:"info"`
	Paths      map[string]*PathItem `json:"paths"`
	Components Components           `json:"components"`
}

// OpenAPIProvider is implemented by every generated <Model>Resource: it
// returns the fragment of the document — the paths it mounts and the
// schemas those paths reference — built from the same IR as the rest of
// the generated file (plan section 5.10). API.Add merges the fragment in.
type OpenAPIProvider interface {
	OpenAPI() (paths map[string]*PathItem, schemas map[string]Schema)
}

// Resource is what every generated <Model>Resource implements: it mounts
// its routes on a mux and describes itself as an OpenAPI fragment. Mount
// takes a list of these so callers don't have to pair up a Register(mux)
// call with an Add(doc) call by hand for every resource.
type Resource interface {
	Register(mux *http.ServeMux)
	OpenAPIProvider
}

// Mount registers every resource's routes on mux and merges their OpenAPI
// fragments into doc, in one call — the usual way to wire up all of an
// app's generated resources at once:
//
//	doc := goninja.NewAPI("my api", "0.1.0")
//	goninja.Mount(mux, doc, taskResource, authorResource, bookResource)
//	goninja.MountDocs(mux, doc, "/docs", goninja.SwaggerUI())
//
// doc may be nil to skip OpenAPI entirely — routes are still mounted, just
// nothing is recorded to document (and MountDocs/`/docs` should then not be
// called at all). To leave one resource's routes mounted but keep it out
// of the document specifically, call its BaseResource.ExcludeFromDocs()
// before passing it here — Mount checks for that and skips doc.Add for it.
func Mount(mux *http.ServeMux, doc *API, resources ...Resource) {
	for _, r := range resources {
		r.Register(mux)
		if doc == nil {
			continue
		}
		if x, ok := r.(interface{ docsExcluded() bool }); ok && x.docsExcluded() {
			continue
		}
		doc.Add(r)
	}
}

// API accumulates OpenAPI fragments across every resource registered with
// it, producing one merged document. Construct with NewAPI, call Add once
// per resource (alongside Resource.Register on the mux), then pass to
// MountDocs.
type API struct {
	title, version string
	paths          map[string]*PathItem
	schemas        map[string]Schema
}

// NewAPI creates an empty API document with the given title/version
// (OpenAPI's required top-level "info" fields).
func NewAPI(title, version string) *API {
	return &API{
		title:   title,
		version: version,
		paths:   map[string]*PathItem{},
		schemas: map[string]Schema{},
	}
}

// Add merges p's OpenAPI fragment (its paths and the schemas they
// reference) into the document. Call once per registered resource,
// alongside its Register(mux) call.
func (a *API) Add(p OpenAPIProvider) {
	paths, schemas := p.OpenAPI()
	for path, item := range paths {
		a.paths[path] = item
	}
	for name, schema := range schemas {
		a.schemas[name] = schema
	}
}

// Spec returns the merged OpenAPI document built from every fragment
// added so far.
func (a *API) Spec() Spec {
	return Spec{
		OpenAPI:    "3.0.3",
		Info:       Info{Title: a.title, Version: a.version},
		Paths:      a.paths,
		Components: Components{Schemas: a.schemas},
	}
}
