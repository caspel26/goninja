package main

// Demonstrates app.SetErrorMapper (main.go, errors.go) actually taking
// effect end to end: a real TaskResource, mounted via goninja.API.Mount
// (not goninjatest.NewServer, which skips SetConfig entirely) so the
// Mount -> SetConfig -> BaseResource.ErrorMapper() resolution chain runs
// for real, over a real HTTP round trip.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/caspel26/goninja"
	"github.com/caspel26/goninja/examples/prototype/internal/api"
	"github.com/caspel26/goninja/examples/prototype/models"
	"github.com/caspel26/goninja/goninjatest"
)

func TestAppSetErrorMapper_MapsAlreadyCompletedAppWide(t *testing.T) {
	db := goninjatest.NewDB(t, &models.Task{})

	taskAPI := api.NewTaskResource(db)
	taskAPI.SetActions(taskActions(taskAPI)...)

	app := goninja.NewAPI("test", "0.0.0")
	app.SetErrorMapper(appErrorMappings()...)

	mux := http.NewServeMux()
	app.Mount(mux, taskAPI)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	createResp, err := http.Post(srv.URL+"/tasks", "application/json", strings.NewReader(`{"title":"write tests"}`))
	if err != nil {
		t.Fatalf("POST /tasks: %v", err)
	}
	defer createResp.Body.Close()
	var created struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createResp.Body).Decode(&created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	completeURL := srv.URL + "/tasks/" + created.ID + "/complete"

	firstResp, err := http.Post(completeURL, "application/json", nil)
	if err != nil {
		t.Fatalf("first POST .../complete: %v", err)
	}
	defer firstResp.Body.Close()
	if firstResp.StatusCode != http.StatusOK {
		t.Fatalf("first POST .../complete status = %d, want 200", firstResp.StatusCode)
	}

	secondResp, err := http.Post(completeURL, "application/json", nil)
	if err != nil {
		t.Fatalf("second POST .../complete: %v", err)
	}
	defer secondResp.Body.Close()
	if secondResp.StatusCode != http.StatusConflict {
		t.Fatalf("second POST .../complete status = %d, want 409", secondResp.StatusCode)
	}

	var body struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(secondResp.Body).Decode(&body); err != nil {
		t.Fatalf("decode conflict response: %v", err)
	}
	if body.Code != "ALREADY_COMPLETED" {
		t.Errorf("code = %q, want %q", body.Code, "ALREADY_COMPLETED")
	}
}
