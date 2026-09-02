package codegen

import (
	"os"
	"strings"
	"testing"
)

func TestValidate_AcceptsAWellFormedModel(t *testing.T) {
	models := []Model{
		{
			Name:       "Book",
			SourceFile: "models/book.go",
			Fields: []Field{
				{Name: "ID", GoType: "string", JSONName: "id", Tags: []string{"list", "retrieve"}},
				{Name: "Title", GoType: "string", JSONName: "title", Tags: []string{"list", "create", "filter"}},
				{Name: "Author", GoType: "Author", JSONName: "author", Tags: []string{"retrieve"}},
				{Name: "Reviews", GoType: "[]Review", JSONName: "reviews", Tags: []string{"retrieve", "byid"}},
			},
		},
		{Name: "Author", Fields: []Field{{Name: "ID", GoType: "string", Tags: []string{"list", "retrieve"}}}},
		{Name: "Review", Fields: []Field{{Name: "ID", GoType: "string", Tags: []string{"list", "retrieve"}}}},
	}

	if err := Validate(models); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestValidate_RejectsBadModels(t *testing.T) {
	tests := []struct {
		name  string
		model Model
		// want is a substring the error message has to contain, chosen to be
		// the part that tells the user what to do about it.
		want string
	}{
		{
			name: "no ID field",
			model: Model{Name: "Book", Fields: []Field{
				{Name: "Title", GoType: "string", Tags: []string{"list"}},
			}},
			want: "no goninja-tagged field named ID",
		},
		{
			name: "ID of an unusable type",
			model: Model{Name: "Book", Fields: []Field{
				{Name: "ID", GoType: "int", Tags: []string{"list"}},
			}},
			want: "ID is int, which the generator cannot use",
		},
		{
			name: "byid on a scalar field",
			model: Model{Name: "Book", Fields: []Field{
				{Name: "ID", GoType: "int64", Tags: []string{"list"}},
				{Name: "Title", GoType: "string", Tags: []string{"retrieve", "byid"}},
			}},
			want: `"byid" modifier only applies to a relation field`,
		},
		{
			name: "pointer belongs-to relation",
			model: Model{Name: "Book", Fields: []Field{
				{Name: "ID", GoType: "int64", Tags: []string{"list"}},
				{Name: "Author", GoType: "*Author", Tags: []string{"retrieve"}},
			}},
			want: "pointers are not supported",
		},
		{
			name: "slice of pointers relation",
			model: Model{Name: "Author", Fields: []Field{
				{Name: "ID", GoType: "int64", Tags: []string{"list"}},
				{Name: "Books", GoType: "[]*Book", Tags: []string{"retrieve"}},
			}},
			want: "pointers are not supported",
		},
		{
			name: "filter on a relation field",
			model: Model{Name: "Book", Fields: []Field{
				{Name: "ID", GoType: "int64", Tags: []string{"list"}},
				{Name: "Author", GoType: "Author", Tags: []string{"retrieve", "filter"}},
			}},
			want: "cannot be filtered",
		},
		{
			name: "named scalar type",
			model: Model{Name: "Book", Fields: []Field{
				{Name: "ID", GoType: "int64", Tags: []string{"list"}},
				{Name: "Status", GoType: "Status", Tags: []string{"list", "filter"}},
			}},
			want: "neither a supported scalar nor an annotated goninja model relation",
		},
		{
			name: "nullable scalar type",
			model: Model{Name: "Book", Fields: []Field{
				{Name: "ID", GoType: "int64", Tags: []string{"list"}},
				{Name: "Title", GoType: "*string", Tags: []string{"retrieve"}},
			}},
			want: "nullable and collection scalar types",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(append([]Model{tt.model}, validRelationModels()...))
			if err == nil {
				t.Fatalf("expected an error mentioning %q, got none", tt.want)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error %q does not mention %q", err, tt.want)
			}
		})
	}
}

func validRelationModels() []Model {
	return []Model{
		{Name: "Author", Fields: []Field{{Name: "ID", GoType: "int64", Tags: []string{"list"}}}},
		{Name: "Book", Fields: []Field{{Name: "ID", GoType: "int64", Tags: []string{"list"}}}},
		{Name: "Review", Fields: []Field{{Name: "ID", GoType: "int64", Tags: []string{"list"}}}},
	}
}

