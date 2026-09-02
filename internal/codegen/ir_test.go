package codegen

import "testing"

func TestField_HasTag(t *testing.T) {
	f := Field{Tags: []string{"list", "retrieve"}}
	if !f.HasTag("list") {
		t.Error("HasTag(list) = false, want true")
	}
	if f.HasTag("create") {
		t.Error("HasTag(create) = true, want false")
	}
}

func TestField_IsRelation(t *testing.T) {
	cases := []struct {
		goType string
		want   bool
	}{
		{"string", false},
		{"int64", false},
		{"time.Time", false},
		{"Author", true},
		{"*Author", true},
		{"[]Author", true},
	}
	for _, c := range cases {
		f := Field{GoType: c.goType}
		if got := f.IsRelation(); got != c.want {
			t.Errorf("IsRelation(%q) = %v, want %v", c.goType, got, c.want)
		}
	}
}

func TestField_IsByID(t *testing.T) {
	f := Field{Tags: []string{"retrieve", "byid"}}
	if !f.IsByID() {
		t.Error("IsByID() = false, want true")
	}
	f2 := Field{Tags: []string{"retrieve"}}
	if f2.IsByID() {
		t.Error("IsByID() = true, want false")
	}
}

func TestField_IsSlice(t *testing.T) {
	cases := []struct {
		goType string
		want   bool
	}{
		{"string", false},
		{"Author", false},
		{"[]Author", true},
	}
	for _, c := range cases {
		f := Field{GoType: c.goType}
		if got := f.IsSlice(); got != c.want {
			t.Errorf("IsSlice(%q) = %v, want %v", c.goType, got, c.want)
		}
	}
}

func TestField_RelatedGoType(t *testing.T) {
	cases := []struct {
		goType string
		want   string
	}{
		{"Author", "Author"},
		{"[]Author", "Author"},
	}
	for _, c := range cases {
		f := Field{GoType: c.goType}
		if got := f.RelatedGoType(); got != c.want {
			t.Errorf("RelatedGoType(%q) = %q, want %q", c.goType, got, c.want)
		}
	}
}

func TestField_RelatedIDOpenAPITypeAndFormat(t *testing.T) {
	f := Field{RelatedIDGoType: "int64"}
	if got := f.RelatedIDOpenAPIType(); got != "integer" {
		t.Errorf("RelatedIDOpenAPIType() = %q, want integer", got)
	}
	if got := f.RelatedIDOpenAPIFormat(); got != "int64" {
		t.Errorf("RelatedIDOpenAPIFormat() = %q, want int64", got)
	}

	f2 := Field{RelatedIDGoType: "string"}
	if got := f2.RelatedIDOpenAPIType(); got != "string" {
		t.Errorf("RelatedIDOpenAPIType() = %q, want string", got)
	}
	if got := f2.RelatedIDOpenAPIFormat(); got != "" {
		t.Errorf("RelatedIDOpenAPIFormat() = %q, want empty", got)
	}
}

func TestField_IsBoolIsStringIsFloatIsNumeric(t *testing.T) {
	if !(Field{GoType: "bool"}).IsBool() {
		t.Error("IsBool() = false, want true")
	}
	if !(Field{GoType: "string"}).IsString() {
		t.Error("IsString() = false, want true")
	}
	if !(Field{GoType: "float64"}).IsFloat() {
		t.Error("IsFloat() = false, want true")
	}
	if (Field{GoType: "string"}).IsFloat() {
		t.Error("IsFloat() on string = true, want false")
	}
	if !(Field{GoType: "int64"}).IsNumeric() {
		t.Error("IsNumeric() on int64 = false, want true")
	}
	if (Field{GoType: "bool"}).IsNumeric() {
		t.Error("IsNumeric() on bool = true, want false")
	}
	if (Field{GoType: "string"}).IsNumeric() {
		t.Error("IsNumeric() on string = true, want false")
	}
	if (Field{GoType: "time.Time"}).IsNumeric() {
		t.Error("IsNumeric() on time.Time = true, want false")
	}
}

func TestField_OpenAPIType(t *testing.T) {
	cases := []struct {
		goType string
		want   string
	}{
		{"bool", "boolean"},
		{"string", "string"},
		{"float64", "number"},
		{"time.Time", "string"},
		{"int64", "integer"},
		{"Author", "object"},
	}
	for _, c := range cases {
		f := Field{GoType: c.goType}
		if got := f.OpenAPIType(); got != c.want {
			t.Errorf("OpenAPIType(%q) = %q, want %q", c.goType, got, c.want)
		}
	}
}

func TestField_OpenAPIFormat(t *testing.T) {
	cases := []struct {
		goType string
		want   string
	}{
		{"time.Time", "date-time"},
		{"float64", "double"},
		{"int64", "int64"},
		{"uint32", "int64"},
		{"string", ""},
		{"bool", ""},
	}
	for _, c := range cases {
		f := Field{GoType: c.goType}
		if got := f.OpenAPIFormat(); got != c.want {
			t.Errorf("OpenAPIFormat(%q) = %q, want %q", c.goType, got, c.want)
		}
	}
}

func TestField_IsRequired(t *testing.T) {
	if !(Field{ValidateTag: "required,max=5"}).IsRequired() {
		t.Error("IsRequired() = false, want true")
	}
	if (Field{ValidateTag: "max=5"}).IsRequired() {
		t.Error("IsRequired() = true, want false")
	}
	if (Field{ValidateTag: ""}).IsRequired() {
		t.Error("IsRequired() on empty tag = true, want false")
	}
}

