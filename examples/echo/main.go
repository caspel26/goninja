// Command echo is a minimal, runnable proof that a goninja-generated
// resource mounts on echo unchanged — see adapters/echo. It uses an
// in-memory SQLite DB so it runs with no external dependencies:
//
//	go run ./examples/echo
//	curl -X POST localhost:8080/tasks -d '{"title":"write docs"}'
//	curl localhost:8080/tasks
//	open http://localhost:8080/docs
package main

import (
	"log"

	"github.com/caspel26/goninja"
	echoadapter "github.com/caspel26/goninja/adapters/echo"
	"github.com/caspel26/goninja/docsui"
	"github.com/caspel26/goninja/examples/echo/internal/api"
	"github.com/caspel26/goninja/examples/echo/models"
	"github.com/glebarez/sqlite"
	"github.com/labstack/echo/v4"
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

	e := echo.New()

	r := echoadapter.New(e)
	app := goninja.NewAPI("goninja echo example", "0.1.0")
	app.Mount(r, api.NewTaskResource(db))
	app.MountDocs(r, "/docs", docsui.SwaggerUI())

	log.Println("listening on :8080")
	log.Fatal(e.Start(":8080"))
}
