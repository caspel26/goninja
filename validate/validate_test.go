package validate

import (
	"testing"

	"github.com/caspel26/goninja"
)

type validateFixture struct {
	Name  string `json:"name" validate:"required,max=5"`
	Email string `json:"email" validate:"required,email"`
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

	ve, ok := err.(goninja.ValidationError)
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
