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
	app.SetErrorMapper(appErrorMappings()...) // maps alreadyCompletedError app-wide; see errors.go

	// PROTOTYPE_API_KEY is optional — set it to see goninja.Authenticator
	// protect create/update/delete (and, via actionAuth below, the two
	// custom actions too) end to end (see auth.go); unset, the prototype
	// stays fully public for frictionless local exploration. actionAuth is
	// built here (nil when apiKey is unset) so it can be passed straight
	// into bookActions/taskActions via Action.Auth, instead of separately
	// relisting "publish"/"complete" in cfg.DefaultAuth.Routes below — see
	// bookActions (bookpublish.go) for why that separate list is exactly
	// the mistake Action.Auth exists to avoid.
	apiKey := os.Getenv("PROTOTYPE_API_KEY")
	var actionAuth *goninja.RouteAuth
	if apiKey != "" {
		actionAuth = &goninja.RouteAuth{Auth: []goninja.Authenticator{newAPIKeyAuth(apiKey)}}
	}

	taskAPI := api.NewTaskResource(db, goninja.Actions(taskActions, actionAuth)) // adds POST /tasks/{id}/complete; see taskcomplete.go

	bookAPI := api.NewBookResource(db, goninja.Actions(bookActions, actionAuth)) // adds POST /books/{id}/publish; see bookpublish.go
	bookAPI.SetErrorMapper(bookErrorMapper())                                    // maps alreadyPublishedError to 409; see errors.go

	resources := []goninja.Resource{
		taskAPI,
		api.NewAuthorResource(db),
		bookAPI,
	}

	if apiKey != "" {
		cfg := goninja.Config{
			DefaultAuth: goninja.AuthPolicy{
				Routes: []goninja.Route{goninja.RouteCreate, goninja.RouteUpdate, goninja.RouteDelete},
				Auth:   []goninja.Authenticator{newAPIKeyAuth(apiKey)},
			},
		}
		app.MountWithConfig(mux, cfg, resources...)
		log.Println("PROTOTYPE_API_KEY set: create/update/delete/publish/complete require X-API-Key")
	} else {
		app.Mount(mux, resources...)
	}
	// docsui.ReDoc() is a drop-in alternative to docsui.SwaggerUI() here.
	app.MountDocs(mux, "/docs", docsui.SwaggerUI())

	log.Println("prototype listening on :8080 (/tasks, /authors, /books — GET, POST, GET/{id}, PUT/{id}, DELETE/{id}, POST/{id}/complete on tasks, POST/{id}/publish on books; /docs for OpenAPI/Swagger UI)")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
