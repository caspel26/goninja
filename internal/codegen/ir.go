// Package codegen implements the goninja Phase 0 prototype: parse Go
// struct declarations annotated with a `goninja` tag and render generated
// Go source from them.
//
// This is deliberately minimal (single tag vocabulary: list/retrieve/create,
// no ORM, in-memory store) — its only purpose is to answer the Phase 0
// decision-gate questions in goninja-implementation-plan.md before any
// investment in the real engine.
package codegen

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
