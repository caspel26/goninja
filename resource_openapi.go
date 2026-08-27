package goninja

import (
	"net/http"
	"strings"

	"github.com/caspel26/goninja/openapi"
)

// ResourceDoc is the per-model half of a generated resource's OpenAPI
// fragment: the parts that genuinely describe *this* model, which only the
// generator knows. Everything else about the fragment — the path/operation
// structure, the list envelope, the shared limit/offset/order parameters,
// the id path parameter, per-route security — is identical for every model
// and lives in BuildResourceOpenAPI instead of being generated into each
// <model>_generated.go.
//
// This mirrors how the rest of the runtime already works: a generated
// handler calls RespondJSON/Validate rather than carrying its own copy, so
// a fix lands once instead of per model. OpenAPI construction was the one
// place that didn't follow that, at ~165 duplicated lines per model.
type ResourceDoc struct {
	// Name is the model's Go name, e.g. "Author" — used for the default
	// OpenAPI tag, schema $refs, and operation summaries.
	Name string
	// NameLower is the lowercase model name, e.g. "author" — used for the
	// default base path ("/authors") and in operation summaries.
	NameLower string
	// IDSchema describes the {id} path parameter: integer/int64 for an
	// int64 ID, string otherwise.
	IDSchema openapi.Schema
	// Schemas holds this model's List/Retrieve/Create/Update schemas,
	// keyed by component name. BuildResourceOpenAPI adds the
	// "<Name>ListEnvelope" wrapper itself, since only the inner $ref
	// varies.
	Schemas map[string]openapi.Schema
	// ListParams holds the query parameters derived from this model's
	// `filter` fields. BuildResourceOpenAPI appends the limit/offset/order
	// parameters every list route shares.
	ListParams []openapi.Parameter
}

// BuildResourceOpenAPI assembles a resource's full OpenAPI fragment from
// doc plus the routing/auth state on r — the paths, the schemas (doc's own
// plus the generated list envelope), and every SecurityScheme the
// resource's protected routes refer to. Generated OpenAPI() methods return
// its result directly.
//
// cfg gates which routes appear, exactly as it gates which routes
// Register mounts, so the document can never advertise a route that isn't
// actually served.
func BuildResourceOpenAPI(r *BaseResource, cfg ResourceConfig, doc ResourceDoc) (
	map[string]*openapi.PathItem, map[string]openapi.Schema, map[string]openapi.SecurityScheme,
) {
	tags := r.OpenAPITags()
	if len(tags) == 0 {
		tags = []string{doc.Name}
	}

	schemas := make(map[string]openapi.Schema, len(doc.Schemas)+1)
	for name, s := range doc.Schemas {
		schemas[name] = s
	}
	schemas[doc.Name+"ListEnvelope"] = openapi.Schema{
		Type: "object",
		Properties: map[string]openapi.Schema{
			"items":  {Type: "array", Items: &openapi.Schema{Ref: schemaRef(doc.Name + "List")}},
			"total":  {Type: "integer", Format: "int64"},
			"limit":  {Type: "integer"},
			"offset": {Type: "integer"},
		},
	}

	listParams := make([]openapi.Parameter, 0, len(doc.ListParams)+3)
	listParams = append(listParams, doc.ListParams...)
	listParams = append(listParams,
		openapi.Parameter{Name: "limit", In: "query", Schema: openapi.Schema{Type: "integer"}},
		openapi.Parameter{Name: "offset", In: "query", Schema: openapi.Schema{Type: "integer"}},
		openapi.Parameter{Name: "order", In: "query", Schema: openapi.Schema{Type: "string"}},
	)

	idParam := openapi.Parameter{Name: "id", In: "path", Required: true, Schema: doc.IDSchema}

	basePath := cfg.PathOr("/" + doc.NameLower + "s")
	paths := map[string]*openapi.PathItem{}
	schemes := map[string]openapi.SecurityScheme{}

	if item := doc.basePathItem(r, cfg, tags, listParams, schemes); item != nil {
		paths[basePath] = item
	}
	if item := doc.itemPathItem(r, cfg, tags, idParam, schemes); item != nil {
		paths[basePath+"/{id}"] = item
	}
	doc.actionPaths(r, cfg, paths, basePath, idParam, tags, schemes)

	return paths, schemas, schemes
}

