package codegen

import (
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

func TestParseModels(t *testing.T) {
	dir := t.TempDir()
	src := `package models

type Task struct {
	ID    int64  ` + "`json:\"id\" goninja:\"list,retrieve\"`" + `
	Title string ` + "`json:\"title\" goninja:\"list,retrieve,create\"`" + `
	Internal string
}
`
	if err := os.WriteFile(filepath.Join(dir, "task.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	models, err := ParseModels(dir)
	if err != nil {
		t.Fatalf("ParseModels: %v", err)
	}
	if len(models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(models))
	}

	m := models[0]
	if m.Name != "Task" {
		t.Errorf("expected model name Task, got %s", m.Name)
	}
	// Internal has no goninja tag and must not appear.
	if len(m.Fields) != 2 {
		t.Fatalf("expected 2 tagged fields, got %d: %+v", len(m.Fields), m.Fields)
	}

	if got := len(m.ListFields()); got != 2 {
		t.Errorf("expected 2 list fields, got %d", got)
	}
	if got := len(m.CreateFields()); got != 1 {
		t.Errorf("expected 1 create field, got %d", got)
	}
	if m.NameLower() != "task" {
		t.Errorf("expected NameLower task, got %s", m.NameLower())
	}
}

func TestParseModels_FilterFieldsAndIDGoType(t *testing.T) {
	dir := t.TempDir()
	src := `package models

type Book struct {
	ID    string  ` + "`json:\"id\" goninja:\"list,retrieve\"`" + `
	Price float64 ` + "`json:\"price\" goninja:\"list,retrieve,create,update,filter\"`" + `
}
`
	if err := os.WriteFile(filepath.Join(dir, "book.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	models, err := ParseModels(dir)
	if err != nil {
		t.Fatalf("ParseModels: %v", err)
	}
	if len(models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(models))
	}

	m := models[0]
	if got := len(m.FilterFields()); got != 1 {
		t.Errorf("expected 1 filter field, got %d", got)
	}
	if got := m.IDGoType(); got != "string" {
		t.Errorf("expected IDGoType string, got %s", got)
	}
}

func TestParseModels_ExprTypeVariants(t *testing.T) {
	dir := t.TempDir()
	src := `package models

import "time"

type Author struct {
	ID string ` + "`json:\"id\" goninja:\"list,retrieve\"`" + `
}

type Book struct {
	ID        string     ` + "`json:\"id\" goninja:\"list,retrieve\"`" + `
	CreatedAt time.Time  ` + "`json:\"created_at\" goninja:\"list,retrieve\"`" + `
	Author    *Author    ` + "`json:\"author\" goninja:\"retrieve\"`" + `
	Tags      []string   ` + "`json:\"tags\" goninja:\"list,retrieve\"`" + `
}
`
	if err := os.WriteFile(filepath.Join(dir, "book.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	models, err := ParseModels(dir)
	if err != nil {
		t.Fatalf("ParseModels: %v", err)
	}

	var book Model
	for _, m := range models {
		if m.Name == "Book" {
			book = m
		}
	}
	if book.Name != "Book" {
		t.Fatalf("Book model not found among %+v", models)
	}

	types := map[string]string{}
	for _, f := range book.Fields {
		types[f.Name] = f.GoType
	}
	if types["CreatedAt"] != "time.Time" {
		t.Errorf("CreatedAt GoType = %q, want time.Time", types["CreatedAt"])
	}
	if types["Author"] != "*Author" {
		t.Errorf("Author GoType = %q, want *Author", types["Author"])
	}
	if types["Tags"] != "[]string" {
		t.Errorf("Tags GoType = %q, want []string", types["Tags"])
	}
}

func TestParseModels_UnsupportedFieldTypeErrors(t *testing.T) {
	dir := t.TempDir()
	src := `package models

type Weird struct {
	ID   string            ` + "`json:\"id\" goninja:\"list,retrieve\"`" + `
	Meta map[string]string ` + "`json:\"meta\" goninja:\"list,retrieve\"`" + `
}
`
	if err := os.WriteFile(filepath.Join(dir, "weird.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := ParseModels(dir); err == nil {
		t.Fatal("ParseModels: err = nil, want an error for an unsupported field type")
	}
}

func TestParseModels_UnsupportedTypeInsidePointerAndSlice(t *testing.T) {
	cases := []string{
		"*map[string]string",
		"[]map[string]string",
	}
	for _, fieldType := range cases {
		dir := t.TempDir()
		src := `package models

type Weird struct {
	ID   string      ` + "`json:\"id\" goninja:\"list,retrieve\"`" + `
	Meta ` + fieldType + ` ` + "`json:\"meta\" goninja:\"list,retrieve\"`" + `
}
`
		if err := os.WriteFile(filepath.Join(dir, "weird.go"), []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}

		if _, err := ParseModels(dir); err == nil {
			t.Errorf("ParseModels with field type %s: err = nil, want an error", fieldType)
		}
	}
}

func TestParseModels_UntaggedFieldsAndFieldsWithoutNamesAreSkipped(t *testing.T) {
	dir := t.TempDir()
	src := `package models

type Mixed struct {
	ID       string ` + "`json:\"id\" goninja:\"list,retrieve\"`" + `
	Untagged string
	Embedded
}

type Embedded struct {
	Value string
}
`
	if err := os.WriteFile(filepath.Join(dir, "mixed.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	models, err := ParseModels(dir)
	if err != nil {
		t.Fatalf("ParseModels: %v", err)
	}

	var mixed Model
	for _, m := range models {
		if m.Name == "Mixed" {
			mixed = m
		}
	}
	if len(mixed.Fields) != 1 || mixed.Fields[0].Name != "ID" {
		t.Errorf("expected only the ID field to be captured, got %+v", mixed.Fields)
	}
}

func TestParseModels_JSONNameFallbacksAndOmitempty(t *testing.T) {
	dir := t.TempDir()
	src := `package models

type Widget struct {
	ID       string ` + "`goninja:\"list,retrieve\"`" + `
	Hidden   string ` + "`json:\"-\" goninja:\"list\"`" + `
	Optional string ` + "`json:\"optional,omitempty\" goninja:\"list\"`" + `
}
`
	if err := os.WriteFile(filepath.Join(dir, "widget.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	models, err := ParseModels(dir)
	if err != nil {
		t.Fatalf("ParseModels: %v", err)
	}
	m := models[0]

	names := map[string]string{}
	for _, f := range m.Fields {
		names[f.Name] = f.JSONName
	}
	if names["ID"] != "iD" {
		t.Errorf(`ID JSONName = %q, want "iD" (no json tag -> lower() only lowercases the first rune)`, names["ID"])
	}
	if names["Hidden"] != "hidden" {
		t.Errorf(`Hidden JSONName = %q, want "hidden" (json:"-" -> lowercased field name)`, names["Hidden"])
	}
	if names["Optional"] != "optional" {
		t.Errorf(`Optional JSONName = %q, want "optional" (comma-suffixed json tag trimmed)`, names["Optional"])
	}
}

func TestParseModels_DerivesDBColumnsIndependentlyOfJSONNames(t *testing.T) {
	dir := t.TempDir()
	src := `package models

import "time"

type Widget struct {
	ID        string    ` + "`goninja:\"list,retrieve\"`" + `
	CreatedAt time.Time ` + "`json:\"createdAt\" goninja:\"list,retrieve,filter\"`" + `
	Label     string    ` + "`json:\"displayName\" gorm:\"column:display_name\" goninja:\"list,retrieve,filter\"`" + `
}
`
	if err := os.WriteFile(filepath.Join(dir, "widget.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	models, err := ParseModels(dir)
	if err != nil {
		t.Fatalf("ParseModels: %v", err)
	}
	fields := map[string]Field{}
	for _, f := range models[0].Fields {
		fields[f.Name] = f
	}

	for name, want := range map[string]string{
		"ID":        "id",
		"CreatedAt": "created_at",
		"Label":     "display_name",
	} {
		if got := fields[name].DBColumn; got != want {
			t.Errorf("%s DBColumn = %q, want %q", name, got, want)
		}
	}
	if got := fields["CreatedAt"].JSONName; got != "createdAt" {
		t.Errorf("CreatedAt JSONName = %q, want createdAt", got)
	}
}

func TestParseModels_NonStructTypeDeclsAreSkipped(t *testing.T) {
	dir := t.TempDir()
	src := `package models

type ID = string

type Task struct {
	ID string ` + "`json:\"id\" goninja:\"list,retrieve\"`" + `
}

`
	if err := os.WriteFile(filepath.Join(dir, "task.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	models, err := ParseModels(dir)
	if err != nil {
		t.Fatalf("ParseModels: %v", err)
	}
	if len(models) != 1 || models[0].Name != "Task" {
		t.Errorf("expected only Task to be captured, got %+v", models)
	}
}

func TestParseModels_ResolvesNamedScalars(t *testing.T) {
	dir := t.TempDir()
	src := `package models

type Status string
type Cents int64
type Price Cents

type Book struct {
	ID     string ` + "`json:\"id\" goninja:\"list,retrieve\"`" + `
	Status Status ` + "`json:\"status\" goninja:\"list,retrieve,create,update,filter\"`" + `
	Price  Price  ` + "`json:\"price\" goninja:\"list,retrieve,create,update,filter\"`" + `
}
`
	if err := os.WriteFile(filepath.Join(dir, "book.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	models, err := ParseModels(dir)
	if err != nil {
		t.Fatalf("ParseModels: %v", err)
	}
	fields := map[string]Field{}
	for _, f := range models[0].Fields {
		fields[f.Name] = f
	}
	if got := fields["Status"].UnderlyingGoType; got != "string" {
		t.Errorf("Status UnderlyingGoType = %q, want string", got)
	}
	if got := fields["Price"].UnderlyingGoType; got != "int64" {
		t.Errorf("Price UnderlyingGoType = %q, want int64", got)
	}
	if fields["Status"].IsRelation() || fields["Price"].IsRelation() {
		t.Errorf("named scalars must not be relations: %+v", fields)
	}
}

func TestParseModels_TaggedFieldWithoutGoninjaTagIsSkipped(t *testing.T) {
	dir := t.TempDir()
	src := `package models

type Task struct {
	ID          string ` + "`json:\"id\" goninja:\"list,retrieve\"`" + `
	UntaggedButHasJSON string ` + "`json:\"other\"`" + `
}
`
	if err := os.WriteFile(filepath.Join(dir, "task.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	models, err := ParseModels(dir)
	if err != nil {
		t.Fatalf("ParseModels: %v", err)
	}
	m := models[0]
	if len(m.Fields) != 1 || m.Fields[0].Name != "ID" {
		t.Errorf("expected only the ID field to be captured, got %+v", m.Fields)
	}
}

func TestParseModels_InvalidGoSyntaxErrors(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "broken.go"), []byte("package models\nfunc {{{"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := ParseModels(dir); err == nil {
		t.Fatal("ParseModels: err = nil, want a parse error")
	}
}

func TestParseModels_NoTaggedStructs(t *testing.T) {
	dir := t.TempDir()
	src := `package models

type Plain struct {
	Name string
}
`
	if err := os.WriteFile(filepath.Join(dir, "plain.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	models, err := ParseModels(dir)
	if err != nil {
		t.Fatalf("ParseModels: %v", err)
	}
	if len(models) != 0 {
		t.Fatalf("expected 0 models, got %d", len(models))
	}
}

func TestExprToString_SelectorWithUnsupportedQualifierErrors(t *testing.T) {
	// A SelectorExpr whose X is itself unsupported (never produced by real
	// Go source for a field type, but exprToString must still propagate
	// the error rather than only handling it at the top level).
	expr := &ast.SelectorExpr{
		X:   &ast.MapType{Key: &ast.Ident{Name: "string"}, Value: &ast.Ident{Name: "string"}},
		Sel: &ast.Ident{Name: "Field"},
	}

	if _, err := exprToString(expr); err == nil {
		t.Fatal("exprToString: err = nil, want an error propagated from the unsupported qualifier")
	}
}

func TestParseGenDeclModels_SkipsNonTypeSpec(t *testing.T) {
	// A *ast.GenDecl with token.TYPE can only legally contain *ast.TypeSpec
	// specs from real source, so this exercises the defensive skip branch
	// directly rather than via ParseModels.
	gen := &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"fmt"`}},
		},
	}

	models, err := parseGenDeclModels("models.go", gen)
	if err != nil {
		t.Fatalf("parseGenDeclModels: %v", err)
	}
	if len(models) != 0 {
		t.Fatalf("expected 0 models, got %d", len(models))
	}
}
