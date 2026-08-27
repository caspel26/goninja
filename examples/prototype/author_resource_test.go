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
	"time"

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

// TestAuthorResource_CreatedAtFilter is the runtime half of the
// unused-strconv regression Author's CreatedAt pins (see models/author.go):
// the compile-level guard lives in internal/codegen, but this proves the
// time.Time filter it generates actually filters over real HTTP, and that
// a malformed timestamp is a 400 rather than a 500 or a silent full list.
func TestAuthorResource_CreatedAtFilter(t *testing.T) {
	db := goninjatest.NewDB(t, &models.Author{}, &models.Book{})

	older := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := time.Date(2024, 6, 15, 12, 30, 0, 0, time.UTC)
	for _, a := range []models.Author{
		{ID: "a1111111-1111-1111-1111-111111111111", Name: "Older", CreatedAt: older},
		{ID: "a2222222-2222-2222-2222-222222222222", Name: "Newer", CreatedAt: newer},
	} {
		if err := db.Create(&a).Error; err != nil {
			t.Fatalf("seed author: %v", err)
		}
	}

	srv := goninjatest.NewServer(t, api.NewAuthorResource(db))

	t.Run("exact match returns only that author", func(t *testing.T) {
		resp, err := http.Get(srv.URL + "/authors?created_at=" + newer.Format(time.RFC3339))
		if err != nil {
			t.Fatalf("GET /authors?created_at=...: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		var envelope struct {
			Items []struct {
				Name string `json:"name"`
			} `json:"items"`
			Total int `json:"total"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if envelope.Total != 1 || len(envelope.Items) != 1 || envelope.Items[0].Name != "Newer" {
			t.Errorf("envelope = %+v, want exactly one item, \"Newer\"", envelope)
		}
	})

	t.Run("malformed timestamp is a 400", func(t *testing.T) {
		resp, err := http.Get(srv.URL + "/authors?created_at=2024-01-01")
		if err != nil {
			t.Fatalf("GET /authors?created_at=2024-01-01: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400 (date without time/offset isn't RFC 3339)", resp.StatusCode)
		}
	})
}
