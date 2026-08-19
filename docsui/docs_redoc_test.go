package docsui

import (
	"strings"
	"testing"
)

func TestReDoc_IndexAndAssets(t *testing.T) {
	ui := ReDoc()

	index := string(ui.Index("/docs/openapi.json"))
	if !strings.Contains(index, "/docs/openapi.json") {
		t.Errorf("Index() does not reference specPath: %s", index)
	}
	if !strings.Contains(index, "redoc.standalone.js") {
		t.Errorf("Index() does not reference redoc.standalone.js: %s", index)
	}

	assets := ui.Assets()
	asset, ok := assets["redoc.standalone.js"]
	if !ok {
		t.Fatal(`Assets()["redoc.standalone.js"] missing`)
	}
	if len(asset.Data) == 0 {
		t.Error("redoc.standalone.js asset is empty")
	}
	if asset.ContentType != "application/javascript; charset=utf-8" {
		t.Errorf("ContentType = %q, want application/javascript; charset=utf-8", asset.ContentType)
	}
}
