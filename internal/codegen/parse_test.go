package codegen

import (
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
