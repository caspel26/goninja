// Command prototype wires the generated code into a real net/http server.
// It carries two distinct models (Task, Author) precisely to prove the
// Phase 1 exit criterion in goninja-implementation-plan.md: the engine
// generalizes beyond the single model used for the Phase 0 decision gate.
//
// Run:
//
//	go run ../../cmd/goninja generate \
//	  -models ./models -out ./internal/api \
//	  -package api -models-import github.com/caspel26/goninja/examples/prototype/models
//	go run .
package main

import (
	"log"
	"net/http"

	"github.com/caspel26/goninja/examples/prototype/internal/api"
)

func main() {
	mux := http.NewServeMux()

	tasks := api.NewTaskResource()
	tasks.Register(mux)

	authors := api.NewAuthorResource()
	authors.Register(mux)

	log.Println("prototype listening on :8080 (GET/POST /tasks, GET /tasks/{id}, GET/POST /authors, GET /authors/{id})")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
