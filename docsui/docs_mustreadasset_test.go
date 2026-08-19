package docsui

import (
	"embed"
	"testing"
)

func TestMustReadAsset_PanicsOnMissingFile(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("mustReadAsset: expected a panic for a missing embedded file")
		}
	}()

	var empty embed.FS
	mustReadAsset(empty, "does-not-exist.js", "application/javascript")
}
