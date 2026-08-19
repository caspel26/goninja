// Package openapi holds the OpenAPI 3.0 document types (Schema, Parameter,
// RequestBody, Response, Operation, PathItem, Spec) and the
// OpenAPIProvider interface every generated <Model>Resource implements via
// its OpenAPI() method. It's just the type library — the goninja.API type
// (root package) is what builds a document out of these and mounts
// resources on a router.
package openapi

// OpenAPI 3.0 types generated code builds fragments of. Deliberately a
// small subset of the spec — just enough to describe what goninja actually
// generates (object schemas, path items with get/post/put/delete,
// query/path parameters, JSON request/response bodies) — not a
// general-purpose OpenAPI library. Property/type shape mirrors the JSON
// spec directly so Spec marshals to a valid document with no custom
// MarshalJSON needed.

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
// Operation.Security is a list of alternative requirements — satisfying
// any one entry is enough, matching how goninja tries a route's configured
// Authenticators in order until one succeeds (see BaseResource.SecurityFor,
// resource.go). Each entry's map key is a SecurityScheme name (Authenticator
// name, resolved into Components.SecuritySchemes); the []string scope list
// is always empty for the http/apiKey schemes goninja generates.
type Operation struct {
	Summary     string                `json:"summary,omitempty"`
	Tags        []string              `json:"tags,omitempty"`
	Parameters  []Parameter           `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses"`
	Security    []map[string][]string `json:"security,omitempty"`
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

// SecurityScheme describes one way a request can authenticate, keyed by
// name in Components.SecuritySchemes and referenced from an Operation's
// Security by that same name. goninja.Authenticator.SecurityScheme (root
// package) is the only place one of these gets built — the object that
// enforces auth at runtime is also the sole source of truth for how it's
// documented, so the two can't drift apart.
type SecurityScheme struct {
	// Type is the OpenAPI scheme type: "http" or "apiKey".
	Type string `json:"type"`
	// Scheme is set for Type "http", e.g. "bearer".
	Scheme string `json:"scheme,omitempty"`
	// In and Name are set for Type "apiKey": where the key travels
	// ("header", "query", "cookie") and under what name.
	In   string `json:"in,omitempty"`
	Name string `json:"name,omitempty"`
}

// Components holds the document's reusable schema definitions, referenced
// from operations via Schema.Ref ("#/components/schemas/<name>"), plus its
// named SecuritySchemes, referenced from an Operation's Security.
type Components struct {
	Schemas         map[string]Schema         `json:"schemas,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
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
// returns the fragment of the document — the paths it mounts, the schemas
// those paths reference, and the security schemes referenced by any
// protected path's Security. goninja.API.Add merges the fragment in.
type OpenAPIProvider interface {
	OpenAPI() (paths map[string]*PathItem, schemas map[string]Schema, securitySchemes map[string]SecurityScheme)
}
