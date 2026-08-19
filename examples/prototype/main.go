// Command prototype wires the Phase 0 generated code into a real
// net/http server. It exists to answer the goninja-implementation-plan.md
// Phase 0 decision-gate questions:
//   - is "edit struct → regenerate → use" acceptable?
//   - is the generated code readable/debuggable?
//   - how complex is it to keep the templates maintainable?
//
// Run:
//
//	go run ../../cmd/goninja generate \
//	  -models ./models -out ./internal/generated \
//	  -package generated -models-import github.com/caspel26/goninja/examples/prototype/models
//	go run .
package main

import (
	"log"
	"net/http"

	"github.com/caspel26/goninja/examples/prototype/internal/generated"
)

func main() {
	mux := http.NewServeMux()

	tasks := generated.NewTaskResource()
	tasks.Register(mux)

	log.Println("prototype listening on :8080 (GET/POST /tasks, GET /tasks/{id})")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