// basePathItem builds the list/create PathItem for the collection path,
// or nil when cfg enables neither route.
func (d ResourceDoc) basePathItem(
	r *BaseResource, cfg ResourceConfig, tags []string,
	listParams []openapi.Parameter, schemes map[string]openapi.SecurityScheme,
) *openapi.PathItem {
	item := &openapi.PathItem{}
	if cfg.RouteEnabled(RouteList) {
		item.Get = &openapi.Operation{
			Summary:    "List " + d.NameLower + "s",
			Tags:       tags,
			Parameters: listParams,
			Security:   security(r, RouteList, cfg, schemes),
			Responses: map[string]openapi.Response{
				"200": jsonResponse("OK", schemaRef(d.Name+"ListEnvelope")),
			},
		}
	}
	if cfg.RouteEnabled(RouteCreate) {
		item.Post = &openapi.Operation{
			Summary:     "Create " + article(d.NameLower) + " " + d.NameLower,
			Tags:        tags,
			Security:    security(r, RouteCreate, cfg, schemes),
			RequestBody: jsonBody(schemaRef(d.Name + "Create")),
			Responses: map[string]openapi.Response{
				"201": jsonResponse("Created", schemaRef(d.Name+"Retrieve")),
				"422": {Description: "Validation error"},
			},
		}
	}
	if item.Get == nil && item.Post == nil {
		return nil
	}
	return item
}

// itemPathItem builds the retrieve/update/delete PathItem for the "/{id}"
// path, or nil when cfg enables none of the three.
func (d ResourceDoc) itemPathItem(
	r *BaseResource, cfg ResourceConfig, tags []string,
	idParam openapi.Parameter, schemes map[string]openapi.SecurityScheme,
) *openapi.PathItem {
	params := []openapi.Parameter{idParam}
	notFound := openapi.Response{Description: "Not found"}

	item := &openapi.PathItem{}
	if cfg.RouteEnabled(RouteRetrieve) {
		item.Get = &openapi.Operation{
			Summary:    "Retrieve " + article(d.NameLower) + " " + d.NameLower,
			Tags:       tags,
			Parameters: params,
			Security:   security(r, RouteRetrieve, cfg, schemes),
			Responses: map[string]openapi.Response{
				"200": jsonResponse("OK", schemaRef(d.Name+"Retrieve")),
				"404": notFound,
			},
		}
	}
	if cfg.RouteEnabled(RouteUpdate) {
		item.Put = &openapi.Operation{
			Summary:     "Update " + article(d.NameLower) + " " + d.NameLower,
			Tags:        tags,
			Parameters:  params,
			Security:    security(r, RouteUpdate, cfg, schemes),
			RequestBody: jsonBody(schemaRef(d.Name + "Update")),
			Responses: map[string]openapi.Response{
				"200": jsonResponse("OK", schemaRef(d.Name+"Retrieve")),
				"404": notFound,
				"422": {Description: "Validation error"},
			},
		}
	}
	if cfg.RouteEnabled(RouteDelete) {
		item.Delete = &openapi.Operation{
			Summary:    "Delete " + article(d.NameLower) + " " + d.NameLower,
			Tags:       tags,
			Parameters: params,
			Security:   security(r, RouteDelete, cfg, schemes),
			Responses: map[string]openapi.Response{
				"204": {Description: "No content"},
				"404": notFound,
			},
		}
	}
	if item.Get == nil && item.Put == nil && item.Delete == nil {
		return nil
	}
	return item
}

