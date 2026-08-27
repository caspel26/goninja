package goninja

import (
	"net/http"
	"testing"

	"github.com/caspel26/goninja/openapi"
)

// docFixture is the minimal ResourceDoc a generated OpenAPI() would pass:
// two schemas and one filter param, standing in for a real model's shape.
func docFixture() ResourceDoc {
	return ResourceDoc{
		Name:      "Author",
		NameLower: "author",
		IDSchema:  openapi.Schema{Type: "string"},
		Schemas: map[string]openapi.Schema{
			"AuthorList":     {Type: "object"},
			"AuthorRetrieve": {Type: "object"},
			"AuthorCreate":   {Type: "object"},
			"AuthorUpdate":   {Type: "object"},
		},
		ListParams: []openapi.Parameter{
			{Name: "name", In: "query", Schema: openapi.Schema{Type: "string"}},
		},
	}
}

func TestBuildResourceOpenAPI_PathsAndSchemas(t *testing.T) {
	var r BaseResource
	paths, schemas, schemes := BuildResourceOpenAPI(&r, ResourceConfig{}, docFixture())

	if _, ok := paths["/authors"]; !ok {
		t.Errorf("paths = %v, want a /authors entry", keysOf(paths))
	}
	if _, ok := paths["/authors/{id}"]; !ok {
		t.Errorf("paths = %v, want an /authors/{id} entry", keysOf(paths))
	}

	// The envelope is added by the helper, not carried in doc.Schemas.
	env, ok := schemas["AuthorListEnvelope"]
	if !ok {
		t.Fatalf("schemas missing AuthorListEnvelope; got %v", keysOfSchemas(schemas))
	}
	items := env.Properties["items"]
	if items.Items == nil || items.Items.Ref != "#/components/schemas/AuthorList" {
		t.Errorf("envelope items = %+v, want a $ref to AuthorList", items)
	}
	// doc's own schemas survive alongside it.
	if _, ok := schemas["AuthorRetrieve"]; !ok {
		t.Error("schemas dropped the doc's own AuthorRetrieve")
	}
	// Nothing is protected on a zero Config, so no schemes are collected.
	if len(schemes) != 0 {
		t.Errorf("schemes = %v, want empty for an unprotected resource", schemes)
	}
}

