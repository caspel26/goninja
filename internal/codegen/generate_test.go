package codegen

import (
	"fmt"
	"os"
	"os/exec"
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
// TestGenerate_TimeFilterCompiles is a regression test for a real bug
// found via examples/prototype/models/book.go's Book.CreatedAt: a
// `filter`-tagged time.Time field fell through parse<Model>Filters'
// exact-match branch to its int64 fallback, generating the invalid
// conversion time.Time(n), which doesn't compile. The fix special-cases
// time.Time to time.Parse(time.RFC3339, v) instead.
func TestGenerate_TimeFilterCompiles(t *testing.T) {
	models := []Model{
		{
			Name: "Widget",
			Fields: []Field{
				{Name: "ID", GoType: "string", JSONName: "id", Tags: []string{"list", "retrieve"}},
				{Name: "CreatedAt", GoType: "time.Time", JSONName: "created_at", Tags: []string{"list", "retrieve", "filter"}},
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

	if strings.Contains(got, "time.Time(n)") {
		t.Error("generated file still contains the invalid time.Time(n) conversion")
	}
	for _, want := range []string{
		`time.Parse(time.RFC3339, v)`,
		`f.CreatedAt = &parsed`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected generated file to contain %q, got:\n%s", want, got)
		}
	}
}

// TestGenerate_EveryScalarFilterTypeCompiles is the broader regression
// guard the time.Time bug (above) exposed a gap in: none of this file's
// other tests actually type-check the generated output — go/format.Source
// (generate.go) only formats syntax, so an invalid conversion like
// time.Time(n) passed silently until it hit a real consumer
// (examples/prototype's Book.CreatedAt). This covers every type
// scalarGoTypes (ir.go) recognizes as a filterable column in one pass, by
// actually compiling the generated output in a throwaway module against
// this repo's own goninja package (a local replace, not a tagged
// release — this is testing the engine against itself, not simulating an
// external consumer the way examples/prototype/ticketdesk-style testing
// does).
//
// Reasoning for why time.Time was the only type actually at risk, checked
// here rather than just asserted: parse<Model>Filters' exact-match branch
// (model.go.tmpl) is bool -> string -> float -> time.Time -> int64
// fallback. Every scalarGoTypes member other than time.Time is bool,
// string, a float, or an integer type — and Go's conversion rules
// guarantee int64(n) converts to any other integer type (int8..uint64,
// byte, rune) or float type without a compile error (only overflow/
// truncation at runtime, which is an existing, separate concern from this
// bug). time.Time was the only member with no defined conversion from
// int64, which is exactly what broke.
func TestGenerate_EveryScalarFilterTypeCompiles(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles a real temp module; skipped in -short")
	}

	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}

	scalarTypes := []string{
		"string", "bool",
		"int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64",
		"byte", "rune",
		"time.Time",
	}

	fields := []Field{
		{Name: "ID", GoType: "string", JSONName: "id", Tags: []string{"list", "retrieve"}},
	}
	var modelFieldsSrc strings.Builder
	modelFieldsSrc.WriteString("\tID string `json:\"id\"`\n")
	for i, gt := range scalarTypes {
		name := fmt.Sprintf("F%dField", i)
		json := fmt.Sprintf("f%d_field", i)
		fields = append(fields, Field{
			Name: name, GoType: gt, JSONName: json,
			Tags: []string{"list", "retrieve", "create", "update", "filter"},
		})
		modelFieldsSrc.WriteString(fmt.Sprintf("\t%s %s `json:%q`\n", name, gt, json))
	}
	models := []Model{{Name: "Everything", Fields: fields}}

	tmp := t.TempDir()
	modelsDir := filepath.Join(tmp, "models")
	apiDir := filepath.Join(tmp, "internal", "api")
	if err := os.MkdirAll(modelsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	modelsSrc := "package models\n\nimport \"time\"\n\nvar _ = time.Time{}\n\ntype Everything struct {\n" +
		modelFieldsSrc.String() + "}\n"
	if err := os.WriteFile(filepath.Join(modelsDir, "everything.go"), []byte(modelsSrc), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Generate(models, apiDir, "api", "compiletest/models", "models"); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	goMod := "module compiletest\n\ngo 1.25.0\n\nrequire github.com/caspel26/goninja v0.0.0\n\n" +
		"replace github.com/caspel26/goninja => " + repoRoot + "\n"
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatal(err)
	}

	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = tmp
	if out, err := tidy.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy: %v\n%s", err, out)
	}

	build := exec.Command("go", "build", "./...")
	build.Dir = tmp
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("generated code does not compile:\n%s", out)
	}
}

// TestGenerate_TimeFilterOnlyFieldImportsTime covers the specific edge
// case UsesTime's fix addresses: a time.Time field tagged only `filter`
// (no list/retrieve/create/update) still needs "time" imported, since
// its generated parse<Model>Filters branch calls time.Parse.
func TestGenerate_TimeFilterOnlyFieldImportsTime(t *testing.T) {
	models := []Model{
		{
			Name: "Widget",
			Fields: []Field{
				{Name: "ID", GoType: "string", JSONName: "id", Tags: []string{"list", "retrieve"}},
				{Name: "ArchivedAt", GoType: "time.Time", JSONName: "archived_at", Tags: []string{"filter"}},
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
	if !strings.Contains(got, "\"time\"") {
		t.Errorf("expected generated file to import \"time\" for a filter-only time.Time field, got:\n%s", got)
	}
}

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
