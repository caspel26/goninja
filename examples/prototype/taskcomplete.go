// taskcomplete.go — never touched by the generator. Mirrors
// bookpublish.go's custom action for Task: POST /tasks/{id}/complete.
// Paired with it deliberately, so alreadyCompletedError (errors.go) and
// bookpublish.go's alreadyPublishedError give appErrorMapper two distinct
// resources' error types to compose — proving app.SetErrorMapper (main.go)
// as a genuinely app-wide registration point, not a stand-in for a
// single resource's own SetErrorMapper.
package main

import (
	"errors"
	"net/http"

	"gorm.io/gorm"

	"github.com/caspel26/goninja"
	"github.com/caspel26/goninja/examples/prototype/internal/api"
	"github.com/caspel26/goninja/examples/prototype/models"
	"github.com/caspel26/goninja/openapi"
)

// taskActions returns the custom actions to declare on r via SetActions.
func taskActions(r *api.TaskResource) []goninja.Action {
	return []goninja.Action{
		{
			Name:    "complete",
			Detail:  true,
			Method:  http.MethodPost,
			UrlPath: "complete",
			Handler: completeTaskHandler(r),
			Summary: "Complete a task",
			Responses: map[string]openapi.Response{
				"200": {Description: "OK"},
				"404": {Description: "Not found"},
			},
		},
	}
}

// completeTaskHandler flips a task's Done flag and returns the updated
// row. Rejects a task that's already done with alreadyCompletedError
// (errors.go) — mapped by app.SetErrorMapper (main.go), not a mapper set
// on TaskResource itself, since this error is handled the same way
// app-wide.
func completeTaskHandler(r *api.TaskResource) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		id := req.PathValue("id")

		var task models.Task
		if err := r.DB(ctx).First(&task, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = goninja.NotFound{Resource: "task", ID: id}
			}
			goninja.Respond(w, r.ErrorMapper(), err)
			return
		}
		if task.Done {
			goninja.Respond(w, r.ErrorMapper(), alreadyCompletedError{TaskID: id})
			return
		}

		if err := r.DB(ctx).Model(&models.Task{}).Where("id = ?", id).
			Update("done", true).Error; err != nil {
			goninja.Respond(w, r.ErrorMapper(), err)
			return
		}
		out, err := r.Retrieve(ctx, id)
		if err != nil {
			goninja.Respond(w, r.ErrorMapper(), err)
			return
		}
		goninja.RespondJSON(w, http.StatusOK, out)
	}
}
