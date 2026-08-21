package docsui

import "embed"

// redocAssets vendors ReDoc's standalone bundle (a single self-contained
// JS file — no separate CSS) so /docs works fully offline when ReDoc is
// selected — no external CDN. See redoc/LICENSE for ReDoc's license.
//
//go:embed redoc/redoc.standalone.js
var redocAssets embed.FS

type redocUI struct{}

// ReDoc is a DocsUI backed by ReDoc, vendored under redoc and embedded at
// build time. Pass it to MountDocs in place of SwaggerUI() for a
// single-page, three-panel doc layout instead of Swagger UI's
// try-it-out console.
func ReDoc() DocsUI { return redocUI{} }

func (redocUI) Assets() map[string]DocsAsset {
	return map[string]DocsAsset{
		"redoc.standalone.js": mustReadAsset(redocAssets, "redoc/redoc.standalone.js", "application/javascript; charset=utf-8"),
	}
}

func (redocUI) Index(specPath string) []byte {
	return []byte(`<!DOCTYPE html>
<html>
<head>
  <title>goninja API docs</title>
  <style>body { margin: 0; }</style>
</head>
<body>
  <redoc spec-url="` + specPath + `"></redoc>
  <script src="redoc.standalone.js"></script>
</body>
</html>
`)
}