func TestValidate_RejectsRelationsOutsideTheGenerationSet(t *testing.T) {
	err := Validate([]Model{{Name: "Book", Fields: []Field{
		{Name: "ID", GoType: "int64", Tags: []string{"list"}},
		{Name: "Author", GoType: "Author", Tags: []string{"retrieve"}},
	}}})
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "annotated goninja model relation") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidate_RejectsMalformedTags(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want string
	}{
		{name: "unknown", tags: []string{"list", "export"}, want: `unknown goninja tag "export"`},
		{name: "duplicate", tags: []string{"list", "list"}, want: `duplicate goninja tag "list"`},
		{name: "byid without retrieve", tags: []string{"byid"}, want: `"byid" modifier requires the "retrieve" tag`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate([]Model{{Name: "Book", Fields: []Field{
				{Name: "ID", GoType: "int64", Tags: []string{"list"}},
				{Name: "Title", GoType: "string", Tags: tt.tags},
			}}})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestValidate_RejectsGeneratedNameCollisions(t *testing.T) {
	tests := []struct {
		name   string
		fields []Field
		want   string
	}{
		{
			name: "list JSON names",
			fields: []Field{
				{Name: "ID", GoType: "int64", Tags: []string{"list"}},
				{Name: "Title", GoType: "string", JSONName: "name", Tags: []string{"list"}},
				{Name: "Name", GoType: "string", JSONName: "name", Tags: []string{"list"}},
			},
			want: "share JSON name \"name\" in the list schema",
		},
		{
			name: "filter range query parameter",
			fields: []Field{
				{Name: "ID", GoType: "int64", Tags: []string{"list"}},
				{Name: "Price", GoType: "int64", JSONName: "price", Tags: []string{"filter"}},
				{Name: "PriceMin", GoType: "int64", JSONName: "price_min", Tags: []string{"filter"}},
			},
			want: "both use query parameter \"price_min\"",
		},
		{
			name: "database columns",
			fields: []Field{
				{Name: "ID", GoType: "int64", Tags: []string{"list"}},
				{Name: "Title", GoType: "string", DBColumn: "label", Tags: []string{"list"}},
				{Name: "Name", GoType: "string", DBColumn: "label", Tags: []string{"retrieve"}},
			},
			want: "use the same database column \"label\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate([]Model{{Name: "Book", Fields: tt.fields}})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestValidate_RejectsModelOutputCollisions(t *testing.T) {
	err := Validate([]Model{
		{Name: "Book", Fields: []Field{{Name: "ID", GoType: "int64", Tags: []string{"list"}}}},
		{Name: "book", Fields: []Field{{Name: "ID", GoType: "int64", Tags: []string{"list"}}}},
	})
	if err == nil || !strings.Contains(err.Error(), "generate the same file and default route name") {
		t.Fatalf("error = %v, want model output collision", err)
	}
}

func TestValidate_ReportsEveryProblemAtOnce(t *testing.T) {
	// One run should tell the whole story rather than surfacing the next
	// problem only after the previous one is fixed.
	models := []Model{
		{Name: "Book", SourceFile: "models/book.go", Fields: []Field{
			{Name: "Author", GoType: "*Author", Tags: []string{"retrieve", "filter"}},
		}},
		{Name: "Author", SourceFile: "models/author.go", Fields: []Field{
			{Name: "ID", GoType: "int64", Tags: []string{"list"}},
		}},
		{Name: "Tag", SourceFile: "models/tag.go", Fields: []Field{
			{Name: "ID", GoType: "uint", Tags: []string{"list"}},
		}},
	}

	err := Validate(models)
	if err == nil {
		t.Fatal("expected an error")
	}

	for _, want := range []string{
		"no goninja-tagged field named ID", // Book has none
		"pointers are not supported",       // Book.Author
		"cannot be filtered",               // Book.Author again
		"ID is uint",                       // Tag.ID
		"models/book.go",                   // each problem names its file
		"models/tag.go",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error is missing %q:\n%v", want, err)
		}
	}
}

func TestValidate_NamesTheModelWhenTheSourceFileIsUnknown(t *testing.T) {
	// A hand-built Model (as in these tests, or a caller using the package
	// directly) has no SourceFile, and must still be identifiable.
	err := Validate([]Model{{Name: "Orphan"}})
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "Orphan:") {
		t.Fatalf("error does not name the model: %v", err)
	}
}

func TestGenerate_RejectsAnInvalidModelWithoutWritingFiles(t *testing.T) {
	out := t.TempDir()
	models := []Model{{Name: "Book", Fields: []Field{
		{Name: "Title", GoType: "string", Tags: []string{"list"}},
	}}}

	err := Generate(models, out, "api", "example.com/m/models", "models")
	if err == nil {
		t.Fatal("expected Generate to reject the model")
	}
	if !strings.Contains(err.Error(), "no goninja-tagged field named ID") {
		t.Fatalf("unexpected error: %v", err)
	}

	entries, readErr := os.ReadDir(out)
	if readErr != nil {
		t.Fatalf("reading %s: %v", out, readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("expected no files written, got %d", len(entries))
	}
}

func TestGenerate_RejectsNamedScalarsWithoutWritingFiles(t *testing.T) {
	out := t.TempDir()
	models := []Model{{Name: "Book", Fields: []Field{
		{Name: "ID", GoType: "int64", Tags: []string{"list"}},
		{Name: "Status", GoType: "Status", Tags: []string{"list", "filter"}},
	}}}

	err := Generate(models, out, "api", "example.com/m/models", "models")
	if err == nil {
		t.Fatal("expected Generate to reject a named scalar")
	}
	if !strings.Contains(err.Error(), "named scalar and external types are not supported yet") {
		t.Fatalf("unexpected error: %v", err)
	}

	entries, readErr := os.ReadDir(out)
	if readErr != nil {
		t.Fatalf("reading %s: %v", out, readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("expected no files written, got %d", len(entries))
	}
}
