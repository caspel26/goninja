package docsui

import "embed"

// swaggerUIAssets vendors the parts of swagger-ui-dist SwaggerUI actually
// serves (bundle JS, standalone preset JS, CSS, favicon) so /docs works
// fully offline — no external CDN. The
// package's own index.html/swagger-initializer.js are not vendored:
// swaggerUI.Index writes its own, small enough to inline, so it can point
// swagger-ui at this API's own /openapi.json path. See
// swagger-ui/LICENSE for swagger-ui's license.
//
//go:embed swagger-ui/swagger-ui-bundle.js swagger-ui/swagger-ui-standalone-preset.js swagger-ui/swagger-ui.css swagger-ui/favicon-32x32.png
var swaggerUIAssets embed.FS

type swaggerUI struct{}

// SwaggerUI is a DocsUI backed by Swagger UI, vendored under
// swagger-ui and embedded at build time — the default MountDocs falls
// back to when no DocsUI is given.
func SwaggerUI() DocsUI { return swaggerUI{} }

func (swaggerUI) Assets() map[string]DocsAsset {
	return map[string]DocsAsset{
		"swagger-ui-bundle.js":            mustReadAsset(swaggerUIAssets, "swagger-ui/swagger-ui-bundle.js", "application/javascript; charset=utf-8"),
		"swagger-ui-standalone-preset.js": mustReadAsset(swaggerUIAssets, "swagger-ui/swagger-ui-standalone-preset.js", "application/javascript; charset=utf-8"),
		"swagger-ui.css":                  mustReadAsset(swaggerUIAssets, "swagger-ui/swagger-ui.css", "text/css; charset=utf-8"),
		"favicon-32x32.png":               mustReadAsset(swaggerUIAssets, "swagger-ui/favicon-32x32.png", "image/png"),
	}
}

func (swaggerUI) Index(specPath string) []byte {
	return []byte(`<!DOCTYPE html>
<html>
<head>
  <title>goninja API docs</title>
  <link rel="stylesheet" href="swagger-ui.css">
  <link rel="icon" href="favicon-32x32.png">
  <style>body { margin: 0; }</style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="swagger-ui-bundle.js"></script>
  <script src="swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function() {
      window.ui = SwaggerUIBundle({
        url: "` + specPath + `",
        dom_id: "#swagger-ui",
        presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
        layout: "StandaloneLayout"
      });
    };
  </script>
</body>
</html>
`)
}
