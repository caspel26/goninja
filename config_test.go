package goninja

import (
	"net/http"

	"github.com/caspel26/goninja/openapi"
)

type fakeResource struct {
	registered      bool
	path            string
	cfg             Config
	excluded        bool
	securitySchemes map[string]openapi.SecurityScheme
}

func (f *fakeResource) SetConfig(cfg Config) { f.cfg = cfg }
func (f *fakeResource) Config() Config       { return f.cfg }

func (f *fakeResource) Register(mux Router) {
	f.registered = true
	mux.HandleFunc("GET "+f.path, func(w http.ResponseWriter, r *http.Request) {})
}

func (f *fakeResource) OpenAPI() (map[string]*openapi.PathItem, map[string]openapi.Schema, map[string]openapi.SecurityScheme) {
	return map[string]*openapi.PathItem{
			f.path: {Get: &openapi.Operation{Summary: "list", Responses: map[string]openapi.Response{}}},
		}, map[string]openapi.Schema{
			"Fake": {Type: "object"},
		}, f.securitySchemes
}

func (f *fakeResource) DocsExcluded() bool { return f.excluded }
