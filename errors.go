package goninja

import "fmt"

// CodedError is implemented by every framework error type below. ErrorCode
// returns the JSON body's "code" field — the type's own conventional
// default, or the value set on its Code field when a caller wants a more
// specific machine-readable identifier than the HTTP status alone provides.
type CodedError interface {
	error
	ErrorCode() string
}

func codeOr(code, fallback string) string {
	if code != "" {
		return code
	}
	return fallback
}

// NotFound is returned by generated Retrieve/Update/Delete methods when no
// row matches the given ID. Framework error type referenced from generated
// code; mapped to 404 by DefaultErrorMapper. Code
// overrides the JSON body's "code" field (default "NOT_FOUND") for a caller
// that wants a more specific machine-readable identifier than the HTTP
// status alone provides.
type NotFound struct {
	Resource string
	ID       any
	Code     string
}

func (e NotFound) Error() string {
	return fmt.Sprintf("%s %v not found", e.Resource, e.ID)
}

func (e NotFound) ErrorCode() string {
	return codeOr(e.Code, "NOT_FOUND")
}

// ValidationError is returned by Validate when one or more fields fail
// their `validate` struct tag rules. Fields maps the JSON field name to the
// failed validator tag (e.g. "required", "max"). Mapped to 422 by
// DefaultErrorMapper. Code overrides the JSON body's "code" field (default
// "VALIDATION_FAILED").
type ValidationError struct {
	Fields map[string]string
	Code   string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation failed: %v", e.Fields)
}

func (e ValidationError) ErrorCode() string {
	return codeOr(e.Code, "VALIDATION_FAILED")
}

// BadRequest is returned by generated handlers for malformed input that
// never reaches a resource method (e.g. invalid JSON, an unparseable path
// ID). Mapped to 400 by DefaultErrorMapper. Code overrides the JSON body's
// "code" field (default "BAD_REQUEST") so a specific validation failure
// (e.g. an unknown ?order= field) can carry a more specific identifier than
// the generic status.
type BadRequest struct {
	Detail string
	Code   string
}

func (e BadRequest) Error() string {
	return e.Detail
}

func (e BadRequest) ErrorCode() string {
	return codeOr(e.Code, "BAD_REQUEST")
}

// Unauthorized is returned when every configured Authenticator declines a
// request. Framework error type mapped to 401 by DefaultErrorMapper. Code
// overrides the JSON body's "code" field (default "UNAUTHORIZED").
type Unauthorized struct {
	Detail string
	Code   string
}

func (e Unauthorized) Error() string {
	if e.Detail == "" {
		return "unauthorized"
	}
	return e.Detail
}

func (e Unauthorized) ErrorCode() string {
	return codeOr(e.Code, "UNAUTHORIZED")
}
