package main

// Performance benchmarks for the three concerns the next release's roadmap
// calls out: base serialization, filter-clause building, and the
// automatic-Preload cost Retrieve always pays on a relation field. Each
// seeds seedRows rows directly via *gorm.DB (so seeding time never
// pollutes the measured loop) then times real HTTP requests against a
// goninjatest server — the same DB/server helpers author_resource_test.go
// already uses, both of which accept testing.TB and so work unmodified
// from a *testing.B.
//
// Run with: make bench (or: go test ./examples/prototype/... -bench=. -run=^$)

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/caspel26/goninja/examples/prototype/internal/api"
	"github.com/caspel26/goninja/examples/prototype/models"
	"github.com/caspel26/goninja/goninjatest"
	"github.com/google/uuid"
)

// drainAndClose fully reads resp's body before closing it — skipping this
// prevents the underlying connection from being reused for keep-alive,
// forcing a new TCP connection (and a new ephemeral port) per request,
// which exhausts the local port range under a benchmark's iteration count.
func drainAndClose(resp *http.Response) {
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
}

// seedRows is the fixed row count every benchmark below seeds, so ns/op
// and allocs/op are comparable across benchmarks and across goninja
// versions.
const seedRows = 1000

func BenchmarkTaskList(b *testing.B) {
	db := goninjatest.NewDB(b, &models.Task{})
	for i := 0; i < seedRows; i++ {
		if err := db.Create(&models.Task{ID: uuid.NewString(), Title: fmt.Sprintf("task %d", i)}).Error; err != nil {
			b.Fatalf("seed: %v", err)
		}
	}
	srv := goninjatest.NewServer(b, api.NewTaskResource(db))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Get(srv.URL + "/tasks")
		if err != nil {
			b.Fatalf("Get: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			b.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
		drainAndClose(resp)
	}
}

func BenchmarkBookListFiltered(b *testing.B) {
	db := goninjatest.NewDB(b, &models.Author{}, &models.Book{})
	author := models.Author{ID: uuid.NewString(), Name: "Benchmark Author"}
	if err := db.Create(&author).Error; err != nil {
		b.Fatalf("seed author: %v", err)
	}
	for i := 0; i < seedRows; i++ {
		book := models.Book{
			ID:        uuid.NewString(),
			Title:     fmt.Sprintf("book %d", i),
			AuthorID:  author.ID,
			Price:     float64(i % 100),
			Published: i%2 == 0,
		}
		if err := db.Create(&book).Error; err != nil {
			b.Fatalf("seed book: %v", err)
		}
	}
	srv := goninjatest.NewServer(b, api.NewBookResource(db))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Get(srv.URL + "/books?published=true&price_min=10&price_max=80")
		if err != nil {
			b.Fatalf("Get: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			b.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
		drainAndClose(resp)
	}
}

func BenchmarkBookRetrievePreload(b *testing.B) {
	db := goninjatest.NewDB(b, &models.Author{}, &models.Book{})
	author := models.Author{ID: uuid.NewString(), Name: "Benchmark Author"}
	if err := db.Create(&author).Error; err != nil {
		b.Fatalf("seed author: %v", err)
	}
	book := models.Book{ID: uuid.NewString(), Title: "The Benchmark Book", AuthorID: author.ID, Price: 42}
	if err := db.Create(&book).Error; err != nil {
		b.Fatalf("seed book: %v", err)
	}
	// seedRows unrelated rows so Retrieve's single-row query and Preload
	// aren't measuring an effectively-empty table.
	for i := 0; i < seedRows; i++ {
		other := models.Book{ID: uuid.NewString(), Title: fmt.Sprintf("book %d", i), AuthorID: author.ID, Price: float64(i)}
		if err := db.Create(&other).Error; err != nil {
			b.Fatalf("seed book: %v", err)
		}
	}
	srv := goninjatest.NewServer(b, api.NewBookResource(db))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Get(srv.URL + "/books/" + book.ID)
		if err != nil {
			b.Fatalf("Get: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			b.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
		drainAndClose(resp)
	}
}
