package codegen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGenerate_TwoModels is the Phase 1 exit criterion from
// goninja-implementation-plan.md: the engine must generalize beyond a
// single hand-tuned model. It also guards against shared-helper
// duplication (writeJSON/idFromPath) when multiple models land in the
// same generated package.
func TestGenerate_TwoModels(t *testing.T) {
	models := []Model{
		{
			Name: "Task",
			Fields: []Field{
				{Name: "ID", GoType: "int64", JSONName: "id", Tags: []string{"list", "retrieve"}},
				{Name: "Title", GoType: "string", JSONName: "title", Tags: []string{"list", "retrieve", "create"}},
			},
		},
		{
			Name: "Author",
			Fields: []Field{
				{Name: "ID", GoType: "int64", JSONName: "id", Tags: []string{"list", "retrieve"}},
				{Name: "Name", GoType: "string", JSONName: "name", Tags: []string{"list", "retrieve", "create"}},
				{Name: "Bio", GoType: "string", JSONName: "bio", Tags: []string{"retrieve", "create"}},
			},
		},
	}

	outDir := t.TempDir()
	if err := Generate(models, outDir, "api", "example.com/app/models", "models"); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	for _, name := range []string{"task_generated.go", "author_generated.go", "runtime_generated.go"} {
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			t.Errorf("expected %s to exist: %v", name, err)
		}
	}

	// writeJSON/idFromPath must be declared exactly once (in the runtime
	// file), never re-declared per model.
	for _, helper := range []string{"func writeJSON", "func idFromPath"} {
		count := 0
		for _, name := range []string{"task_generated.go", "author_generated.go", "runtime_generated.go"} {
			b, err := os.ReadFile(filepath.Join(outDir, name))
			if err != nil {
				t.Fatal(err)
			}
			count += strings.Count(string(b), helper)
		}
		if count != 1 {
			t.Errorf("expected %q declared exactly once across generated files, found %d", helper, count)
		}
	}
}

func TestGenerate_NoModels(t *testing.T) {
	if err := Generate(nil, t.TempDir(), "api", "example.com/app/models", "models"); err == nil {
		t.Fatal("expected error for empty model list, got nil")
	}
}
