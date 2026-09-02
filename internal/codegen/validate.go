package codegen

import (
	"errors"
	"fmt"
	"strings"
)

// idGoTypes are the ID field types the generator knows how to handle:
// "int64" is parsed out of the path with strconv.ParseInt, "string" is taken
// as-is and treated as a UUID primary key. Anything else would be silently
// handled as the latter and produce a resource whose Retrieve could never
// match a row.
var idGoTypes = map[string]bool{"int64": true, "string": true}

// Validate reports everything about models that the generator cannot turn
// into working code.
//
// Without it these mistakes surface far from their cause — as a compile error
// inside a generated `DO NOT EDIT` file, or as a route that quietly answers
// the wrong thing — which is the opposite of what goninja promises. Every
// problem found is reported, not just the first, so one run tells the whole
// story.
func Validate(models []Model) error {
	modelNames := make(map[string]struct{}, len(models))
	for _, m := range models {
		modelNames[m.Name] = struct{}{}
	}

	var problems []error
	for _, m := range models {
		problems = append(problems, m.validate(modelNames)...)
	}
	return errors.Join(problems...)
}

// validate returns every problem with a single model.
func (m Model) validate(modelNames map[string]struct{}) []error {
	var errs []error
	where := m.Name
	if m.SourceFile != "" {
		where = m.SourceFile + ": " + m.Name
	}

	errs = append(errs, m.validateID(where)...)
	for _, f := range m.Fields {
		errs = append(errs, f.validate(where, modelNames)...)
	}
	return errs
}

// validateID checks the primary key, which every generated Retrieve, Update
// and Delete is typed on.
func (m Model) validateID(where string) []error {
	for _, f := range m.Fields {
		if f.Name != "ID" {
			continue
		}
		if !idGoTypes[f.GoType] {
			return []error{fmt.Errorf(
				"%s: ID is %s, which the generator cannot use as a primary key; "+
					"use int64 for a serial key or string for a UUID",
				where, f.GoType)}
		}
		return nil
	}

	return []error{fmt.Errorf(
		"%s: no goninja-tagged field named ID; every model needs one, typed "+
			"int64 or string, and it must carry a goninja tag to be exposed "+
			"(e.g. `goninja:\"list,retrieve\"`)",
		where)}
}

// validate returns every problem with a single field.
func (f Field) validate(where string, modelNames map[string]struct{}) []error {
	var errs []error
	at := where + "." + f.Name
	relation := f.IsRelation()
	knownRelation := false
	if relation {
		_, knownRelation = modelNames[f.RelatedModelName()]
		if !knownRelation {
			errs = append(errs, fmt.Errorf(
				"%s: %s is neither a supported scalar nor an annotated goninja model relation; "+
					"named scalar and external types are not supported yet",
				at, f.GoType))
		}
	}

	if f.IsByID() && !relation {
		errs = append(errs, fmt.Errorf(
			`%s: the "byid" modifier only applies to a relation field, and %s is not one`,
			at, f.GoType))
	}

	// RelatedGoType strips only the "[]" wrapper, so a pointer would be
	// rendered into type names like "*AuthorRetrieve" and fail to compile.
	if relation && strings.Contains(f.GoType, "*") {
		errs = append(errs, fmt.Errorf(
			"%s: relation field is %s; pointers are not supported, use a struct "+
				"value for a belongs-to or a plain slice for a has-many",
			at, f.GoType))
	}

	if knownRelation && f.HasTag("filter") {
		errs = append(errs, fmt.Errorf(
			"%s: a relation field cannot be filtered, since it is not a column; "+
				"tag its foreign key field instead",
			at))
	}

	return errs
}
