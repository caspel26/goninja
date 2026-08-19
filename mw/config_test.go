package mw

import (
	"net/http"
	"testing"

	"github.com/caspel26/goninja/openapi"
)

type fakeResource struct {
	registered bool
	path       string
	cfg        Config
	excluded   bool
}

func (f *fakeResource) SetConfig(cfg Config) { f.cfg = cfg }
func (f *fakeResource) Config() Config       { return f.cfg }

func (f *fakeResource) Register(mux *http.ServeMux) {
	f.registered = true
	mux.HandleFunc("GET "+f.path, func(w http.ResponseWriter, r *http.Request) {})
}

func (f *fakeResource) OpenAPI() (map[string]*openapi.PathItem, map[string]openapi.Schema) {
	return map[string]*openapi.PathItem{
			f.path: {Get: &openapi.Operation{Summary: "list", Responses: map[string]openapi.Response{}}},
		}, map[string]openapi.Schema{
			"Fake": {Type: "object"},
		}
}

func (f *fakeResource) DocsExcluded() bool { return f.excluded }

func TestMountWithConfig_SetsConfigAndRegisters(t *testing.T) {
	mux := http.NewServeMux()
	doc := openapi.NewAPI("Test API", "1.0.0")
	r := &fakeResource{path: "/fakes"}
	cfg := Config{DefaultAuth: AuthPolicy{Protected: []string{"create"}}}

	MountWithConfig(mux, doc, cfg, r)

	if !r.registered {
		t.Error("MountWithConfig did not call Register on the resource")
	}
	got := r.Config()
	if len(got.DefaultAuth.Protected) != 1 || got.DefaultAuth.Protected[0] != "create" {
		t.Errorf("resource Config() = %+v, want the cfg passed to MountWithConfig", got)
	}
	if _, ok := doc.Spec().Paths["/fakes"]; !ok {
		t.Error("MountWithConfig did not add the resource's OpenAPI fragment to doc")
	}
}

func TestMountWithConfig_NilDocAndExclusion(t *testing.T) {
	mux := http.NewServeMux()
	r := &fakeResource{path: "/fakes", excluded: true}

	MountWithConfig(mux, nil, Config{}, r) // must not panic

	doc := openapi.NewAPI("Test API", "1.0.0")
	r2 := &fakeResource{path: "/others", excluded: true}
	MountWithConfig(mux, doc, Config{}, r2)

	if _, ok := doc.Spec().Paths["/others"]; ok {
		t.Error("MountWithConfig added a DocsExcluded resource's fragment to doc, want it skipped")
	}
}
