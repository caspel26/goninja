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
	if err.Error() == "" {
		t.Error("Error() = empty string, want a message containing the failed fields")
	}
}

func TestBadRequest_Error(t *testing.T) {
	err := BadRequest{Detail: "invalid limit"}
	if got := err.Error(); got != "invalid limit" {
		t.Errorf("Error() = %q, want %q", got, "invalid limit")
	}
}

func TestUnauthorized_Error(t *testing.T) {
	if got := (Unauthorized{}).Error(); got != "unauthorized" {
		t.Errorf("Error() = %q, want %q", got, "unauthorized")
	}
	if got := (Unauthorized{Detail: "token expired"}).Error(); got != "token expired" {
		t.Errorf("Error() = %q, want %q", got, "token expired")
	}
}

func TestErrorCode_DefaultsAndOverrides(t *testing.T) {
	tests := []struct {
		name string
		err  CodedError
		want string
	}{
		{"not found default", NotFound{Resource: "Book", ID: 1}, "NOT_FOUND"},
		{"not found override", NotFound{Resource: "Book", ID: 1, Code: "BOOK_NOT_FOUND"}, "BOOK_NOT_FOUND"},
		{"validation default", ValidationError{}, "VALIDATION_FAILED"},
		{"validation override", ValidationError{Code: "CUSTOM"}, "CUSTOM"},
		{"bad request default", BadRequest{}, "BAD_REQUEST"},
		{"bad request override", BadRequest{Code: "INVALID_ORDER_FIELD"}, "INVALID_ORDER_FIELD"},
		{"unauthorized default", Unauthorized{}, "UNAUTHORIZED"},
		{"unauthorized override", Unauthorized{Code: "TOKEN_EXPIRED"}, "TOKEN_EXPIRED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.ErrorCode(); got != tt.want {
				t.Errorf("ErrorCode() = %q, want %q", got, tt.want)
			}
		})
	}
}
