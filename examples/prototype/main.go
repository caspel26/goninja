// Command prototype wires the generated code into a real net/http server
// backed by a real Postgres database. It carries three models — Task,
// Author, Book — precisely to prove two plan exit criteria at once:
//
//   - Phase 1: the engine generalizes beyond one hand-tuned model.
//   - Phase 2: CRUD works end to end against Postgres, and Retrieve
//     automatically preloads relation fields (Book.Author) without the
//     caller asking for it.
//
// Migrations are delegated to gorm's AutoMigrate, per plan section 6
// (Phase 2) — goninja does not generate migrations.
//
// Run:
//
//	go run ../../cmd/goninja generate \
//	  -models ./models -out ./internal/api \
//	  -package api -models-import github.com/caspel26/goninja/examples/prototype/models
//	go run .   # expects PROTOTYPE_DSN, e.g.
//	           # "host=localhost user=$(whoami) dbname=goninja_prototype sslmode=disable"
package main

import (
	"log"
	"net/http"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/caspel26/goninja/examples/prototype/internal/api"
	"github.com/caspel26/goninja/examples/prototype/models"
)

func main() {
	dsn := os.Getenv("PROTOTYPE_DSN")
	if dsn == "" {
		log.Fatal("PROTOTYPE_DSN is required, e.g. \"host=localhost user=youruser dbname=goninja_prototype sslmode=disable\"")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Warn)})
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}

	if err := db.AutoMigrate(&models.Author{}, &models.Book{}, &models.Task{}); err != nil {
		log.Fatalf("running migrations: %v", err)
	}

	mux := http.NewServeMux()

	api.NewTaskResource(db).Register(mux)
	api.NewAuthorResource(db).Register(mux)
	api.NewBookResource(db).Register(mux)

	log.Println("prototype listening on :8080 (/tasks, /authors, /books — GET, POST, GET/{id}, PUT/{id}, DELETE/{id})")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
