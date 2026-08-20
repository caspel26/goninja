package main

// Pins the exact stdlib-style route patterns a generated Register method
// emits ("GET /books/{id}", ...) — this is the contract every router
// adapter (gin/echo/chi, under adapters/) translates from, so it's covered
// here once rather than re-derived per adapter.

import (
	"net/http"
	"reflect"
	"sort"
	"testing"

	"github.com/caspel26/goninja/examples/prototype/internal/api"
	"github.com/caspel26/goninja/examples/prototype/models"
	"github.com/caspel26/goninja/goninjatest"
)

type recordingRouter struct {
	patterns []string
}

func (r *recordingRouter) HandleFunc(pattern string, _ func(http.ResponseWriter, *http.Request)) {
	r.patterns = append(r.patterns, pattern)
}

func TestGeneratedRegister_EmitsStablePatterns(t *testing.T) {
	db := goninjatest.NewDB(t, &models.Author{}, &models.Book{})
	r := &recordingRouter{}

	api.NewBookResource(db).Register(r)

	want := []string{
		"GET /books",
		"POST /books",
		"GET /books/{id}",
		"PUT /books/{id}",
		"DELETE /books/{id}",
	}
	sort.Strings(want)
	got := append([]string(nil), r.patterns...)
	sort.Strings(got)

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Register patterns = %v, want %v", got, want)
	}
}
