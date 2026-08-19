package openapi

import "testing"

type fakeProvider struct {
	path string
}

func (f *fakeProvider) OpenAPI() (map[string]*PathItem, map[string]Schema) {
	return map[string]*PathItem{
			f.path: {Get: &Operation{Summary: "list", Responses: map[string]Response{}}},
		}, map[string]Schema{
			"Fake": {Type: "object"},
		}
}

func TestSpec_MarshalsExpectedShape(t *testing.T) {
	p := &fakeProvider{path: "/fakes"}
	paths, schemas := p.OpenAPI()

	spec := Spec{
		OpenAPI:    "3.0.3",
		Info:       Info{Title: "Test API", Version: "1.0.0"},
		Paths:      paths,
		Components: Components{Schemas: schemas},
	}

	if spec.OpenAPI != "3.0.3" {
		t.Errorf("OpenAPI version = %q, want 3.0.3", spec.OpenAPI)
	}
	if _, ok := spec.Paths["/fakes"]; !ok {
		t.Error(`Spec.Paths missing "/fakes"`)
	}
	if _, ok := spec.Components.Schemas["Fake"]; !ok {
		t.Error(`Spec.Components.Schemas missing "Fake"`)
	}
}
