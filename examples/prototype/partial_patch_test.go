package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/caspel26/goninja/examples/prototype/internal/api"
	"github.com/caspel26/goninja/examples/prototype/models"
	"github.com/caspel26/goninja/goninjatest"
)

// PATCH must distinguish properties missing from the request from explicit
// zero values. This exercises the generated handler over SQLite and HTTP,
// rather than only inspecting the pointer-shaped DTO.
func TestBookResource_PatchPreservesOmittedFieldsAndWritesZeroValues(t *testing.T) {
	db := goninjatest.NewDB(t, &models.Author{}, &models.Book{})
	author := models.Author{ID: "a1111111-1111-1111-1111-111111111111", Name: "Author"}
	if err := db.Create(&author).Error; err != nil {
		t.Fatal(err)
	}
	book := models.Book{
		ID: "b1111111-1111-1111-1111-111111111111", Title: "Keep this title",
		AuthorID: author.ID, Price: 19.99, Published: true,
	}
	if err := db.Create(&book).Error; err != nil {
		t.Fatal(err)
	}

	srv := goninjatest.NewServer(t, api.NewBookResource(db))
	req, err := http.NewRequest(http.MethodPatch, srv.URL+"/books/"+book.ID,
		strings.NewReader(`{"published":false,"price":0}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("PATCH status = %d, want 200", resp.StatusCode)
	}

	var got struct {
		Title     string  `json:"title"`
		Price     float64 `json:"price"`
		Published bool    `json:"published"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Title != book.Title || got.Price != 0 || got.Published {
		t.Errorf("PATCH response = %+v, want title preserved and explicit zero values written", got)
	}

	var stored models.Book
	if err := db.First(&stored, "id = ?", book.ID).Error; err != nil {
		t.Fatal(err)
	}
	if stored.Title != book.Title || stored.Price != 0 || stored.Published {
		t.Errorf("stored book = %+v, want title preserved and explicit zero values written", stored)
	}
}
