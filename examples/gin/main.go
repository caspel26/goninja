// Command gin is a minimal, runnable proof that a goninja-generated
// resource mounts on gin unchanged — see adapters/gin. It uses an
// in-memory SQLite DB so it runs with no external dependencies:
//
//	go run ./examples/gin
//	curl -X POST localhost:8080/tasks -d '{"title":"write docs"}'
//	curl localhost:8080/tasks
//	open http://localhost:8080/docs
package main

import (
	"log"

	"github.com/caspel26/goninja"
	ginadapter "github.com/caspel26/goninja/adapters/gin"
	"github.com/caspel26/goninja/docsui"
	"github.com/caspel26/goninja/examples/gin/internal/api"
	"github.com/caspel26/goninja/examples/gin/models"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
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

	engine := gin.Default()

	r := ginadapter.New(engine)
	app := goninja.NewAPI("goninja gin example", "0.1.0")
	app.Mount(r, api.NewTaskResource(db))
	app.MountDocs(r, "/docs", docsui.SwaggerUI())

	log.Println("listening on :8080")
	log.Fatal(engine.Run(":8080"))
}
