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

// ErrorMapping associates a predicate over an error with how to map it —
// built with NewErrorMapping[T] for a specific error type. Composed via
// NewErrorMapper into an ErrorMapper that tries each Mapping in order —
// one handler per error type, declared as data instead of a hand-written
// switch.
type ErrorMapping struct {
	Matches func(err error) bool
	Map     func(err error) (status int, body any)
}

// NewErrorMapping returns an ErrorMapping that matches any error
// satisfying errors.As into T (so a wrapped error still matches, same as
// DefaultErrorMapper's own checks) and maps it via fn.
func NewErrorMapping[T error](fn func(err T) (status int, body any)) ErrorMapping {
	return ErrorMapping{
		Matches: func(err error) bool {
			var target T
			return errors.As(err, &target)
		},
		Map: func(err error) (int, any) {
			var target T
			errors.As(err, &target)
			return fn(target)
		},
	}
}

// ComposedErrorMapper tries each Mapping in order and returns the first
// match; if none match, it falls back to Fallback (DefaultErrorMapper{}
// if Fallback is nil). Build one with NewErrorMapper rather than directly.
type ComposedErrorMapper struct {
	Mappings []ErrorMapping
	Fallback ErrorMapper
}

func (m ComposedErrorMapper) MapError(err error) (int, any) {
	for _, mapping := range m.Mappings {
		if mapping.Matches(err) {
			return mapping.Map(err)
		}
	}
	fallback := m.Fallback
	if fallback == nil {
		fallback = DefaultErrorMapper{}
	}
	return fallback.MapError(err)
}

// NewErrorMapper composes mappings into an ErrorMapper, tried in order —
// the declarative equivalent of registering one exception handler per
// error type, instead of writing a MapError switch by hand. Falls back to
// DefaultErrorMapper for anything no mapping matches.
func NewErrorMapper(mappings ...ErrorMapping) ErrorMapper {
	return ComposedErrorMapper{Mappings: mappings}
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
