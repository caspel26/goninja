package chiadapter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	chiadapter "github.com/caspel26/goninja/adapters/chi"
	"github.com/caspel26/goninja/docsui"
	"github.com/caspel26/goninja/openapi"
	"github.com/go-chi/chi/v5"
)

type fakeSpecSource struct{}

func (fakeSpecSource) Spec() openapi.Spec {
	return openapi.Spec{OpenAPI: "3.0.3", Info: openapi.Info{Title: "Test", Version: "1.0.0"}}
}

func TestAdapter_MountDocsServesSpecAndIndex(t *testing.T) {
	r := chi.NewRouter()
	docsui.MountDocs(chiadapter.New(r), fakeSpecSource{}, "/docs", docsui.SwaggerUI())

	srv := httptest.NewServer(r)
	defer srv.Close()

	for _, path := range []string{"/docs/openapi.json", "/docs/"} {
		resp, err := http.Get(srv.URL + path)
		if err != nil {
			t.Fatalf("Get(%s): %v", path, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Get(%s) status = %d, want %d", path, resp.StatusCode, http.StatusOK)
		}
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Get(srv.URL + "/docs")
	if err != nil {
		t.Fatalf("Get(/docs): %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusMovedPermanently {
		t.Fatalf("Get(/docs) status = %d, want %d (redirect to /docs/)", resp.StatusCode, http.StatusMovedPermanently)
	}
}
