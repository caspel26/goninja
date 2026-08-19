package docsui

import (
	"strings"
	"testing"
)

func TestSwaggerUI_IndexAndAssets(t *testing.T) {
	ui := SwaggerUI()

	index := string(ui.Index("/docs/openapi.json"))
	if !strings.Contains(index, "/docs/openapi.json") {
		t.Errorf("Index() does not reference specPath: %s", index)
	}
	if !strings.Contains(index, "swagger-ui-bundle.js") {
		t.Errorf("Index() does not reference swagger-ui-bundle.js: %s", index)
	}

	assets := ui.Assets()
	want := map[string]string{
		"swagger-ui-bundle.js":            "application/javascript; charset=utf-8",
		"swagger-ui-standalone-preset.js": "application/javascript; charset=utf-8",
		"swagger-ui.css":                  "text/css; charset=utf-8",
		"favicon-32x32.png":               "image/png",
	}
	for name, contentType := range want {
		asset, ok := assets[name]
		if !ok {
			t.Fatalf("Assets()[%q] missing", name)
		}
		if len(asset.Data) == 0 {
			t.Errorf("%s asset is empty", name)
		}
		if asset.ContentType != contentType {
			t.Errorf("%s ContentType = %q, want %q", name, asset.ContentType, contentType)
		}
	}
}
