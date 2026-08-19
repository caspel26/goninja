// bookpublish.go — never touched by the generator. Demonstrates adding a
// custom action beyond the generated CRUD set (README's "Adding custom
// routes beyond CRUD"), goninja's equivalent of django-ninja-aio-crud's
// @action: POST /books/{id}/publish, mounted and documented automatically
// alongside the generated GET/POST/PUT/DELETE routes.
//
// Handler logic lives here, next to the model it operates on; main.go
// wires it in with a single explicit r.SetActions(bookActions(r)...) call
// so everything actually mounted stays visible at the call site instead of
// hidden behind a constructor.
package main

import (
	"net/http"

	"github.com/caspel26/goninja"
	"github.com/caspel26/goninja/examples/prototype/internal/api"
	"github.com/caspel26/goninja/examples/prototype/models"
	"github.com/caspel26/goninja/openapi"
)

// bookActions returns the custom actions to declare on r via SetActions.
func bookActions(r *api.BookResource) []goninja.Action {
	return []goninja.Action{
		{
			Name:    "publish",
			Detail:  true,
			Method:  http.MethodPost,
			UrlPath: "publish",
			Handler: publishBookHandler(r),
			Summary: "Publish a book",
			Responses: map[string]openapi.Response{
				"200": {Description: "OK"},
				"404": {Description: "Not found"},
			},
		},
	}
}

// publishBookHandler flips a book's Published flag and returns the updated
// row — reuses the generated Retrieve to build the response instead of
// duplicating its Preload/error-mapping logic.
func publishBookHandler(r *api.BookResource) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		id := req.PathValue("id")
		if err := r.DB(ctx).Model(&models.Book{}).Where("id = ?", id).
			Update("published", true).Error; err != nil {
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
