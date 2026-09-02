package codegen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"

	"gorm.io/gorm/schema"
)

// ParseModels scans every .go file directly inside dir and extracts every
// struct type that has at least one field carrying a `goninja` tag.
func ParseModels(dir string) ([]Model, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("codegen: parsing %s: %w", dir, err)
	}

	var models []Model
	for _, pkg := range pkgs {
		namedScalars := collectNamedScalars(pkg.Files)
		for path, file := range pkg.Files {
			fileModels, err := parseFileModelsWithScalars(path, file, namedScalars)
			if err != nil {
				return nil, err
			}
			models = append(models, fileModels...)
		}
	}
	return models, nil
}

// parseFileModels extracts every tagged struct type declared directly in
// file (split out of ParseModels to keep each loop level's nesting, and
// therefore cognitive complexity, in check).
func parseFileModels(path string, file *ast.File) ([]Model, error) {
	return parseFileModelsWithScalars(path, file, nil)
}

func parseFileModelsWithScalars(path string, file *ast.File, namedScalars map[string]string) ([]Model, error) {
	var models []Model
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		declModels, err := parseGenDeclModelsWithScalars(path, gen, namedScalars)
		if err != nil {
			return nil, err
		}
		models = append(models, declModels...)
	}
	return models, nil
}

// parseGenDeclModels extracts every tagged struct type among gen's specs
// (a `type (...)` block can declare more than one).
func parseGenDeclModels(path string, gen *ast.GenDecl) ([]Model, error) {
	return parseGenDeclModelsWithScalars(path, gen, nil)
}

func parseGenDeclModelsWithScalars(path string, gen *ast.GenDecl, namedScalars map[string]string) ([]Model, error) {
	var models []Model
	for _, spec := range gen.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		st, ok := ts.Type.(*ast.StructType)
		if !ok {
			continue
		}
		model, hasTaggedField, err := parseStructWithScalars(ts.Name.Name, st, namedScalars)
		if err != nil {
			return nil, fmt.Errorf("codegen: %s: %w", ts.Name.Name, err)
		}
		if hasTaggedField {
			model.SourceFile = path
			models = append(models, model)
		}
	}
	return models, nil
}

func parseStruct(name string, st *ast.StructType) (Model, bool, error) {
	return parseStructWithScalars(name, st, nil)
}

func parseStructWithScalars(name string, st *ast.StructType, namedScalars map[string]string) (Model, bool, error) {
	model := Model{Name: name}
	hasTaggedField := false

	for _, f := range st.Fields.List {
		if len(f.Names) == 0 || f.Tag == nil {
			continue
		}
		goType, err := exprToString(f.Type)
		if err != nil {
			return model, false, err
		}

		rawTag := strings.Trim(f.Tag.Value, "`")
		tag := reflect.StructTag(rawTag)

		goninjaTag := tag.Get("goninja")
		if goninjaTag == "" {
			continue
		}
		hasTaggedField = true

		validateTag := tag.Get("validate")
		gormTag := tag.Get("gorm")

		for _, fname := range f.Names {
			underlying := namedScalars[strings.TrimPrefix(strings.TrimPrefix(goType, "*"), "[]")]
			model.Fields = append(model.Fields, Field{
				Name:             fname.Name,
				GoType:           goType,
				UnderlyingGoType: underlying,
				JSONName:         jsonNameFor(fname.Name, tag.Get("json")),
				DBColumn:         dbColumnFor(fname.Name, gormTag),
				Tags:             splitTag(goninjaTag),
				ValidateTag:      validateTag,
			})
		}
	}

	return model, hasTaggedField, nil
}

// collectNamedScalars finds aliases and defined types whose underlying type
// is one of goninja's supported scalar columns. The map is deliberately
// package-local: a named type from another package has different import and
// database semantics, so validation rejects it with an actionable error.
func collectNamedScalars(files map[string]*ast.File) map[string]string {
	definitions := make(map[string]string)
	for _, file := range files {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				name, err := exprToString(ts.Type)
				if err == nil {
					definitions[ts.Name.Name] = name
				}
			}
		}
	}

	resolved := make(map[string]string)
	for name := range definitions {
		if scalar := resolveNamedScalar(name, definitions); scalar != "" {
			resolved[name] = scalar
		}
	}
	return resolved
}

func resolveNamedScalar(name string, definitions map[string]string) string {
	seen := make(map[string]bool)
	for {
		if seen[name] {
			return ""
		}
		seen[name] = true
		underlying, ok := definitions[name]
		if !ok {
			// time.Time needs special parsing and assignment code. Supporting a
			// named wrapper correctly requires a separate contract, so keep it
			// explicit rather than accepting a type the template cannot safely
			// construct yet.
			if name != goTypeTime && scalarGoTypes[name] {
				return name
			}
			return ""
		}
		name = underlying
	}
}

func jsonNameFor(fieldName, tagValue string) string {
	if tagValue == "" || tagValue == "-" {
		return lower(fieldName)
	}
	if idx := strings.Index(tagValue, ","); idx >= 0 {
		return tagValue[:idx]
	}
	return tagValue
}

func dbColumnFor(fieldName, gormTag string) string {
	if column := schema.ParseTagSetting(gormTag, ";")["COLUMN"]; column != "" {
		return column
	}
	return defaultDBColumn(fieldName)
}

func splitTag(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// exprToString renders a type expression back to source form for the small
// subset of types the prototype needs to support (identifiers, selectors,
// pointers, slices).
func exprToString(expr ast.Expr) (string, error) {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name, nil
	case *ast.SelectorExpr:
		pkg, err := exprToString(t.X)
		if err != nil {
			return "", err
		}
		return pkg + "." + t.Sel.Name, nil
	case *ast.StarExpr:
		inner, err := exprToString(t.X)
		if err != nil {
			return "", err
		}
		return "*" + inner, nil
	case *ast.ArrayType:
		inner, err := exprToString(t.Elt)
		if err != nil {
			return "", err
		}
		return "[]" + inner, nil
	default:
		return "", fmt.Errorf("unsupported field type %T", expr)
	}
}
