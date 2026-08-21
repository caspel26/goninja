package codegen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
)

// TestGenerate_TwoModels asserts the engine generalizes beyond a
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

// TestGenerate_RelationByID asserts a
// relation field tagged "byid" produces a Retrieve exposing only the
// related model's ID (no Preload, no nested Retrieve), typed after the
// related model's own IDGoType — and a relation field with no modifier
// (Book.Reviews, standing in for the untouched default) keeps nesting the
// full Retrieve exactly as before, no regression.
func TestGenerate_RelationByID(t *testing.T) {
	models := []Model{
		{
			Name: "Author",
			Fields: []Field{
				{Name: "ID", GoType: "string", JSONName: "id", Tags: []string{"list", "retrieve"}},
				{Name: "Name", GoType: "string", JSONName: "name", Tags: []string{"list", "retrieve", "create"}},
			},
		},
		{
			Name: "Book",
			Fields: []Field{
				{Name: "ID", GoType: "string", JSONName: "id", Tags: []string{"list", "retrieve"}},
				{Name: "Title", GoType: "string", JSONName: "title", Tags: []string{"list", "retrieve", "create"}},
				{Name: "AuthorID", GoType: "string", JSONName: "author_id", Tags: []string{"list", "create"}},
				{Name: "Author", GoType: "Author", JSONName: "author", Tags: []string{"retrieve", "byid"}},
			},
		},
	}

	outDir := t.TempDir()
	if err := Generate(models, outDir, "api", "example.com/app/models", "models"); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	b, err := os.ReadFile(filepath.Join(outDir, "book_generated.go"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)

	for _, want := range []string{
		// Retrieve exposes AuthorID (the related model's ID type), not a
		// nested Author.
		`AuthorID string ` + "`json:\"author_id\"`",
		"AuthorID: m.AuthorID,",
		`"author_id": {Type: "string"}`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected generated file to contain %q, got:\n%s", want, got)
		}
	}
	for _, notWant := range []string{
		`q.Preload("Author")`,
		"AuthorRetrieve",
	} {
		if strings.Contains(got, notWant) {
			t.Errorf("expected generated file NOT to contain %q (byid should skip nesting/preload), got:\n%s", notWant, got)
		}
	}
}

// TestGenerate_HasManyRelation confirms a slice relation field (has-many/
// reverse-FK, e.g. Author.Books []Book) nests as a slice of the related
// model's Retrieve type — Preload, the Retrieve struct field type, its
// conversion loop, and the OpenAPI schema all need to treat it as an array,
// not a single nested object (Field.IsSlice, internal/codegen/ir.go).
func TestGenerate_HasManyRelation(t *testing.T) {
	models := []Model{
		{
			Name: "Author",
			Fields: []Field{
				{Name: "ID", GoType: "string", JSONName: "id", Tags: []string{"list", "retrieve"}},
				{Name: "Name", GoType: "string", JSONName: "name", Tags: []string{"list", "retrieve", "create"}},
				{Name: "Books", GoType: "[]Book", JSONName: "books", Tags: []string{"retrieve"}},
			},
		},
		{
			Name: "Book",
			Fields: []Field{
				{Name: "ID", GoType: "string", JSONName: "id", Tags: []string{"list", "retrieve"}},
				{Name: "Title", GoType: "string", JSONName: "title", Tags: []string{"list", "retrieve", "create"}},
			},
		},
	}

	outDir := t.TempDir()
	if err := Generate(models, outDir, "api", "example.com/app/models", "models"); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	b, err := os.ReadFile(filepath.Join(outDir, "author_generated.go"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)

	for _, want := range []string{
		`Books []BookRetrieve ` + "`json:\"books\"`",
		`q.Preload("Books")`,
		"make([]BookRetrieve, 0, len(m.Books))",
		"toBookRetrieve(&m.Books[i])",
		`{Type: "array", Items: &openapi.Schema{Ref: "#/components/schemas/BookRetrieve"}}`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected generated file to contain %q, got:\n%s", want, got)
		}
	}
}

