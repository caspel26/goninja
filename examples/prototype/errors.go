// errors.go — never touched by the generator. Demonstrates a declarative
// custom error mapper built with goninja.NewErrorMapper/NewErrorMapping —
// goninja's equivalent of FastAPI's @app.exception_handler(T)/Django
// Ninja's @api.exception_handler(T) — at two different scopes:
// bookErrorMapper is a whole ErrorMapper set on BookResource alone
// (SetErrorMapper, main.go) since alreadyPublishedError only ever comes
// from bookpublish.go; appErrorMappings is a []ErrorMapping (including
// alreadyCompletedError from taskcomplete.go, a different resource's
// error) passed to app.SetErrorMapper (main.go), set once app-wide
// instead of being repeated per resource.
package main

import (
	"fmt"
	"net/http"

	"github.com/caspel26/goninja"
)

// alreadyPublishedError is returned by publishBookHandler (bookpublish.go)
// when the target book is already published.
type alreadyPublishedError struct {
	BookID string
}

func (e alreadyPublishedError) Error() string {
	return fmt.Sprintf("book %s is already published", e.BookID)
}

// alreadyCompletedError is returned by completeTaskHandler
// (taskcomplete.go) when the target task is already done.
type alreadyCompletedError struct {
	TaskID string
}

func (e alreadyCompletedError) Error() string {
	return fmt.Sprintf("task %s is already completed", e.TaskID)
}

// bookErrorMapper special-cases alreadyPublishedError as 409 Conflict and
// falls back to goninja.DefaultErrorMapper (NotFound/ValidationError/
// BadRequest/Unauthorized) for everything else on BookResource.
func bookErrorMapper() goninja.ErrorMapper {
	return goninja.NewErrorMapper(
		goninja.NewErrorMapping(func(err alreadyPublishedError) (int, any) {
			return http.StatusConflict, map[string]string{
				"code":  "ALREADY_PUBLISHED",
				"error": err.Error(),
			}
		}),
	)
}

// appErrorMappings are the ErrorMappings passed to app.SetErrorMapper
// (main.go) — here just alreadyCompletedError, but the point of setting
// them app-wide instead of on TaskResource alone is that adding a second
// resource with a similar "already done" error later means adding one
// more entry to this slice, not one more SetErrorMapper call per
// resource. API.SetErrorMapper takes ErrorMappings rather than a whole
// ErrorMapper precisely so entries like this one compose safely with
// whatever else gets added here later.
func appErrorMappings() []goninja.ErrorMapping {
	return []goninja.ErrorMapping{
		goninja.NewErrorMapping(func(err alreadyCompletedError) (int, any) {
			return http.StatusConflict, map[string]string{
				"code":  "ALREADY_COMPLETED",
				"error": err.Error(),
			}
		}),
	}
}
