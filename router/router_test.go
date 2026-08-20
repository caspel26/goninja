package router

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestParsePattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		want    Pattern
		wantErr bool
	}{
		{
			name:    "method and single param",
			pattern: "GET /books/{id}",
			want:    Pattern{Method: "GET", Path: "/books/{id}", Params: []string{"id"}},
		},
		{
			name:    "no method",
			pattern: "/books",
			want:    Pattern{Method: "", Path: "/books"},
		},
		{
			name:    "no params",
			pattern: "GET /books",
			want:    Pattern{Method: "GET", Path: "/books"},
		},
		{
			name:    "multiple params",
			pattern: "GET /authors/{authorID}/books/{id}",
			want:    Pattern{Method: "GET", Path: "/authors/{authorID}/books/{id}", Params: []string{"authorID", "id"}},
		},
		{
			name:    "trailing slash is a subtree match",
			pattern: "GET /docs/",
			want:    Pattern{Method: "GET", Path: "/docs/", Subtree: true},
		},
		{
			name:    "trailing wildcard",
			pattern: "GET /docs/{rest...}",
			want:    Pattern{Method: "GET", Path: "/docs/{rest...}", Wildcard: "rest"},
		},
		{
			name:    "host-prefixed pattern is rejected",
			pattern: "GET example.com/books",
			wantErr: true,
		},
		{
			name:    "empty path is rejected",
			pattern: "GET ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePattern(tt.pattern)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParsePattern(%q) = %+v, nil; want error", tt.pattern, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParsePattern(%q) unexpected error: %v", tt.pattern, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ParsePattern(%q) = %+v, want %+v", tt.pattern, got, tt.want)
			}
		})
	}
}

func TestPattern_TranslatePath(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		style   ParamStyle
		want    string
	}{
		{"brace is identity", "/books/{id}", StyleBrace, "/books/{id}"},
		{"single param to colon", "/books/{id}", StyleColon, "/books/:id"},
		{"multiple params to colon", "/authors/{authorID}/books/{id}", StyleColon, "/authors/:authorID/books/:id"},
		{"wildcard to star", "/docs/{rest...}", StyleColon, "/docs/*rest"},
		{"no params unaffected", "/books", StyleColon, "/books"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := ParsePattern("GET " + tt.pattern)
			if err != nil {
				t.Fatalf("ParsePattern: %v", err)
			}
			if got := p.TranslatePath(tt.style); got != tt.want {
				t.Fatalf("TranslatePath(%v) = %q, want %q", tt.style, got, tt.want)
			}
		})
	}
}

func TestBindPathValues(t *testing.T) {
	req := httptest.NewRequest("GET", "/books/42", nil)
	values := map[string]string{"id": "42", "authorID": "7"}

	BindPathValues(req, []string{"id", "authorID"}, func(name string) string {
		return values[name]
	})

	if got := req.PathValue("id"); got != "42" {
		t.Fatalf("PathValue(id) = %q, want %q", got, "42")
	}
	if got := req.PathValue("authorID"); got != "7" {
		t.Fatalf("PathValue(authorID) = %q, want %q", got, "7")
	}
}

func TestBindPathValues_OverwritesExistingServeMuxMatch(t *testing.T) {
	mux := http.NewServeMux()
	var captured *http.Request
	mux.HandleFunc("GET /books/{id}", func(w http.ResponseWriter, r *http.Request) {
		captured = r
	})

	req := httptest.NewRequest("GET", "/books/1", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if captured == nil {
		t.Fatal("handler was not called")
	}
	if got := captured.PathValue("id"); got != "1" {
		t.Fatalf("PathValue(id) before rebind = %q, want %q", got, "1")
	}

	BindPathValues(captured, []string{"id"}, func(string) string { return "99" })
	if got := captured.PathValue("id"); got != "99" {
		t.Fatalf("PathValue(id) after rebind = %q, want %q", got, "99")
	}
}
