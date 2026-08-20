package chiadapter_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	chiadapter "github.com/caspel26/goninja/adapters/chi"
	"github.com/caspel26/goninja/openapi"
	"github.com/caspel26/goninja/router"
	"github.com/go-chi/chi/v5"
)

// fakeResource mounts the same five CRUD patterns plus one detail action a
// real generated resource would, recording the id path param it saw.
type fakeResource struct {
	seenID string
}

func (f *fakeResource) Register(mux router.Router) {
	mux.HandleFunc("GET /books", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("list"))
	})
	mux.HandleFunc("POST /books", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	})
	mux.HandleFunc("GET /books/{id}", func(w http.ResponseWriter, r *http.Request) {
		f.seenID = r.PathValue("id")
		_, _ = w.Write([]byte("retrieve:" + f.seenID))
	})
	mux.HandleFunc("PUT /books/{id}", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("update:" + r.PathValue("id")))
	})
	mux.HandleFunc("DELETE /books/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("POST /books/{id}/publish", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("publish:" + r.PathValue("id")))
	})
}

func (f *fakeResource) OpenAPI() (map[string]*openapi.PathItem, map[string]openapi.Schema, map[string]openapi.SecurityScheme) {
	return nil, nil, nil
}

func TestAdapter_CRUDAndDetailAction(t *testing.T) {
	r := chi.NewRouter()
	res := &fakeResource{}
	res.Register(chiadapter.New(r))

	srv := httptest.NewServer(r)
	defer srv.Close()

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantBody   string
	}{
		{"list", http.MethodGet, "/books", http.StatusOK, "list"},
		{"create", http.MethodPost, "/books", http.StatusCreated, "created"},
		{"retrieve", http.MethodGet, "/books/42", http.StatusOK, "retrieve:42"},
		{"update", http.MethodPut, "/books/42", http.StatusOK, "update:42"},
		{"delete", http.MethodDelete, "/books/42", http.StatusNoContent, ""},
		{"detail action", http.MethodPost, "/books/42/publish", http.StatusOK, "publish:42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, srv.URL+tt.path, nil)
			if err != nil {
				t.Fatalf("NewRequest: %v", err)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Do: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
			if tt.wantBody == "" {
				return
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("ReadAll: %v", err)
			}
			if string(body) != tt.wantBody {
				t.Fatalf("body = %q, want %q", body, tt.wantBody)
			}
		})
	}
}

func TestAdapter_MountsUnderSubRouter(t *testing.T) {
	r := chi.NewRouter()
	res := &fakeResource{}
	r.Route("/api/v1", func(sub chi.Router) {
		res.Register(chiadapter.New(sub))
	})

	srv := httptest.NewServer(r)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/books/7")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "retrieve:7" {
		t.Fatalf("body = %q, want %q", body, "retrieve:7")
	}
}

type protectedResource struct{}

func (protectedResource) Register(mux router.Router) {
	mux.HandleFunc("GET /secret", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":"UNAUTHORIZED","error":"unauthorized"}`))
	})
}

func (protectedResource) OpenAPI() (map[string]*openapi.PathItem, map[string]openapi.Schema, map[string]openapi.SecurityScheme) {
	return nil, nil, nil
}

func TestAdapter_ProtectedRouteReturns401(t *testing.T) {
	r := chi.NewRouter()
	protectedResource{}.Register(chiadapter.New(r))

	srv := httptest.NewServer(r)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/secret")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}