// actionPaths adds a path entry for every Action carrying a Summary,
// mounted exactly where Register mounts it. An Action without a Summary is
// mounted but left undocumented, which is the documented way to keep an
// internal-only route out of the spec.
func (d ResourceDoc) actionPaths(
	r *BaseResource, cfg ResourceConfig, paths map[string]*openapi.PathItem,
	basePath string, idParam openapi.Parameter, tags []string,
	schemes map[string]openapi.SecurityScheme,
) {
	for _, a := range r.Actions() {
		if a.Summary == "" {
			continue
		}
		p := ActionPath(basePath, a)
		item, ok := paths[p]
		if !ok {
			item = &openapi.PathItem{}
			paths[p] = item
		}
		op := &openapi.Operation{
			Summary:   a.Summary,
			Tags:      tags,
			Responses: a.Responses,
			Security:  securityForAction(r, a, cfg, schemes),
		}
		if a.Detail {
			op.Parameters = []openapi.Parameter{idParam}
		}
		switch a.Method {
		case http.MethodGet:
			item.Get = op
		case http.MethodPost:
			item.Post = op
		case http.MethodPut:
			item.Put = op
		case http.MethodDelete:
			item.Delete = op
		}
	}
}

// ActionPath is where an Action mounts under basePath: the item path when
// Detail, the collection path otherwise, with UrlPath appended if set.
// Generated Register methods and actionPaths both call this, so a route's
// mounted path and its documented path can't drift apart.
func ActionPath(basePath string, a Action) string {
	p := basePath
	if a.Detail {
		p += "/{id}"
	}
	if a.UrlPath != "" {
		p += "/" + a.UrlPath
	}
	return p
}

// security resolves route's Security requirement and merges the schemes it
// refers to into schemes.
func security(r *BaseResource, route Route, cfg ResourceConfig, schemes map[string]openapi.SecurityScheme) []map[string][]string {
	reqs, s := r.SecurityFor(route, cfg)
	for name, scheme := range s {
		schemes[name] = scheme
	}
	return reqs
}

// securityForAction is security for an Action — via SecurityForAction, so
// what's documented is what ProtectAction enforces, including the action's
// own Auth when set.
func securityForAction(r *BaseResource, a Action, cfg ResourceConfig, schemes map[string]openapi.SecurityScheme) []map[string][]string {
	reqs, s := r.SecurityForAction(a, cfg)
	for name, scheme := range s {
		schemes[name] = scheme
	}
	return reqs
}

func schemaRef(name string) string { return "#/components/schemas/" + name }

func jsonResponse(desc, ref string) openapi.Response {
	return openapi.Response{
		Description: desc,
		Content: map[string]openapi.MediaType{
			"application/json": {Schema: openapi.Schema{Ref: ref}},
		},
	}
}

func jsonBody(ref string) *openapi.RequestBody {
	return &openapi.RequestBody{
		Required: true,
		Content: map[string]openapi.MediaType{
			"application/json": {Schema: openapi.Schema{Ref: ref}},
		},
	}
}

// consonantVowelPrefixes are vowel-initial prefixes pronounced with a
// leading consonant sound, so they take "a" rather than "an" — "a user",
// "a unit", "a euro". "User" especially: it's about the most common model
// name there is, and "Create an user" in every one of its doc summaries is
// not a good look.
var consonantVowelPrefixes = []string{"uni", "use", "usu", "uti", "ubi", "eu"}

// article returns the indefinite article for word, so a summary reads
// "Create an author" and "Create a user" rather than "Create a author"
// and "Create an user". Spelling-based with the "u"-sounds-like-"you"
// exceptions above; a word whose article depends on pronunciation alone
// ("an hour", "a URL") still gets it wrong, which isn't worth a
// pronunciation dictionary in a doc summary.
func article(word string) string {
	if word == "" {
		return "a"
	}
	lower := strings.ToLower(word)
	if !strings.ContainsRune("aeiou", rune(lower[0])) {
		return "a"
	}
	for _, p := range consonantVowelPrefixes {
		if strings.HasPrefix(lower, p) {
			return "a"
		}
	}
	return "an"
}
