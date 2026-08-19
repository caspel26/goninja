package goninja

import (
	"net/http"

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
