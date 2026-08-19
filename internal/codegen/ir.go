// Package codegen implements the goninja Phase 0 prototype: parse Go
// struct declarations annotated with a `goninja` tag and render generated
// Go source from them.
//
// This is deliberately minimal (single tag vocabulary: list/retrieve/create,
// no ORM, in-memory store) — its only purpose is to answer the Phase 0
// decision-gate questions in goninja-implementation-plan.md before any
// investment in the real engine.
package codegen

import "strings"

// scalarGoTypes are the field types treated as plain columns. Anything
// else (a named struct type, or a slice of one) is treated as a relation
// — see Field.IsRelation.
var scalarGoTypes = map[string]bool{
	"string": true, "bool": true,
	"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
	"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
	"float32": true, "float64": true,
	"byte": true, "rune": true,
	"time.Time": true,
}

// Field is one struct field annotated with a `goninja` tag.
type Field struct {
	// Name is the Go field name, e.g. "Title".
	Name string
	// GoType is the field's type as written in source, e.g. "string", "int64".
	GoType string
	// JSONName is the field's json tag name, defaulting to the lowercased
	// field name when no `json` tag is present.
	JSONName string
	// Tags are the comma-separated values from the `goninja` struct tag,
	// e.g. []string{"list", "retrieve"}.
	Tags []string
	// ValidateTag is the raw content of the field's `validate` struct tag
	// (go-playground/validator syntax, e.g. "required,max=120"), empty if
	// none was present. Only meaningful on Create/Update schema fields.
	ValidateTag string
}

// HasTag reports whether the field is annotated with the given goninja tag.
func (f Field) HasTag(tag string) bool {
	for _, t := range f.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// IsRelation reports whether the field's type is not a plain scalar column
// — i.e. a named struct type (or slice of one), such as a GORM
// belongs-to/has-many association. Retrieve uses this to decide which
// fields to Preload.
func (f Field) IsRelation() bool {
	t := strings.TrimPrefix(f.GoType, "*")
	t = strings.TrimPrefix(t, "[]")
	t = strings.TrimPrefix(t, "*")
	return !scalarGoTypes[t]
}

// Model is a struct type discovered in the models package.
type Model struct {
	// Name is the Go type name, e.g. "Task".
	Name string
	// Fields are the struct's goninja-annotated fields, in declaration order.
	Fields []Field
}

// ListFields returns the fields exposed on the list schema.
func (m Model) ListFields() []Field {
	return m.fieldsWithTag("list")
}

// RetrieveFields returns the fields exposed on the retrieve schema.
func (m Model) RetrieveFields() []Field {
	return m.fieldsWithTag("retrieve")
}

// CreateFields returns the fields accepted on create.
func (m Model) CreateFields() []Field {
	return m.fieldsWithTag("create")
}

// UpdateFields returns the fields accepted on update.
func (m Model) UpdateFields() []Field {
	return m.fieldsWithTag("update")
}

func (m Model) fieldsWithTag(tag string) []Field {
	var out []Field
	for _, f := range m.Fields {
		if f.HasTag(tag) {
			out = append(out, f)
		}
	}
	return out
}

// NameLower is the model name lowercased, used for route paths and
// receiver-free helper names, e.g. "Task" -> "task".
func (m Model) NameLower() string {
	return lower(m.Name)
}

func lower(s string) string {
	if s == "" {
		return s
	}
	b := []byte(s)
	if b[0] >= 'A' && b[0] <= 'Z' {
		b[0] += 'a' - 'A'
	}
	return string(b)
}
