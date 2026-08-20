// Package chiadapter lets goninja's generated resources mount on a chi
// router or sub-router. Wrap it once with New and pass the result anywhere
// goninja.Router is expected (goninja.API's Mount/MountWithConfig/
// MountDocs, or a resource's own Register).
package chiadapter

import (
	"net/http"

	"github.com/caspel26/goninja/router"
	"github.com/go-chi/chi/v5"
)

// Adapter satisfies goninja.Router by translating a stdlib-style pattern
// ("GET /books/{id}") into chi's own route registration. chi already uses
// the same "{name}" path-param syntax as net/http, so no path translation
// is needed — only binding chi's own per-request param store back onto
// req.PathValue.
type Adapter struct {
	r chi.Router
}

// New wraps r — the top-level chi.Router or any sub-router mounted via
// r.Route/r.Mount — as a goninja.Router.
func New(r chi.Router) *Adapter {
	return &Adapter{r: r}
}

// HandleFunc registers pattern on the wrapped chi router, binding each
// named wildcard chi matched back onto the request via SetPathValue before
// calling handler, so a generated handler's req.PathValue("id") call sees
// the right value without any changes to the handler itself.
func (a *Adapter) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	p, err := router.ParsePattern(pattern)
	if err != nil {
		panic("chiadapter: " + err.Error())
	}

	path := p.TranslatePath(router.StyleBrace)
	fn := func(w http.ResponseWriter, r *http.Request) {
		router.BindPathValues(r, p.Params, func(name string) string {
			return chi.URLParam(r, name)
		})
		handler(w, r)
	}

	if p.Method == "" {
		a.r.HandleFunc(path, fn)
		return
	}
	a.r.MethodFunc(p.Method, path, fn)
}
