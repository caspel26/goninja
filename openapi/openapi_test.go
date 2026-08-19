package openapi

import (
	"net/http"
	"testing"
)

type fakeResource struct {
	registered bool
	path       string
	excluded   bool
}

func (f *fakeResource) Register(mux *http.ServeMux) {
	f.registered = true
	mux.HandleFunc("GET "+f.path, func(w http.ResponseWriter, r *http.Request) {})
}

func (f *fakeResource) OpenAPI() (map[string]*PathItem, map[string]Schema) {
	return map[string]*PathItem{
			f.path: {Get: &Operation{Summary: "list", Responses: map[string]Response{}}},
		}, map[string]Schema{
			"Fake": {Type: "object"},
		}
}

func (f *fakeResource) DocsExcluded() bool { return f.excluded }

func TestAPI_AddAndSpec(t *testing.T) {
	doc := NewAPI("Test API", "1.0.0")
	r := &fakeResource{path: "/fakes"}
	doc.Add(r)

	spec := doc.Spec()
	if spec.OpenAPI != "3.0.3" {
		t.Errorf("OpenAPI version = %q, want 3.0.3", spec.OpenAPI)
	}
	if spec.Info.Title != "Test API" || spec.Info.Version != "1.0.0" {
		t.Errorf("Info = %+v, want Title=Test API Version=1.0.0", spec.Info)
	}
	if _, ok := spec.Paths["/fakes"]; !ok {
		t.Error(`Spec().Paths missing "/fakes" after Add`)
	}
	if _, ok := spec.Components.Schemas["Fake"]; !ok {
		t.Error(`Spec().Components.Schemas missing "Fake" after Add`)
	}
}

func TestMount_RegistersAndDocuments(t *testing.T) {
	mux := http.NewServeMux()
	doc := NewAPI("Test API", "1.0.0")
	r := &fakeResource{path: "/fakes"}

	Mount(mux, doc, r)

	if !r.registered {
		t.Error("Mount did not call Register on the resource")
	}
	if _, ok := doc.Spec().Paths["/fakes"]; !ok {
		t.Error("Mount did not add the resource's OpenAPI fragment to doc")
	}
}

func TestMount_NilDocSkipsOpenAPI(t *testing.T) {
	mux := http.NewServeMux()
	r := &fakeResource{path: "/fakes"}

	Mount(mux, nil, r) // must not panic

	if !r.registered {
		t.Error("Mount(mux, nil, ...) should still call Register")
	}
}

func TestMount_ExcludedFromDocs(t *testing.T) {
	mux := http.NewServeMux()
	doc := NewAPI("Test API", "1.0.0")
	r := &fakeResource{path: "/fakes", excluded: true}

	Mount(mux, doc, r)

	if !r.registered {
		t.Error("Mount should still register an excluded resource's routes")
	}
	if _, ok := doc.Spec().Paths["/fakes"]; ok {
		t.Error("Mount added a DocsExcluded resource's fragment to doc, want it skipped")
	}
}
