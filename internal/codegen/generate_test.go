package codegen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGenerate_TwoModels is the Phase 1 exit criterion from
// goninja-implementation-plan.md: the engine must generalize beyond a
// single hand-tuned model. Shared handler helpers (JSON responses, error
// mapping, validation) live in the goninja package rather than being
// generated per model, so there's no shared-file duplication risk to guard
// here beyond each model's own file existing.
func TestGenerate_TwoModels(t *testing.T) {
	models := []Model{
		{
			Name: "Task",
			Fields: []Field{
				{Name: "ID", GoType: "int64", JSONName: "id", Tags: []string{"list", "retrieve"}},
				{Name: "Title", GoType: "string", JSONName: "title", Tags: []string{"list", "retrieve", "create"}, ValidateTag: "required"},
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

	for _, name := range []string{"task_generated.go", "author_generated.go"} {
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			t.Errorf("expected %s to exist: %v", name, err)
		}
	}

	b, err := os.ReadFile(filepath.Join(outDir, "task_generated.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `validate:"required"`) {
		t.Errorf("expected Task's validate tag to appear in generated schema, got:\n%s", b)
	}
}

func TestGenerate_NoModels(t *testing.T) {
	if err := Generate(nil, t.TempDir(), "api", "example.com/app/models", "models"); err == nil {
		t.Fatal("expected error for empty model list, got nil")
	}
}

// TestGenerate_FiltersAndUUID is the Phase 4 exit criterion from
// goninja-implementation-plan.md: a `filter`-tagged field must produce a
// working Filters struct/query, and a string-typed ID field (a UUID
// primary key, per Model.IDGoType) must generate without falling back to
// the historical int64 assumption.
func TestGenerate_FiltersAndUUID(t *testing.T) {
	models := []Model{
		{
			Name: "Widget",
			Fields: []Field{
				{Name: "ID", GoType: "string", JSONName: "id", Tags: []string{"list", "retrieve"}},
				{Name: "Name", GoType: "string", JSONName: "name", Tags: []string{"list", "retrieve", "create", "update", "filter"}},
				{Name: "Price", GoType: "float64", JSONName: "price", Tags: []string{"list", "retrieve", "create", "update", "filter"}},
				{Name: "Active", GoType: "bool", JSONName: "active", Tags: []string{"list", "retrieve", "create", "update", "filter"}},
			},
		},
	}

	outDir := t.TempDir()
	if err := Generate(models, outDir, "api", "example.com/app/models", "models"); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	b, err := os.ReadFile(filepath.Join(outDir, "widget_generated.go"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)

	for _, want := range []string{
		"type WidgetFilters struct",
		"PriceMin *float64",
		"PriceMax *float64",
		"func (r *WidgetResource) List(ctx context.Context, f WidgetFilters) ([]WidgetList, int64, error)",
		"func (r *WidgetResource) Retrieve(ctx context.Context, id string) (*WidgetRetrieve, error)",
		"m.ID = goninja.NewUUID()",
		"goninja.ListEnvelope[WidgetList]",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected generated file to contain %q, got:\n%s", want, got)
		}
	}
}
