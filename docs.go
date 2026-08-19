package goninja

import (
	"embed"
	"net/http"
	"strings"
)

// DocsUI renders a documentation viewer for a mounted OpenAPI document.
// goninja ships SwaggerUI and ReDoc (plan section 6/Fase 5); implement
// DocsUI yourself to plug in something else — a hosted viewer, a custom
// static-site generator, whatever — MountDocs never hardcodes which
// renderer backs /docs.
type DocsUI interface {
	// Index returns the HTML page that boots this UI against specPath
	// (the "<path>/openapi.json" route MountDocs mounts alongside it).
	Index(specPath string) []byte
	// Assets returns the static files this UI needs, keyed by the route
	// suffix they're served under (e.g. "swagger-ui-bundle.js") — Index's
	// HTML should reference each by that same relative filename.
	Assets() map[string]DocsAsset
}

// DocsAsset is one static file a DocsUI serves alongside its Index page.
type DocsAsset struct {
	Data        []byte
	ContentType string
}

// MountDocs serves api's merged OpenAPI document and a documentation UI at
// path (e.g. "/docs"): the spec as JSON at "<path>/openapi.json", the UI
// itself at "<path>/" (a request to "<path>" without the trailing slash
// redirects there — a DocsUI's Index references its Assets by relative
// filename, e.g. "swagger-ui.css", which only resolves correctly once
// "<path>/" is the page's actual base URL). ui selects the renderer —
// pass SwaggerUI() or ReDoc() (both fully embedded, no external CDN) or
// your own DocsUI; nil defaults to SwaggerUI().
func MountDocs(mux *http.ServeMux, api *API, path string, ui DocsUI) {
	if ui == nil {
		ui = SwaggerUI()
	}
	path = "/" + strings.Trim(path, "/")
	specPath := path + "/openapi.json"

	mux.HandleFunc("GET "+specPath, func(w http.ResponseWriter, r *http.Request) {
		RespondJSON(w, http.StatusOK, api.Spec())
	})

	for route, asset := range ui.Assets() {
		asset := asset
		mux.HandleFunc("GET "+path+"/"+route, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", asset.ContentType)
			w.Write(asset.Data)
		})
	}

	index := ui.Index(specPath)
	mux.HandleFunc("GET "+path+"/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(index)
	})
	mux.HandleFunc("GET "+path, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, path+"/", http.StatusMovedPermanently)
	})
}

// mustReadAsset reads name out of fsys, an embed.FS baked in at build
// time — a missing file here is a goninja packaging bug, not something a
// caller can hit at runtime, hence the panic instead of a plumbed error.
func mustReadAsset(fsys embed.FS, name, contentType string) DocsAsset {
	b, err := fsys.ReadFile(name)
	if err != nil {
		panic("goninja: embedded doc asset missing: " + name)
	}
	return DocsAsset{Data: b, ContentType: contentType}
}
