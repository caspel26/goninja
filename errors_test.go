package goninja

import "testing"

func TestNotFound_Error(t *testing.T) {
	err := NotFound{Resource: "Book", ID: "abc"}
	want := "Book abc not found"
	if got := err.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestValidationError_Error(t *testing.T) {
	err := ValidationError{Fields: map[string]string{"name": "required"}}
	if got := err.Error(); got == "" {
		t.Error("Error() = empty string, want a message containing the failed fields")
	}
}

func TestBadRequest_Error(t *testing.T) {
	err := BadRequest{Detail: "invalid limit"}
	if got := err.Error(); got != "invalid limit" {
		t.Errorf("Error() = %q, want %q", got, "invalid limit")
	}
}
