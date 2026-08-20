// Package docsui renders a documentation viewer (Swagger UI or ReDoc, both
// vendored/embedded — no external CDN) for a merged OpenAPI document.
// MountDocs takes a SpecSource interface rather than a concrete document
// type so this package never has to import goninja.API's package directly
// (goninja already imports docsui, so importing back would cycle).
package docsui

import (
	"embed"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/caspel26/goninja/openapi"
	"github.com/caspel26/goninja/router"
)

// DocsUI renders a documentation viewer for a mounted OpenAPI document.
// goninja ships SwaggerUI and ReDoc; implement DocsUI yourself to plug in
// something else — a hosted viewer, a custom static-site generator,
// whatever — MountDocs never hardcodes which renderer backs /docs.
type DocsUI interface {
	// Index returns the HTML page that boots this UI against specPath
	// (the "<path>/openapi.json" route MountDocs mounts alongside it).
	Index(specPath string) []byte
	// Assets returns the static files this UI needs, keyed by the route
	// suffix they're served under (e.g. "swagger-ui-bundle.js") — Index's
	// HTML should reference each by that same relative filename.
	Assets() map[string]DocsAsset
}

// headerContentType is the HTTP header name MountDocs sets on every route
// it registers (spec JSON, each UI asset, the index page).
const headerContentType = "Content-Type"

// DocsAsset is one static file a DocsUI serves alongside its Index page.
type DocsAsset struct {
	Data        []byte
	ContentType string
}

// SpecSource is anything that can produce a merged OpenAPI document —
// goninja.API (root package) is the usual one, but MountDocs only needs
// this narrow interface rather than importing that concrete type, which
// would cycle back to root (see the package doc comment).
type SpecSource interface {
	Spec() openapi.Spec
}

// MountDocs serves api's merged OpenAPI document and a documentation UI at
// path (e.g. "/docs"): the spec as JSON at "<path>/openapi.json", the UI
// itself at "<path>/" (a request to "<path>" without the trailing slash
// redirects there — a DocsUI's Index references its Assets by relative
// filename, e.g. "swagger-ui.css", which only resolves correctly once
// "<path>/" is the page's actual base URL). ui selects the renderer —
// pass SwaggerUI() or ReDoc() (both fully embedded, no external CDN) or
// your own DocsUI; nil defaults to SwaggerUI().
func MountDocs(mux router.Router, api SpecSource, path string, ui DocsUI) {
	if ui == nil {
		ui = SwaggerUI()
	}
	path = "/" + strings.Trim(path, "/")
	specPath := path + "/openapi.json"

	mux.HandleFunc("GET "+specPath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headerContentType, "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(api.Spec())
	})

	for route, asset := range ui.Assets() {
		asset := asset
		mux.HandleFunc("GET "+path+"/"+route, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(headerContentType, asset.ContentType)
			w.Write(asset.Data)
		})
	}

	index := ui.Index(specPath)
	mux.HandleFunc("GET "+path+"/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headerContentType, "text/html; charset=utf-8")
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
