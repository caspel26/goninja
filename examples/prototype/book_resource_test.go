package main

// Proves examples/prototype/models/book.go's CreatedAt time.Time filter
// works end to end over real HTTP against a real generated BookResource —
// not just that it compiles (internal/codegen has that regression test
// separately). This is the exact field/tag combination that broke in an
// external consumer project before the fix.

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/caspel26/goninja/examples/prototype/internal/api"
	"github.com/caspel26/goninja/examples/prototype/models"
	"github.com/caspel26/goninja/goninjatest"
)

func TestBookResource_CreatedAtFilter(t *testing.T) {
	db := goninjatest.NewDB(t, &models.Author{}, &models.Book{})

	author := models.Author{ID: "a1111111-1111-1111-1111-111111111111", Name: "Author One"}
	if err := db.Create(&author).Error; err != nil {
		t.Fatalf("seed author: %v", err)
	}

	older := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := time.Date(2024, 6, 15, 12, 30, 0, 0, time.UTC)
	books := []models.Book{
		{ID: "b1111111-1111-1111-1111-111111111111", Title: "Old Book", AuthorID: author.ID, CreatedAt: older},
		{ID: "b2222222-2222-2222-2222-222222222222", Title: "New Book", AuthorID: author.ID, CreatedAt: newer},
	}
	for i := range books {
		if err := db.Create(&books[i]).Error; err != nil {
			t.Fatalf("seed book: %v", err)
		}
	}

	srv := goninjatest.NewServer(t, api.NewBookResource(db))

	t.Run("exact match returns only the matching book", func(t *testing.T) {
		resp, err := http.Get(srv.URL + "/books?created_at=" + newer.Format(time.RFC3339))
		if err != nil {
			t.Fatalf("GET /books?created_at=...: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		var envelope struct {
			Items []struct {
				Title string `json:"title"`
			} `json:"items"`
			Total int `json:"total"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if envelope.Total != 1 || len(envelope.Items) != 1 || envelope.Items[0].Title != "New Book" {
			t.Errorf("envelope = %+v, want exactly one item, \"New Book\"", envelope)
		}
	})

	t.Run("invalid value is a 400, not a 500", func(t *testing.T) {
		resp, err := http.Get(srv.URL + "/books?created_at=not-a-timestamp")
		if err != nil {
			t.Fatalf("GET /books?created_at=not-a-timestamp: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", resp.StatusCode)
		}
	})
}
