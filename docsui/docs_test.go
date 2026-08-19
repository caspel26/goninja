package docsui

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/caspel26/goninja/openapi"
)

type fakeSpecSource struct{}

func (fakeSpecSource) Spec() openapi.Spec {
	return openapi.Spec{OpenAPI: "3.0.3", Info: openapi.Info{Title: "Test API", Version: "1.0.0"}}
}

type fakeDocsUI struct{}

func (fakeDocsUI) Index(specPath string) []byte {
	return []byte("<html>" + specPath + "</html>")
}

func (fakeDocsUI) Assets() map[string]DocsAsset {
	return map[string]DocsAsset{
		"style.css": {Data: []byte("body{}"), ContentType: "text/css"},
	}
}

func TestMountDocs(t *testing.T) {
	mux := http.NewServeMux()
	doc := fakeSpecSource{}
	MountDocs(mux, doc, "/docs", fakeDocsUI{})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	t.Run("spec is served as JSON", func(t *testing.T) {
		resp, err := http.Get(srv.URL + "/docs/openapi.json")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("index is served at path with trailing slash", func(t *testing.T) {
		resp, err := http.Get(srv.URL + "/docs/")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("bare path redirects to trailing slash", func(t *testing.T) {
		client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}
		resp, err := client.Get(srv.URL + "/docs")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusMovedPermanently {
			t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusMovedPermanently)
		}
		if loc := resp.Header.Get("Location"); loc != "/docs/" {
			t.Errorf("Location = %q, want /docs/", loc)
		}
	})

	t.Run("assets are served under path", func(t *testing.T) {
		resp, err := http.Get(srv.URL + "/docs/style.css")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
		if ct := resp.Header.Get("Content-Type"); ct != "text/css" {
			t.Errorf("Content-Type = %q, want text/css", ct)
		}
	})
}

func TestMountDocs_NilUIDefaultsToSwaggerUI(t *testing.T) {
	mux := http.NewServeMux()
	doc := fakeSpecSource{}
	MountDocs(mux, doc, "docs", nil) // no leading slash, and nil UI — must not panic

	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/docs/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}
