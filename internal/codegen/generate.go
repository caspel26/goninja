package codegen

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

var (
	modelTmpl   = template.Must(template.ParseFS(templatesFS, "templates/model.go.tmpl"))
	runtimeTmpl = template.Must(template.ParseFS(templatesFS, "templates/runtime.go.tmpl"))
)

// Generate renders one <model>_generated.go file per model plus a shared
// runtime_generated.go into outDir, all under the given package name.
// modelsImportPath/modelsPkg identify the package the models were parsed
// from, so the generated code can import and reference them (e.g.
// "github.com/caspel26/goninja/examples/prototype/models" / "models").
func Generate(models []Model, outDir, packageName, modelsImportPath, modelsPkg string) error {
	if len(models) == 0 {
		return fmt.Errorf("codegen: no goninja-annotated models found")
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("codegen: creating %s: %w", outDir, err)
	}

	if err := renderFile(runtimeTmpl, struct{ Package string }{packageName},
		filepath.Join(outDir, "runtime_generated.go")); err != nil {
		return err
	}

	for _, m := range models {
		data := struct {
			Package      string
			Model        Model
			ModelsImport string
			ModelsPkg    string
		}{
			Package:      packageName,
			Model:        m,
			ModelsImport: modelsImportPath,
			ModelsPkg:    modelsPkg,
		}

		path := filepath.Join(outDir, m.NameLower()+"_generated.go")
		if err := renderFile(modelTmpl, data, path); err != nil {
			return err
		}
	}

	return nil
}

func renderFile(tmpl *template.Template, data any, path string) error {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("codegen: rendering %s: %w", path, err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("codegen: %s produced invalid Go source: %w\n---\n%s", path, err, buf.String())
	}

	if err := os.WriteFile(path, formatted, 0o644); err != nil {
		return fmt.Errorf("codegen: writing %s: %w", path, err)
	}
	return nil
}
