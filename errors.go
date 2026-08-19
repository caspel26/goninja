package goninja

import "fmt"

// NotFound is returned by generated Retrieve/Update/Delete methods when no
// row matches the given ID. Framework error type referenced from generated
// code; see plan section 5.11 for how an ErrorMapper (Phase 3+) is meant to
// translate it into an HTTP response.
type NotFound struct {
	Resource string
	ID       any
}

func (e NotFound) Error() string {
	return fmt.Sprintf("%s %v not found", e.Resource, e.ID)
}
