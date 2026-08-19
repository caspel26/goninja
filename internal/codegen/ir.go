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
	// RelatedIDGoType is the Go type of the related model's ID field,
	// resolved by Generate (generate.go) across the full model list — only
	// meaningful when IsRelation() && IsByID() (plan section 5.12). Left
	// empty by ParseModels itself; Generate fills it in before rendering,
	// falling back to "string" if the related model can't be found in the
	// same generation run.
	RelatedIDGoType string
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

// IsByID reports whether a relation field carries the "byid" modifier on
// its goninja tag (e.g. `goninja:"retrieve,byid"`, plan section 5.12): its
// Retrieve schema exposes only the related model's ID, as "<field>_id",
// skipping that field's Preload — instead of nesting the related model's
// full Retrieve type, today's only behavior for a relation field with no
// modifier. Meaningless on a non-relation field.
func (f Field) IsByID() bool {
	return f.HasTag("byid")
}

// RelatedIDOpenAPIType/RelatedIDOpenAPIFormat mirror OpenAPIType/
// OpenAPIFormat, but for a byid relation field's synthesized "<field>_id"
// property — typed after the related model's own ID (RelatedIDGoType,
// resolved by Generate) rather than the relation field's own unused struct
// type.
func (f Field) RelatedIDOpenAPIType() string {
	if f.RelatedIDGoType == "int64" {
		return "integer"
	}
	return "string"
}

// RelatedIDOpenAPIFormat is RelatedIDGoType's OpenAPI "format", mirroring
// OpenAPIFormat.
func (f Field) RelatedIDOpenAPIFormat() string {
	if f.RelatedIDGoType == "int64" {
		return "int64"
	}
	return ""
}

// floatGoTypes are the Go types treated as floating point for filter
// generation (min/max range filters, strconv.ParseFloat parsing).
var floatGoTypes = map[string]bool{"float32": true, "float64": true}

// IsBool reports whether the field's Go type is bool.
func (f Field) IsBool() bool { return f.GoType == "bool" }

// IsString reports whether the field's Go type is string.
func (f Field) IsString() bool { return f.GoType == "string" }

// IsFloat reports whether the field's Go type is float32/float64.
func (f Field) IsFloat() bool { return floatGoTypes[f.GoType] }

// IsNumeric reports whether the field's Go type is an int/uint/float
// scalar (excluding bool and string) — used to decide whether a `filter`
// field also gets Min/Max range filters, in addition to exact match.
func (f Field) IsNumeric() bool {
	return scalarGoTypes[f.GoType] && !f.IsBool() && !f.IsString() && f.GoType != "time.Time"
}

// OpenAPIType returns the JSON Schema "type" for the field's Go type, used
// when generating a schema property for it (plan section 5.10/Fase 5). A
// relation field's own type is never used here — the template always emits
// a $ref to the related model's Retrieve schema instead (see IsRelation).
func (f Field) OpenAPIType() string {
	switch {
	case f.IsBool():
		return "boolean"
	case f.IsString():
		return "string"
	case f.IsFloat():
		return "number"
	case f.GoType == "time.Time":
		return "string"
	case f.IsNumeric():
		return "integer"
	default:
		return "object"
	}
}

// OpenAPIFormat returns the JSON Schema "format" for the field's Go type,
// or "" when the type has none (e.g. plain string, boolean).
func (f Field) OpenAPIFormat() string {
	switch {
	case f.GoType == "time.Time":
		return "date-time"
	case f.IsFloat():
		return "double"
	case strings.HasPrefix(f.GoType, "int") || strings.HasPrefix(f.GoType, "uint"):
		return "int64"
	default:
		return ""
	}
}

// IsRequired reports whether the field's validate tag includes "required",
// used to populate a Create/Update schema's "required" list.
func (f Field) IsRequired() bool {
	for _, part := range strings.Split(f.ValidateTag, ",") {
		if part == "required" {
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

// UpdateFields returns the fields accepted on update.
func (m Model) UpdateFields() []Field {
	return m.fieldsWithTag("update")
}

// FilterFields returns the fields exposed on the generated <Model>Filters
// struct and honored by List's query building (plan section 5.5/Fase 4).
func (m Model) FilterFields() []Field {
	return m.fieldsWithTag("filter")
}

// IDGoType is the Go type of the model's ID field, e.g. "int64" or
// "string". Every model is still assumed to carry a field literally named
// "ID" (see CLAUDE.md's generator conventions) — this only generalizes its
// *type*, so generated Retrieve/Update/Delete signatures and path-value
// parsing adapt: "int64" parses the path value with strconv.ParseInt,
// anything else (in practice "string", for a UUID primary key generated by
// goninja.NewUUID in Create) is taken as-is. Defaults to "int64" if no ID
// field is found, matching the historical hardcoded assumption.
func (m Model) IDGoType() string {
	for _, f := range m.Fields {
		if f.Name == "ID" {
			return f.GoType
		}
	}
	return "int64"
}

// UsesTime reports whether any generated schema field is time.Time,
// meaning the generated file needs to import "time".
func (m Model) UsesTime() bool {
	for _, f := range m.Fields {
		if f.GoType == "time.Time" && (f.HasTag("list") || f.HasTag("retrieve") || f.HasTag("create") || f.HasTag("update")) {
			return true
		}
	}
	return false
}

// NeedsStrconv reports whether the generated file has any use for the
// "strconv" import: an int64 ID (path-value parsing) or a `filter` field
// whose query-parameter parsing isn't a plain string passthrough (bool,
// int, or float).
func (m Model) NeedsStrconv() bool {
	if m.IDGoType() == "int64" {
		return true
	}
	for _, f := range m.FilterFields() {
		if !f.IsString() {
			return true
		}
	}
	return false
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
