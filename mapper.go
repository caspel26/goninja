package goninja

import (
	"encoding/json"
	"errors"
	"net/http"
)

// ErrorMapper translates an error returned by a resource method (or a
// handler-level error like BadRequest) into an HTTP status code and a JSON
// response body. Generated handlers call it through
// BaseResource.ErrorMapper(), which falls back to DefaultErrorMapper when
// none has been set — see BaseResource.SetErrorMapper.
type ErrorMapper interface {
	MapError(err error) (status int, body any)
}

// DefaultErrorMapper is the ErrorMapper used when a resource has none
// configured. It maps the framework's own error types (NotFound,
// ValidationError, BadRequest, Unauthorized) to their conventional status
// codes and everything else to 500, without leaking the underlying error
// message.
type DefaultErrorMapper struct{}

func (DefaultErrorMapper) MapError(err error) (int, any) {
	var nf NotFound
	if errors.As(err, &nf) {
		return http.StatusNotFound, map[string]string{
			"code":  nf.ErrorCode(),
			"error": nf.Error(),
		}
	}

	var ve ValidationError
	if errors.As(err, &ve) {
		return http.StatusUnprocessableEntity, map[string]any{
			"code":   ve.ErrorCode(),
			"errors": ve.Fields,
		}
	}

	var br BadRequest
	if errors.As(err, &br) {
		return http.StatusBadRequest, map[string]string{
			"code":  br.ErrorCode(),
			"error": br.Detail,
		}
	}

	var ua Unauthorized
	if errors.As(err, &ua) {
		return http.StatusUnauthorized, map[string]string{
			"code":  ua.ErrorCode(),
			"error": ua.Error(),
		}
	}

	return http.StatusInternalServerError, map[string]string{
		"code":  "INTERNAL",
		"error": "internal error",
	}
}

// Respond maps err through mapper (DefaultErrorMapper if mapper is nil) and
// writes the resulting status/body as JSON.
func Respond(w http.ResponseWriter, mapper ErrorMapper, err error) {
	if mapper == nil {
		mapper = DefaultErrorMapper{}
	}
	status, body := mapper.MapError(err)
	RespondJSON(w, status, body)
}

// RespondJSON writes v as a JSON response with the given status code. Used
// by generated handlers for both success and (via Respond) error responses,
// so there's a single JSON-encoding path per generated package rather than
// a per-model duplicate.
func RespondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
