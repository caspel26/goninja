package main

// Proves Config.StrictAuth end to end against a real generated resource —
// not just that the codegen text calls CheckStrictAuth (internal/codegen
// has that test separately), but that it actually panics/doesn't panic at
// Register time through goninja.API.MountWithConfig.

import (
	"net/http"
	"testing"

	"github.com/caspel26/goninja"
	"github.com/caspel26/goninja/examples/prototype/internal/api"
	"github.com/caspel26/goninja/examples/prototype/models"
	"github.com/caspel26/goninja/goninjatest"
)

func TestStrictAuth_PanicsOnUnclassifiedRoute(t *testing.T) {
	db := goninjatest.NewDB(t, &models.Task{})

	defer func() {
		if recover() == nil {
			t.Fatal("expected Register to panic under StrictAuth with no auth policy configured at all")
		}
	}()

	app := goninja.NewAPI("test", "0.0.0")
	cfg := goninja.Config{StrictAuth: true} // no DefaultAuth.Routes, no per-resource override
	app.MountWithConfig(http.NewServeMux(), cfg, api.NewTaskResource(db))
}

func TestStrictAuth_PassesWhenEveryRouteIsClassified(t *testing.T) {
	db := goninjatest.NewDB(t, &models.Task{})

	app := goninja.NewAPI("test", "0.0.0")
	cfg := goninja.Config{
		StrictAuth: true,
		DefaultAuth: goninja.AuthPolicy{
			Routes: []goninja.Route{
				goninja.RouteList, goninja.RouteRetrieve,
				goninja.RouteCreate, goninja.RouteUpdate, goninja.RouteDelete,
			},
		},
	}
	// Must not panic: every CRUD route is named in DefaultAuth.Routes.
	app.MountWithConfig(http.NewServeMux(), cfg, api.NewTaskResource(db))
}
