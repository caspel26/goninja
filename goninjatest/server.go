package goninjatest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/caspel26/goninja"
)

// NewServer mounts resources on a fresh *http.ServeMux (via Register(mux),
// the same call goninja.API.Mount makes) and serves it from a real
// httptest.Server — no OpenAPI merging, docs, or Config/auth wiring, since
// a resource test exercises its routes directly rather than the mounted
// app. The server is closed automatically via t.Cleanup.
func NewServer(t testing.TB, resources ...goninja.Resource) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	for _, r := range resources {
		r.Register(mux)
	}

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}
