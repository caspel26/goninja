package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRun_NoArgsReturnsError(t *testing.T) {
	if err := run(nil); err == nil {
		t.Fatal("run(nil): err = nil, want an error for a missing command")
	}
}

func TestRun_UnknownCommandReturnsError(t *testing.T) {
	if err := run([]string{"bogus"}); err == nil {
		t.Fatal(`run(["bogus"]): err = nil, want an error for an unknown command`)
	}
}

func TestRun_HelpReturnsNil(t *testing.T) {
	for _, arg := range []string{"-h", "--help", "help"} {
		if err := run([]string{arg}); err != nil {
			t.Errorf("run([%q]): err = %v, want nil", arg, err)
		}
	}
}

func TestRunGenerate_MissingModelsImportReturnsError(t *testing.T) {
	err := run([]string{"generate", "-models", t.TempDir(), "-out", t.TempDir()})
	if err == nil {
		t.Fatal("runGenerate: err = nil, want an error when -models-import is missing")
	}
}

func TestRunGenerate_BadFlagReturnsError(t *testing.T) {
	err := run([]string{"generate", "-not-a-flag"})
	if err == nil {
		t.Fatal("runGenerate: err = nil, want a flag-parsing error")
	}
}

func TestRunGenerate_GeneratesWithoutWatch(t *testing.T) {
	modelsDir := t.TempDir()
	outDir := t.TempDir()
	src := "package models\n\ntype Item struct {\n" +
		"\tID int64 `gorm:\"primaryKey\" goninja:\"list,retrieve\"`\n" +
		"}\n"
	if err := os.WriteFile(filepath.Join(modelsDir, "item.go"), []byte(src), 0o644); err != nil {
		t.Fatalf("writing model file: %v", err)
	}

	err := run([]string{
		"generate",
		"-models", modelsDir,
		"-out", outDir,
		"-models-import", "example.com/models",
	})
	if err != nil {
		t.Fatalf("runGenerate: unexpected error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "item_generated.go")); err != nil {
		t.Fatalf("expected generated file: %v", err)
	}
}
