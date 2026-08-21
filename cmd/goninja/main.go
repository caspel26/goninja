// Command goninja is the CLI entry point for the goninja code generator.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/caspel26/goninja/internal/codegen"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "goninja:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return fmt.Errorf("missing command")
	}

	switch args[0] {
	case "generate":
		return runGenerate(args[1:])
	case "-h", "--help", "help":
		printUsage()
		return nil
	default:
		printUsage()
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runGenerate(args []string) error {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	modelsDir := fs.String("models", "./models", "directory containing goninja-annotated model structs")
	outDir := fs.String("out", "./internal/api", "output directory for generated code")
	pkg := fs.String("package", "api", "package name for generated code")
	modelsImport := fs.String("models-import", "", "import path of the models package (required)")
	modelsPkg := fs.String("models-pkg", "models", "package name of the models package, as used in Go source")
	watch := fs.Bool("watch", false, "watch -models for changes and regenerate automatically")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *modelsImport == "" {
		return fmt.Errorf("-models-import is required (e.g. github.com/you/yourapp/models)")
	}

	gen := generator{outDir: *outDir, pkg: *pkg, modelsImport: *modelsImport, modelsPkg: *modelsPkg}
	if err := gen.run(*modelsDir); err != nil {
		return err
	}

	if !*watch {
		return nil
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return watchAndRegenerate(ctx, *modelsDir, gen)
}

// generator bundles the fixed codegen.Generate arguments so a regeneration
// (initial or watch-triggered) is just gen.run(modelsDir).
type generator struct {
	outDir       string
	pkg          string
	modelsImport string
	modelsPkg    string
}

func (g generator) run(modelsDir string) error {
	models, err := codegen.ParseModels(modelsDir)
	if err != nil {
		return err
	}
	if err := codegen.Generate(models, g.outDir, g.pkg, g.modelsImport, g.modelsPkg); err != nil {
		return err
	}
	fmt.Printf("goninja: generated %d model(s) into %s\n", len(models), g.outDir)
	return nil
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `usage: goninja <command> [flags]

commands:
  generate   generate code from goninja-annotated model structs

flags for generate:
  -models string         directory containing model structs (default "./models")
  -out string            output directory for generated code (default "./internal/api")
  -package string        package name for generated code (default "api")
  -models-import string  import path of the models package (required)
  -models-pkg string     package name of the models package (default "models")
  -watch                 watch -models for changes and regenerate automatically`)
}
