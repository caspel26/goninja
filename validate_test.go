package goninja

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

type validateFixture struct {
	Name  string `json:"name" validate:"required,max=5"`
	Email string `json:"email" validate:"required,email"`
}

type validateFixtureNoJSONTag struct {
	Age int `validate:"required"`
}

func TestValidate_OK(t *testing.T) {
	v := validateFixture{Name: "Ann", Email: "ann@example.com"}
	if err := Validate(v); err != nil {
		t.Fatalf("Validate: unexpected error: %v", err)
	}
}

func TestValidate_ReturnsValidationErrorKeyedByJSONName(t *testing.T) {
	v := validateFixture{Name: "way too long", Email: "not-an-email"}

	err := Validate(v)
	if err == nil {
		t.Fatal("Validate: err = nil, want a ValidationError")
	}

	ve, ok := err.(ValidationError)
	if !ok {
		t.Fatalf("Validate: err type = %T, want goninja.ValidationError", err)
	}

	if tag, ok := ve.Fields["name"]; !ok || tag != "max" {
		t.Errorf(`Fields["name"] = %q, ok=%v, want "max", ok=true`, tag, ok)
	}
	if tag, ok := ve.Fields["email"]; !ok || tag != "email" {
		t.Errorf(`Fields["email"] = %q, ok=%v, want "email", ok=true`, tag, ok)
	}
}

func TestValidate_FallsBackToGoFieldNameWithoutJSONTag(t *testing.T) {
	err := Validate(validateFixtureNoJSONTag{})
	if err == nil {
		t.Fatal("Validate: err = nil, want a ValidationError")
	}

	ve, ok := err.(ValidationError)
	if !ok {
		t.Fatalf("Validate: err type = %T, want goninja.ValidationError", err)
	}

	if tag, ok := ve.Fields["Age"]; !ok || tag != "required" {
		t.Errorf(`Fields["Age"] = %q, ok=%v, want "required", ok=true`, tag, ok)
	}
}

type validateFixtureCustomTag struct {
	Code string `json:"code" validate:"oddlen"`
}

func TestRegisterValidation_CustomTagIsUsedByValidate(t *testing.T) {
	if err := RegisterValidation("oddlen", func(fl validator.FieldLevel) bool {
		return len(fl.Field().String())%2 == 1
	}); err != nil {
		t.Fatalf("RegisterValidation: unexpected error: %v", err)
	}

	if err := Validate(validateFixtureCustomTag{Code: "abc"}); err != nil {
		t.Errorf("Validate: unexpected error for odd-length code: %v", err)
	}

	err := Validate(validateFixtureCustomTag{Code: "abcd"})
	if err == nil {
		t.Fatal("Validate: err = nil, want a ValidationError for even-length code")
	}
	ve, ok := err.(ValidationError)
	if !ok {
		t.Fatalf("Validate: err type = %T, want goninja.ValidationError", err)
	}
	if tag, ok := ve.Fields["code"]; !ok || tag != "oddlen" {
		t.Errorf(`Fields["code"] = %q, ok=%v, want "oddlen", ok=true`, tag, ok)
	}
}

func TestValidate_NonValidationError(t *testing.T) {
	// Passing a non-struct triggers validator.InvalidValidationError, which
	// is not a validator.ValidationErrors, so Validate returns it verbatim.
	err := Validate(nil)
	if err == nil {
		t.Fatal("Validate(nil): err = nil, want a non-nil error")
	}
	if _, ok := err.(ValidationError); ok {
		t.Error("Validate(nil) returned a ValidationError, want the raw validator error")
	}
}