func TestModel_FieldAccessors(t *testing.T) {
	m := Model{
		Name: "Task",
		Fields: []Field{
			{Name: "ID", GoType: "int64", Tags: []string{"list", "retrieve"}},
			{Name: "Title", GoType: "string", Tags: []string{"list", "retrieve", "create", "update", "filter"}},
		},
	}
	if len(m.ListFields()) != 2 {
		t.Errorf("ListFields() = %d, want 2", len(m.ListFields()))
	}
	if len(m.RetrieveFields()) != 2 {
		t.Errorf("RetrieveFields() = %d, want 2", len(m.RetrieveFields()))
	}
	if len(m.CreateFields()) != 1 {
		t.Errorf("CreateFields() = %d, want 1", len(m.CreateFields()))
	}
	if len(m.UpdateFields()) != 1 {
		t.Errorf("UpdateFields() = %d, want 1", len(m.UpdateFields()))
	}
	if len(m.FilterFields()) != 1 {
		t.Errorf("FilterFields() = %d, want 1", len(m.FilterFields()))
	}
	if got := m.IDGoType(); got != "int64" {
		t.Errorf("IDGoType() = %q, want int64", got)
	}
	if got := m.NameLower(); got != "task" {
		t.Errorf("NameLower() = %q, want task", got)
	}
}

func TestModel_ListSelectColumns(t *testing.T) {
	m := Model{
		Name: "Author",
		Fields: []Field{
			{Name: "ID", GoType: "string", JSONName: "id", Tags: []string{"list", "retrieve"}},
			{Name: "Name", GoType: "string", JSONName: "name", Tags: []string{"list", "retrieve"}},
			{Name: "Bio", GoType: "string", JSONName: "bio", Tags: []string{"retrieve"}}, // list-excluded
		},
	}
	got := m.ListSelectColumns()
	want := []string{"id", "name"}
	if len(got) != len(want) {
		t.Fatalf("ListSelectColumns() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ListSelectColumns()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestModel_ListSelectColumns_UsesDBColumns(t *testing.T) {
	m := Model{Fields: []Field{
		{Name: "ID", GoType: "string", JSONName: "id", DBColumn: "id", Tags: []string{"list"}},
		{Name: "CreatedAt", GoType: "time.Time", JSONName: "createdAt", DBColumn: "created_at", Tags: []string{"list"}},
		{Name: "Label", GoType: "string", JSONName: "displayName", DBColumn: "display_name", Tags: []string{"list"}},
	}}
	got := m.ListSelectColumns()
	want := []string{"id", "created_at", "display_name"}
	if len(got) != len(want) {
		t.Fatalf("ListSelectColumns() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ListSelectColumns()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestModel_ListSelectColumns_NilOnRelationField(t *testing.T) {
	// A `list`-tagged relation field isn't a column — a Select naming it
	// would be a SQL error, so ListSelectColumns must bail out to nil
	// (disabling the optimization) rather than emit a broken query.
	m := Model{
		Name: "Book",
		Fields: []Field{
			{Name: "ID", GoType: "string", JSONName: "id", Tags: []string{"list"}},
			{Name: "Author", GoType: "Author", JSONName: "author", Tags: []string{"list"}},
		},
	}
	if got := m.ListSelectColumns(); got != nil {
		t.Errorf("ListSelectColumns() with a relation field = %v, want nil", got)
	}
}

func TestModel_IDGoType_DefaultsWithoutIDField(t *testing.T) {
	m := Model{Name: "Weird", Fields: []Field{{Name: "Name", GoType: "string"}}}
	if got := m.IDGoType(); got != "int64" {
		t.Errorf("IDGoType() without an ID field = %q, want int64 (default)", got)
	}
}

func TestModel_UsesTime(t *testing.T) {
	m := Model{Fields: []Field{
		{Name: "CreatedAt", GoType: "time.Time", Tags: []string{"list"}},
	}}
	if !m.UsesTime() {
		t.Error("UsesTime() = false, want true")
	}

	m2 := Model{Fields: []Field{
		{Name: "Internal", GoType: "time.Time"}, // no list/retrieve/create/update tag
	}}
	if m2.UsesTime() {
		t.Error("UsesTime() with an untagged time.Time field = true, want false")
	}

	m3 := Model{Fields: []Field{{Name: "Title", GoType: "string", Tags: []string{"list"}}}}
	if m3.UsesTime() {
		t.Error("UsesTime() without any time.Time field = true, want false")
	}
}

func TestModel_NeedsStrconv(t *testing.T) {
	intID := Model{Fields: []Field{{Name: "ID", GoType: "int64"}}}
	if !intID.NeedsStrconv() {
		t.Error("NeedsStrconv() with an int64 ID = false, want true")
	}

	stringIDNoFilters := Model{Fields: []Field{{Name: "ID", GoType: "string"}}}
	if stringIDNoFilters.NeedsStrconv() {
		t.Error("NeedsStrconv() with a string ID and no non-string filters = true, want false")
	}

	stringIDNumericFilter := Model{Fields: []Field{
		{Name: "ID", GoType: "string"},
		{Name: "Price", GoType: "float64", Tags: []string{"filter"}},
	}}
	if !stringIDNumericFilter.NeedsStrconv() {
		t.Error("NeedsStrconv() with a string ID and a numeric filter field = false, want true")
	}

	stringIDStringFilter := Model{Fields: []Field{
		{Name: "ID", GoType: "string"},
		{Name: "Name", GoType: "string", Tags: []string{"filter"}},
	}}
	if stringIDStringFilter.NeedsStrconv() {
		t.Error("NeedsStrconv() with a string ID and only a string filter field = true, want false")
	}
}

func TestLower(t *testing.T) {
	cases := map[string]string{
		"Task":  "task",
		"task":  "task",
		"":      "",
		"Book1": "book1",
	}
	for in, want := range cases {
		if got := lower(in); got != want {
			t.Errorf("lower(%q) = %q, want %q", in, got, want)
		}
	}
}
