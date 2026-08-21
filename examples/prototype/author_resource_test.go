package main

// Demonstrates testing a generated resource
// end to end in under 10 lines using goninjatest, against a real
// generated resource (api.AuthorResource) rather than a hand-written
// fake. No Postgres needed: goninjatest.NewDB is a plain GORM connection,
// and the generated resource code has no Postgres-specific behavior.

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/caspel26/goninja/examples/prototype/internal/api"
	"github.com/caspel26/goninja/examples/prototype/models"
	"github.com/caspel26/goninja/goninjatest"
)

func TestAuthorResource_CreateAndList(t *testing.T) {
	db := goninjatest.NewDB(t, &models.Author{}, &models.Book{})
	srv := goninjatest.NewServer(t, api.NewAuthorResource(db))

	body := `{"name":"Ursula K. Le Guin","bio":"American author."}`
	resp, err := http.Post(srv.URL+"/authors", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST /authors: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /authors status = %d, want 201", resp.StatusCode)
	}

	listResp, err := http.Get(srv.URL + "/authors")
	if err != nil {
		t.Fatalf("GET /authors: %v", err)
	}
	defer listResp.Body.Close()

	var envelope struct {
		Total int `json:"total"`
	}
	if err := json.NewDecoder(listResp.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode list envelope: %v", err)
	}
	if envelope.Total != 1 {
		t.Errorf("total = %d, want 1", envelope.Total)
	}
}

// TestAuthorResource_OrderValidation covers the request-time half of the
// "fail loudly" pair: an order field outside the generated whitelist is a
// 400, not a silently unordered 200. The whitelist is also what keeps
// ordering injection-safe, so this doubles as a check that it is consulted.
func TestAuthorResource_OrderValidation(t *testing.T) {
	db := goninjatest.NewDB(t, &models.Author{}, &models.Book{})
	srv := goninjatest.NewServer(t, api.NewAuthorResource(db))

	tests := []struct {
		name  string
		query string
		want  int
	}{
		{"known field", "?order=name", http.StatusOK},
		{"known field descending", "?order=-name", http.StatusOK},
		{"no order at all", "", http.StatusOK},
		{"unknown field", "?order=nonexistent", http.StatusBadRequest},
		{"typo in a real field", "?order=-nmae", http.StatusBadRequest},
		{"injection attempt", "?order=name;DROP TABLE authors", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(srv.URL + "/authors" + tt.query)
			if err != nil {
				t.Fatalf("GET /authors%s: %v", tt.query, err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != tt.want {
				t.Fatalf("GET /authors%s: got %d, want %d", tt.query, resp.StatusCode, tt.want)
			}
		})
	}
}
