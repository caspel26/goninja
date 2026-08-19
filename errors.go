package goninja

import "fmt"

// NotFound is returned by generated Retrieve/Update/Delete methods when no
// row matches the given ID. Framework error type referenced from generated
// code; mapped to 404 by DefaultErrorMapper (plan section 5.11).
type NotFound struct {
	Resource string
	ID       any
}

func (e NotFound) Error() string {
	return fmt.Sprintf("%s %v not found", e.Resource, e.ID)
}

// ValidationError is returned by Validate when one or more fields fail
// their `validate` struct tag rules. Fields maps the JSON field name to the
// failed validator tag (e.g. "required", "max"). Mapped to 422 by
// DefaultErrorMapper.
type ValidationError struct {
	Fields map[string]string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation failed: %v", e.Fields)
}

// BadRequest is returned by generated handlers for malformed input that
// never reaches a resource method (e.g. invalid JSON, an unparseable path
// ID). Mapped to 400 by DefaultErrorMapper.
type BadRequest struct {
	Detail string
}

func (e BadRequest) Error() string {
	return e.Detail
}