func TestBuildResourceOpenAPI_AppendsSharedListParams(t *testing.T) {
	var r BaseResource
	paths, _, _ := BuildResourceOpenAPI(&r, ResourceConfig{}, docFixture())

	var names []string
	for _, p := range paths["/authors"].Get.Parameters {
		names = append(names, p.Name)
	}
	want := []string{"name", "limit", "offset", "order"}
	if len(names) != len(want) {
		t.Fatalf("list params = %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Errorf("list params = %v, want %v (doc's own first, then the shared three)", names, want)
		}
	}
}

func TestBuildResourceOpenAPI_UsesConfiguredTagsAndPath(t *testing.T) {
	var r BaseResource
	r.SetOpenAPITags("Writers")
	paths, _, _ := BuildResourceOpenAPI(&r, ResourceConfig{Path: "/v1/authors"}, docFixture())

	item, ok := paths["/v1/authors"]
	if !ok {
		t.Fatalf("paths = %v, want the ResourceConfig.Path override honored", keysOf(paths))
	}
	if got := item.Get.Tags; len(got) != 1 || got[0] != "Writers" {
		t.Errorf("tags = %v, want the SetOpenAPITags override", got)
	}
	if _, ok := paths["/v1/authors/{id}"]; !ok {
		t.Errorf("paths = %v, want the item path under the overridden base", keysOf(paths))
	}
}

func TestBuildResourceOpenAPI_RestrictedRoutes(t *testing.T) {
	var r BaseResource

	t.Run("only list leaves no item path at all", func(t *testing.T) {
		paths, _, _ := BuildResourceOpenAPI(&r, ResourceConfig{Routes: []Route{RouteList}}, docFixture())
		if _, ok := paths["/authors/{id}"]; ok {
			t.Error("documented an /authors/{id} path with every item route disabled")
		}
		base := paths["/authors"]
		if base.Get == nil || base.Post != nil {
			t.Errorf("base path = %+v, want GET only", base)
		}
	})

	t.Run("only retrieve leaves no collection path at all", func(t *testing.T) {
		paths, _, _ := BuildResourceOpenAPI(&r, ResourceConfig{Routes: []Route{RouteRetrieve}}, docFixture())
		if _, ok := paths["/authors"]; ok {
			t.Error("documented an /authors path with list and create both disabled")
		}
		item := paths["/authors/{id}"]
		if item.Get == nil || item.Put != nil || item.Delete != nil {
			t.Errorf("item path = %+v, want GET only", item)
		}
	})
}

// TestBuildResourceOpenAPI_DocumentsActions is the behavioral coverage for
// documenting Actions that used to live in generated code (see
// internal/codegen's TestGenerate_ActionsDispatch).
func TestBuildResourceOpenAPI_DocumentsActions(t *testing.T) {
	var r BaseResource
	r.SetActions(
		Action{Name: "publish", Detail: true, Method: http.MethodPost, UrlPath: "publish", Summary: "Publish an author"},
		Action{Name: "sync", Method: http.MethodPost, UrlPath: "sync", Summary: "Sync authors"},
		Action{Name: "internal", Method: http.MethodGet, UrlPath: "internal"}, // no Summary -> undocumented
	)

	paths, _, _ := BuildResourceOpenAPI(&r, ResourceConfig{}, docFixture())

	detail, ok := paths["/authors/{id}/publish"]
	if !ok {
		t.Fatalf("paths = %v, want a detail action path", keysOf(paths))
	}
	if detail.Post == nil || detail.Post.Summary != "Publish an author" {
		t.Errorf("detail action op = %+v, want the action's Summary", detail.Post)
	}
	if len(detail.Post.Parameters) != 1 || detail.Post.Parameters[0].Name != "id" {
		t.Errorf("detail action params = %+v, want the id path param", detail.Post.Parameters)
	}

	collection, ok := paths["/authors/sync"]
	if !ok {
		t.Fatalf("paths = %v, want a collection action path", keysOf(paths))
	}
	if collection.Post == nil || len(collection.Post.Parameters) != 0 {
		t.Errorf("collection action = %+v, want no id param", collection.Post)
	}

	if _, ok := paths["/authors/internal"]; ok {
		t.Error("documented an Action with no Summary, want it mounted but undocumented")
	}
}

func TestBuildResourceOpenAPI_ActionMethods(t *testing.T) {
	var r BaseResource
	r.SetActions(
		Action{Name: "g", Method: http.MethodGet, UrlPath: "x", Summary: "get"},
		Action{Name: "p", Method: http.MethodPut, UrlPath: "y", Summary: "put"},
		Action{Name: "d", Method: http.MethodDelete, UrlPath: "z", Summary: "delete"},
		Action{Name: "weird", Method: "PATCH", UrlPath: "w", Summary: "patch"},
	)
	paths, _, _ := BuildResourceOpenAPI(&r, ResourceConfig{}, docFixture())

	if paths["/authors/x"].Get == nil {
		t.Error("GET action not documented under Get")
	}
	if paths["/authors/y"].Put == nil {
		t.Error("PUT action not documented under Put")
	}
	if paths["/authors/z"].Delete == nil {
		t.Error("DELETE action not documented under Delete")
	}
	// An unsupported method still creates the path entry but sets no
	// operation — surfacing it as an empty PathItem rather than silently
	// attaching it to the wrong verb.
	w := paths["/authors/w"]
	if w == nil || w.Get != nil || w.Post != nil || w.Put != nil || w.Delete != nil {
		t.Errorf("PATCH action = %+v, want an entry with no operation set", w)
	}
}

func TestBuildResourceOpenAPI_CollectsSecuritySchemes(t *testing.T) {
	auth := &fakeAuthenticator{name: "bearer", allow: true}
	var r BaseResource
	r.SetConfig(Config{DefaultAuth: AuthPolicy{Routes: []Route{RouteCreate}, Auth: []Authenticator{auth}}})
	r.SetActions(Action{
		Name: "publish", Method: http.MethodPost, UrlPath: "publish", Summary: "Publish",
		Auth: &RouteAuth{Auth: []Authenticator{&fakeAuthenticator{name: "adminKey", allow: true}}},
	})

	paths, _, schemes := BuildResourceOpenAPI(&r, ResourceConfig{}, docFixture())

	if _, ok := schemes["bearer"]; !ok {
		t.Errorf("schemes = %v, want the CRUD route's scheme", keysOfSchemes(schemes))
	}
	if _, ok := schemes["adminKey"]; !ok {
		t.Errorf("schemes = %v, want the Action.Auth scheme too", keysOfSchemes(schemes))
	}
	if reqs := paths["/authors"].Post.Security; len(reqs) != 1 {
		t.Errorf("create Security = %v, want one requirement", reqs)
	}
	if reqs := paths["/authors"].Get.Security; reqs != nil {
		t.Errorf("list Security = %v, want nil for an unprotected route", reqs)
	}
}

func TestArticle(t *testing.T) {
	tests := map[string]string{
		// Plain vowel / consonant starts.
		"author": "an", "agent": "an", "item": "an", "order": "an",
		"book": "a", "ticket": "a", "": "a", "Author": "an",
		// "u" pronounced "you" takes "a"...
		"user": "a", "unit": "a", "utility": "a", "euro": "a", "User": "a",
		// ...but a genuinely vowel-sounding "u" still takes "an".
		"upload": "an", "update": "an",
	}
	for word, want := range tests {
		if got := article(word); got != want {
			t.Errorf("article(%q) = %q, want %q", word, got, want)
		}
	}
}

func TestActionPath(t *testing.T) {
	tests := []struct {
		name string
		a    Action
		want string
	}{
		{"detail with path", Action{Detail: true, UrlPath: "publish"}, "/books/{id}/publish"},
		{"detail without path", Action{Detail: true}, "/books/{id}"},
		{"collection with path", Action{UrlPath: "bulk"}, "/books/bulk"},
		{"collection without path", Action{}, "/books"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ActionPath("/books", tt.a); got != tt.want {
				t.Errorf("ActionPath = %q, want %q", got, tt.want)
			}
		})
	}
}

func keysOf(m map[string]*openapi.PathItem) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func keysOfSchemas(m map[string]openapi.Schema) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func keysOfSchemes(m map[string]openapi.SecurityScheme) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
