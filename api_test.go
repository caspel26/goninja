package goninja

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/caspel26/goninja/openapi"
)

func TestAPI_MountRegistersAndDocuments(t *testing.T) {
	mux := http.NewServeMux()
	api := NewAPI("Test API", "1.0.0")
	r := &fakeResource{path: "/fakes"}

	api.Mount(mux, r)

	if !r.registered {
		t.Error("Mount did not call Register on the resource")
	}
	spec := api.Spec()
	if spec.OpenAPI != "3.0.3" {
		t.Errorf("OpenAPI version = %q, want 3.0.3", spec.OpenAPI)
	}
	if _, ok := spec.Paths["/fakes"]; !ok {
		t.Error("Mount did not add the resource's OpenAPI fragment to the API")
	}
}

func TestAPI_MountExcludedFromDocs(t *testing.T) {
	mux := http.NewServeMux()
	api := NewAPI("Test API", "1.0.0")
	r := &fakeResource{path: "/fakes", excluded: true}

	api.Mount(mux, r)

	if !r.registered {
		t.Error("Mount should still register an excluded resource's routes")
	}
	if _, ok := api.Spec().Paths["/fakes"]; ok {
		t.Error("Mount added a DocsExcluded resource's fragment, want it skipped")
	}
}

func TestAPI_RegisterDirectlySkipsDocs(t *testing.T) {
	mux := http.NewServeMux()
	api := NewAPI("Test API", "1.0.0")
	r := &fakeResource{path: "/fakes"}

	r.Register(mux) // bypass Mount entirely — routes mount, nothing documented

	if !r.registered {
		t.Error("expected Register to run")
	}
	if _, ok := api.Spec().Paths["/fakes"]; ok {
		t.Error("expected the API's spec to stay empty when Mount is bypassed")
	}
}

func TestAPI_MountWithConfig_SetsConfigAndRegisters(t *testing.T) {
	mux := http.NewServeMux()
	api := NewAPI("Test API", "1.0.0")
	r := &fakeResource{path: "/fakes"}
	cfg := Config{DefaultAuth: AuthPolicy{Routes: []Route{RouteCreate}}}

	api.MountWithConfig(mux, cfg, r)

	if !r.registered {
		t.Error("MountWithConfig did not call Register on the resource")
	}
	got := r.Config()
	if len(got.DefaultAuth.Routes) != 1 || got.DefaultAuth.Routes[0] != RouteCreate {
		t.Errorf("resource Config() = %+v, want the cfg passed to MountWithConfig", got)
	}
	if _, ok := api.Spec().Paths["/fakes"]; !ok {
		t.Error("MountWithConfig did not add the resource's OpenAPI fragment to the API")
	}
}

func TestAPI_MountWithConfig_ExcludedFromDocs(t *testing.T) {
	mux := http.NewServeMux()
	api := NewAPI("Test API", "1.0.0")
	r := &fakeResource{path: "/fakes", excluded: true}
	cfg := Config{DefaultAuth: AuthPolicy{Routes: []Route{RouteCreate}}}

	api.MountWithConfig(mux, cfg, r)

	if !r.registered {
		t.Error("MountWithConfig should still register an excluded resource's routes")
	}
	if _, ok := api.Spec().Paths["/fakes"]; ok {
		t.Error("MountWithConfig added a DocsExcluded resource's fragment, want it skipped")
	}
}

// markerMapping builds an ErrorMapping that matches every error (via
// NewErrorMapping[error], since errors.As into a *error target always
// succeeds) and reports status so tests can tell which mapping answered
// without comparing ComposedErrorMapper values directly — it holds a
// []ErrorMapping of funcs, which isn't comparable with ==.
func markerMapping(status int) ErrorMapping {
	return NewErrorMapping(func(err error) (int, any) { return status, nil })
}

func TestAPI_Mount_AppliesSetErrorMapper(t *testing.T) {
	mux := http.NewServeMux()
	api := NewAPI("Test API", "1.0.0")
	api.SetErrorMapper(markerMapping(599))
	r := &fakeResource{path: "/fakes"}

	api.Mount(mux, r)

	status, _ := r.Config().DefaultErrorMapper.MapError(errors.New("boom"))
	if status != 599 {
		t.Errorf("status = %d, want 599 (the mapping passed to API.SetErrorMapper)", status)
	}
}

func TestAPI_Mount_NoConfigCallWithoutSetErrorMapper(t *testing.T) {
	mux := http.NewServeMux()
	api := NewAPI("Test API", "1.0.0")
	r := &fakeResource{path: "/fakes", cfg: Config{DefaultAuth: AuthPolicy{Routes: []Route{RouteCreate}}}}

	api.Mount(mux, r)

	if got := r.Config(); len(got.DefaultAuth.Routes) == 0 {
		t.Error("Mount overwrote the resource's existing Config even though API.SetErrorMapper was never called")
	}
}

func TestAPI_MountWithConfig_FallsBackToAPIErrorMapper(t *testing.T) {
	mux := http.NewServeMux()
	api := NewAPI("Test API", "1.0.0")
	api.SetErrorMapper(markerMapping(599))
	r := &fakeResource{path: "/fakes"}

	api.MountWithConfig(mux, Config{DefaultAuth: AuthPolicy{Routes: []Route{RouteCreate}}}, r)

	status, _ := r.Config().DefaultErrorMapper.MapError(errors.New("boom"))
	if status != 599 {
		t.Errorf("status = %d, want 599 (the API's SetErrorMapper value)", status)
	}
}

func TestAPI_MountWithConfig_ExplicitDefaultErrorMapperWins(t *testing.T) {
	mux := http.NewServeMux()
	api := NewAPI("Test API", "1.0.0")
	api.SetErrorMapper(markerMapping(599))
	explicit := NewErrorMapper(markerMapping(600))
	r := &fakeResource{path: "/fakes"}

	api.MountWithConfig(mux, Config{DefaultErrorMapper: explicit}, r)

	status, _ := r.Config().DefaultErrorMapper.MapError(errors.New("boom"))
	if status != 600 {
		t.Errorf("status = %d, want 600 (the explicit cfg value)", status)
	}
}

func TestAPI_Add_MergesSecuritySchemes(t *testing.T) {
	api := NewAPI("Test API", "1.0.0")
	r := &fakeResource{
		path: "/fakes",
		securitySchemes: map[string]openapi.SecurityScheme{
			"bearer": {Type: "http", Scheme: "bearer"},
		},
	}

	api.Add(r)

	got := api.Spec().Components.SecuritySchemes
	if _, ok := got["bearer"]; !ok {
		t.Errorf("SecuritySchemes = %+v, want a merged \"bearer\" entry", got)
	}
}

func TestAPI_MountDocs(t *testing.T) {
	mux := http.NewServeMux()
	api := NewAPI("Test API", "1.0.0")
	r := &fakeResource{path: "/fakes"}
	api.Mount(mux, r)
	api.MountDocs(mux, "/docs", nil)

	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/docs/openapi.json")
	if err != nil {
		t.Fatalf("GET /docs/openapi.json: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}
