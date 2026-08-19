package goninja

import (
	"net/http"
	"testing"
)

func TestMountWithConfig_SetsConfigAndRegisters(t *testing.T) {
	mux := http.NewServeMux()
	doc := NewAPI("Test API", "1.0.0")
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
	r := &fakeResource{path: "/fakes"}
	r.ExcludeFromDocs()

	MountWithConfig(mux, nil, Config{}, r) // must not panic

	doc := NewAPI("Test API", "1.0.0")
	r2 := &fakeResource{path: "/others"}
	r2.ExcludeFromDocs()
	MountWithConfig(mux, doc, Config{}, r2)

	if _, ok := doc.Spec().Paths["/others"]; ok {
		t.Error("MountWithConfig added an ExcludeFromDocs resource's fragment to doc, want it skipped")
	}
}
