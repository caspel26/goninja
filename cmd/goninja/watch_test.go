package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestShouldRegenerate(t *testing.T) {
	tests := []struct {
		name string
		op   fsnotify.Op
		file string
		want bool
	}{
		{"write to a .go file regenerates", fsnotify.Write, "model.go", true},
		{"create of a .go file regenerates", fsnotify.Create, "model.go", true},
		{"rename into place regenerates", fsnotify.Rename, "model.go", true},
		{"chmod is ignored", fsnotify.Chmod, "model.go", false},
		{"remove is ignored", fsnotify.Remove, "model.go", false},
		{"write to a non-.go file is ignored", fsnotify.Write, "README.md", false},
		{"write to a dotfile with no extension is ignored", fsnotify.Write, ".gitignore", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := fsnotify.Event{Name: tt.file, Op: tt.op}
			if got := shouldRegenerate(event); got != tt.want {
				t.Errorf("shouldRegenerate(%+v) = %v, want %v", event, got, tt.want)
			}
		})
	}
}

func TestWatchAndRegenerate_MissingDirReturnsError(t *testing.T) {
	gen := generator{outDir: t.TempDir(), pkg: "api", modelsImport: "example.com/models", modelsPkg: "models"}
	err := watchAndRegenerate(context.Background(), filepath.Join(t.TempDir(), "does-not-exist"), gen)
	if err == nil {
		t.Fatal("watchAndRegenerate: err = nil, want an error for a missing models directory")
	}
}

func TestWatchAndRegenerate_ContextCanceledReturnsNil(t *testing.T) {
	modelsDir := t.TempDir()
	gen := generator{outDir: t.TempDir(), pkg: "api", modelsImport: "example.com/models", modelsPkg: "models"}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	errCh := make(chan error, 1)
	go func() { errCh <- watchAndRegenerate(ctx, modelsDir, gen) }()

	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("watchAndRegenerate: err = %v, want nil on context cancellation", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("watchAndRegenerate: did not return after context was canceled")
	}
}

func TestWatchAndRegenerate_RegeneratesOnFileWrite(t *testing.T) {
	modelsDir := t.TempDir()
	outDir := t.TempDir()
	modelFile := filepath.Join(modelsDir, "item.go")

	writeModel := func(field string) {
		src := "package models\n\ntype Item struct {\n" +
			"\tID int64 `gorm:\"primaryKey\" goninja:\"list,retrieve\"`\n" +
			field +
			"}\n"
		if err := os.WriteFile(modelFile, []byte(src), 0o644); err != nil {
			t.Fatalf("writing model file: %v", err)
		}
	}
	writeModel("")

	gen := generator{outDir: outDir, pkg: "api", modelsImport: "example.com/models", modelsPkg: "models"}
	if err := gen.run(modelsDir); err != nil {
		t.Fatalf("seeding initial generation: %v", err)
	}
	generated := filepath.Join(outDir, "item_generated.go")
	before, err := os.Stat(generated)
	if err != nil {
		t.Fatalf("expected initial generated file: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 1)
	go func() { errCh <- watchAndRegenerate(ctx, modelsDir, gen) }()

	// Give the watcher time to register before the write it needs to see.
	time.Sleep(100 * time.Millisecond)
	writeModel("\tName string `goninja:\"list,retrieve,create,update\"`\n")

	deadline := time.Now().Add(3 * time.Second)
	var after os.FileInfo
	for time.Now().Before(deadline) {
		after, err = os.Stat(generated)
		if err == nil && after.ModTime().After(before.ModTime()) {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if err != nil || !after.ModTime().After(before.ModTime()) {
		t.Fatal("watchAndRegenerate: generated file was not regenerated after the model file changed")
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("watchAndRegenerate: err = %v, want nil after shutdown", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("watchAndRegenerate: did not shut down after context cancellation")
	}
}