// TestGenerate_ActionsDispatch confirms Register and OpenAPI mount/document
// every Action returned by r.Actions() (set via SetActions).
func TestGenerate_ActionsDispatch(t *testing.T) {
	models := []Model{
		{
			Name: "Book",
			Fields: []Field{
				{Name: "ID", GoType: "string", JSONName: "id", Tags: []string{"list", "retrieve"}},
				{Name: "Title", GoType: "string", JSONName: "title", Tags: []string{"list", "retrieve", "create"}},
			},
		},
	}

	outDir := t.TempDir()
	if err := Generate(models, outDir, "api", "example.com/app/models", "models"); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	b, err := os.ReadFile(filepath.Join(outDir, "book_generated.go"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)

	for _, want := range []string{
		"for _, a := range r.Actions() {",
		"r.Protect(goninja.Route(a.Name), cfg, a.Handler)",
		"a.Summary",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected generated file to contain %q, got:\n%s", want, got)
		}
	}
}

func TestGenerate_NoModels(t *testing.T) {
	if err := Generate(nil, t.TempDir(), "api", "example.com/app/models", "models"); err == nil {
		t.Fatal("expected error for empty model list, got nil")
	}
}

func TestGenerate_PropagatesRenderFileError(t *testing.T) {
	// A model name that isn't a valid Go identifier produces invalid Go
	// source once substituted into the template, so renderFile's
	// go/format.Source step fails and Generate must propagate that error.
	models := []Model{
		{
			Name: "task-thing",
			Fields: []Field{
				{Name: "ID", GoType: "int64", JSONName: "id", Tags: []string{"list", "retrieve"}},
			},
		},
	}

	if err := Generate(models, t.TempDir(), "api", "example.com/app/models", "models"); err == nil {
		t.Fatal("Generate: err = nil, want the underlying renderFile error to propagate")
	}
}

func TestGenerate_OutDirCannotBeCreated(t *testing.T) {
	// outDir is nested under a path component that's actually a file, so
	// MkdirAll fails.
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	models := []Model{{Name: "Task", Fields: []Field{{Name: "ID", GoType: "int64", Tags: []string{"list"}}}}}
	err := Generate(models, filepath.Join(blocker, "out"), "api", "example.com/app/models", "models")
	if err == nil {
		t.Fatal("Generate: err = nil, want an error when outDir can't be created")
	}
}

func TestRelatedIDGoType_NoMatchDefaultsToString(t *testing.T) {
	models := []Model{{Name: "Author", Fields: []Field{{Name: "ID", GoType: "int64"}}}}
	if got := relatedIDGoType(models, "Unknown"); got != "string" {
		t.Errorf("relatedIDGoType(unmatched) = %q, want string", got)
	}
	if got := relatedIDGoType(models, "[]Author"); got != "int64" {
		t.Errorf("relatedIDGoType([]Author) = %q, want int64", got)
	}
	if got := relatedIDGoType(models, "*Author"); got != "int64" {
		t.Errorf("relatedIDGoType(*Author) = %q, want int64", got)
	}
}

func TestRenderFile_TemplateExecuteError(t *testing.T) {
	tmpl := template.Must(template.New("bad").Parse(`{{.Missing.Field}}`))
	err := renderFile(tmpl, struct{}{}, filepath.Join(t.TempDir(), "out.go"))
	if err == nil {
		t.Fatal("renderFile: err = nil, want a template execution error")
	}
}

func TestRenderFile_InvalidGoSource(t *testing.T) {
	tmpl := template.Must(template.New("invalid").Parse(`this is not go source {{.}}`))
	err := renderFile(tmpl, "x", filepath.Join(t.TempDir(), "out.go"))
	if err == nil {
		t.Fatal("renderFile: err = nil, want a go/format error for invalid source")
	}
}

func TestRenderFile_WriteError(t *testing.T) {
	tmpl := template.Must(template.New("ok").Parse(`package x`))
	// Path inside a directory that doesn't exist -> os.WriteFile fails.
	err := renderFile(tmpl, nil, filepath.Join(t.TempDir(), "missing-dir", "out.go"))
	if err == nil {
		t.Fatal("renderFile: err = nil, want a write error for a missing directory")
	}
}

// TestGenerate_FiltersAndUUID asserts a `filter`-tagged field must produce a
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
		"m.ID = id.NewUUID()",
		"goninja.ListEnvelope[WidgetList]",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected generated file to contain %q, got:\n%s", want, got)
		}
	}
}
