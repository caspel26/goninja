package goninjatest_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/caspel26/goninja/goninjatest"
	"github.com/caspel26/goninja/openapi"
)

type fakeResource struct {
	path string
}

func (r *fakeResource) Register(mux *http.ServeMux) {
	mux.HandleFunc(r.path, func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
}

func (r *fakeResource) OpenAPI() (map[string]*openapi.PathItem, map[string]openapi.Schema) {
	return nil, nil
}

func TestNewServer_MountsAndServesResources(t *testing.T) {
	srv := goninjatest.NewServer(t, &fakeResource{path: "/widgets"})

	resp, err := http.Get(srv.URL + "/widgets")
	if err != nil {
		t.Fatalf("GET /widgets: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "ok" {
		t.Errorf("body = %q, want %q", body, "ok")
	}
}

func TestNewServer_MountsMultipleResources(t *testing.T) {
	srv := goninjatest.NewServer(t, &fakeResource{path: "/widgets"}, &fakeResource{path: "/gadgets"})

	for _, path := range []string{"/widgets", "/gadgets"} {
		resp, err := http.Get(srv.URL + path)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s status = %d, want 200", path, resp.StatusCode)
		}
	}
}
