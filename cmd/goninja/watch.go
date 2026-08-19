package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// debounce coalesces the burst of events a single save can produce (many
// editors write a temp file then rename it over the target) into one
// regeneration.
const debounce = 300 * time.Millisecond

// watchAndRegenerate watches modelsDir for .go file changes and re-runs
// gen.run on each one, debounced, until ctx is done. It returns nil on a
// clean shutdown (ctx canceled) or the first unrecoverable watcher setup
// error; a failed regeneration is reported to stderr and the watch loop
// continues, since a transient typo shouldn't kill the whole watch session.
func watchAndRegenerate(ctx context.Context, modelsDir string, gen generator) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("watch: %w", err)
	}
	defer watcher.Close()

	if err := watcher.Add(modelsDir); err != nil {
		return fmt.Errorf("watch: %w", err)
	}

	fmt.Printf("goninja: watching %s for changes (Ctrl+C to stop)\n", modelsDir)

	var timer *time.Timer
	defer func() {
		if timer != nil {
			timer.Stop()
		}
	}()

	regenerate := func() {
		if err := gen.run(modelsDir); err != nil {
			fmt.Fprintln(os.Stderr, "goninja:", err)
		}
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if !shouldRegenerate(event) {
				continue
			}
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(debounce, regenerate)
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintln(os.Stderr, "goninja: watch error:", err)
		}
	}
}

// shouldRegenerate reports whether a filesystem event is a model file
// change worth regenerating for: a .go file being written, created, or
// renamed into place (the common editor save-via-temp-file pattern).
func shouldRegenerate(event fsnotify.Event) bool {
	if filepath.Ext(event.Name) != ".go" {
		return false
	}
	return event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Rename)
}
