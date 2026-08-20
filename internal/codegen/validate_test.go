package codegen

import (
	"os"
	"strings"
	"testing"
)

func TestValidate_AcceptsAWellFormedModel(t *testing.T) {
	models := []Model{{
		Name:       "Book",
		SourceFile: "models/book.go",
		Fields: []Field{
			{Name: "ID", GoType: "string", Tags: []string{"list", "retrieve"}},
			{Name: "Title", GoType: "string", Tags: []string{"list", "create", "filter"}},
			{Name: "Author", GoType: "Author", Tags: []string{"retrieve"}},
			{Name: "Reviews", GoType: "[]Review", Tags: []string{"retrieve", "byid"}},
		},
	}}

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate([]Model{tt.model})
			if err == nil {
				t.Fatalf("expected an error mentioning %q, got none", tt.want)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error %q does not mention %q", err, tt.want)
			}
		})
	}
}

func TestValidate_ReportsEveryProblemAtOnce(t *testing.T) {
	// One run should tell the whole story rather than surfacing the next
	// problem only after the previous one is fixed.
	models := []Model{
		{Name: "Book", SourceFile: "models/book.go", Fields: []Field{
			{Name: "Author", GoType: "*Author", Tags: []string{"retrieve", "filter"}},
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
