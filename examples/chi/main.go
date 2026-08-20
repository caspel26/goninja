// Command chi is a minimal, runnable proof that a goninja-generated
// resource mounts on chi unchanged — see adapters/chi. It uses an
// in-memory SQLite DB so it runs with no external dependencies:
//
//	go run ./examples/chi
//	curl -X POST localhost:8080/tasks -d '{"title":"write docs"}'
//	curl localhost:8080/tasks
//	open http://localhost:8080/docs
package main

import (
	"log"
	"net/http"

	"github.com/caspel26/goninja"
	chiadapter "github.com/caspel26/goninja/adapters/chi"
	"github.com/caspel26/goninja/docsui"
	"github.com/caspel26/goninja/examples/chi/internal/api"
	"github.com/caspel26/goninja/examples/chi/models"
	"github.com/glebarez/sqlite"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&models.Task{}); err != nil {
		log.Fatalf("automigrate: %v", err)
	}

	mux := chi.NewRouter()

	r := chiadapter.New(mux)
	app := goninja.NewAPI("goninja chi example", "0.1.0")
	app.Mount(r, api.NewTaskResource(db))
	app.MountDocs(r, "/docs", docsui.SwaggerUI())

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
