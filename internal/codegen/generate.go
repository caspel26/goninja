package codegen

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

var modelTmpl = template.Must(template.ParseFS(templatesFS, "templates/model.go.tmpl"))

// Generate renders one <model>_generated.go file per model into outDir,
// under the given package name. There's no shared runtime file: every
// helper a generated file needs (JSON responses, error mapping,
// validation) lives in the goninja package and is called directly, so
// nothing needs deduplicating across models.
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

	resolveByIDFields(models)

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

// resolveByIDFields fills in RelatedIDGoType on every byid relation field
// (goninja:"...,byid" — plan section 5.12) by looking up the related
// model's own ID type across the full model list, so the generated
// Retrieve schema and OpenAPI fragment can type "<field>_id" correctly
// instead of assuming string for every related model.
func resolveByIDFields(models []Model) {
	for i := range models {
		for j := range models[i].Fields {
			f := &models[i].Fields[j]
			if f.IsRelation() && f.IsByID() {
				f.RelatedIDGoType = relatedIDGoType(models, f.GoType)
			}
		}
	}
}

// relatedIDGoType finds the model matching typeName (a relation field's Go
// type, e.g. "Author" or "[]Review") among models and returns its
// IDGoType, defaulting to "string" if no match is found in this generation
// run.
func relatedIDGoType(models []Model, typeName string) string {
	t := strings.TrimPrefix(typeName, "[]")
	t = strings.TrimPrefix(t, "*")
	for _, m := range models {
		if m.Name == t {
			return m.IDGoType()
		}
	}
	return "string"
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
