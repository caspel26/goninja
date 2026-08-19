// Command prototype wires generated code into a real net/http server
// backed by a real Postgres database. It carries three models — Task,
// Author, Book — to exercise the engine across more than one model and
// prove CRUD works end to end against Postgres, with Retrieve
// automatically preloading relation fields (Book.Author) without the
// caller asking for it.
//
// Migrations are delegated to gorm's AutoMigrate — goninja does not
// generate migrations.
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

	"github.com/caspel26/goninja"
	"github.com/caspel26/goninja/docsui"
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
	app := goninja.NewAPI("goninja prototype", "0.1.0")

	bookAPI := api.NewBookResource(db)
	bookAPI.SetActions(bookActions(bookAPI)...) // adds POST /books/{id}/publish; see bookpublish.go

	app.Mount(mux,
		api.NewTaskResource(db),
		api.NewAuthorResource(db),
		bookAPI,
	)
	// docsui.ReDoc() is a drop-in alternative to docsui.SwaggerUI() here.
	app.MountDocs(mux, "/docs", docsui.SwaggerUI())

	log.Println("prototype listening on :8080 (/tasks, /authors, /books — GET, POST, GET/{id}, PUT/{id}, DELETE/{id}, POST/{id}/publish; /docs for OpenAPI/Swagger UI)")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
