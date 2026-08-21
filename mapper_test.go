package goninja

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDefaultErrorMapper_MapError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"not found", NotFound{Resource: "Book", ID: 1}, http.StatusNotFound, "NOT_FOUND"},
		{"validation error", ValidationError{Fields: map[string]string{"name": "required"}}, http.StatusUnprocessableEntity, "VALIDATION_FAILED"},
		{"bad request", BadRequest{Detail: "bad"}, http.StatusBadRequest, "BAD_REQUEST"},
		{"unauthorized", Unauthorized{}, http.StatusUnauthorized, "UNAUTHORIZED"},
		{"unknown error", errors.New("boom"), http.StatusInternalServerError, "INTERNAL"},
		{"not found with custom code", NotFound{Resource: "Book", ID: 1, Code: "BOOK_NOT_FOUND"}, http.StatusNotFound, "BOOK_NOT_FOUND"},
		{"validation error with custom code", ValidationError{Fields: map[string]string{"name": "required"}, Code: "CUSTOM_VALIDATION"}, http.StatusUnprocessableEntity, "CUSTOM_VALIDATION"},
		{"bad request with custom code", BadRequest{Detail: "bad", Code: "INVALID_ORDER_FIELD"}, http.StatusBadRequest, "INVALID_ORDER_FIELD"},
		{"unauthorized with custom code", Unauthorized{Code: "TOKEN_EXPIRED"}, http.StatusUnauthorized, "TOKEN_EXPIRED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, body := DefaultErrorMapper{}.MapError(tt.err)
			if status != tt.wantStatus {
				t.Errorf("status = %d, want %d", status, tt.wantStatus)
			}

			b, err := json.Marshal(body)
			if err != nil {
				t.Fatalf("marshal body: %v", err)
			}
			var decoded map[string]any
			if err := json.Unmarshal(b, &decoded); err != nil {
				t.Fatalf("unmarshal body: %v", err)
			}
			if decoded["code"] != tt.wantCode {
				t.Errorf("code = %v, want %q", decoded["code"], tt.wantCode)
			}
		})
	}
}

func TestDefaultErrorMapper_DoesNotLeakUnknownErrorMessage(t *testing.T) {
	_, body := DefaultErrorMapper{}.MapError(errors.New("sensitive internal detail"))
	b, _ := json.Marshal(body)
	if got := string(b); strings.Contains(got, "sensitive internal detail") {
		t.Errorf("MapError body leaked the underlying error message: %s", got)
	}
}

type customError struct{ msg string }

func (e customError) Error() string { return e.msg }

func TestNewErrorMapper_FirstMatchWins(t *testing.T) {
	mapper := NewErrorMapper(
		NewErrorMapping(func(err customError) (int, any) {
			return http.StatusTeapot, map[string]string{"code": "CUSTOM", "error": err.msg}
		}),
		NewErrorMapping(func(err NotFound) (int, any) {
			return http.StatusNotFound, map[string]string{"code": "SHOULD_NOT_MATCH_FIRST"}
		}),
	)

	status, body := mapper.MapError(customError{msg: "boom"})
	if status != http.StatusTeapot {
		t.Errorf("status = %d, want %d", status, http.StatusTeapot)
	}
	if got := body.(map[string]string)["error"]; got != "boom" {
		t.Errorf("body error = %q, want %q", got, "boom")
	}
}

func TestNewErrorMapper_WrappedErrorStillMatches(t *testing.T) {
	mapper := NewErrorMapper(NewErrorMapping(func(err NotFound) (int, any) {
		return http.StatusNotFound, map[string]string{"code": "WRAPPED_MATCH"}
	}))

	wrapped := fmt.Errorf("lookup failed: %w", NotFound{Resource: "Book", ID: 1})
	status, body := mapper.MapError(wrapped)
	if status != http.StatusNotFound {
		t.Errorf("status = %d, want %d", status, http.StatusNotFound)
	}
	if got := body.(map[string]string)["code"]; got != "WRAPPED_MATCH" {
		t.Errorf("code = %q, want %q", got, "WRAPPED_MATCH")
	}
}

func TestNewErrorMapper_FallsBackToDefaultErrorMapper(t *testing.T) {
	mapper := NewErrorMapper(NewErrorMapping(func(err customError) (int, any) {
		return http.StatusTeapot, nil
	}))

	status, _ := mapper.MapError(NotFound{Resource: "Book", ID: 1})
	if status != http.StatusNotFound {
		t.Errorf("status = %d, want %d (DefaultErrorMapper fallback)", status, http.StatusNotFound)
	}
}

func TestComposedErrorMapper_UsesExplicitFallback(t *testing.T) {
	mapper := ComposedErrorMapper{
		Fallback: NewErrorMapper(NewErrorMapping(func(err error) (int, any) {
			return http.StatusTeapot, map[string]string{"code": "FALLBACK"}
		})),
	}

	status, body := mapper.MapError(errors.New("anything"))
	if status != http.StatusTeapot {
		t.Errorf("status = %d, want %d", status, http.StatusTeapot)
	}
	if got := body.(map[string]string)["code"]; got != "FALLBACK" {
		t.Errorf("code = %q, want %q", got, "FALLBACK")
	}
}

func TestRespond(t *testing.T) {
	w := httptest.NewRecorder()
	Respond(w, nil, NotFound{Resource: "Book", ID: 1})

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}

func TestRespondJSON(t *testing.T) {
	w := httptest.NewRecorder()
	RespondJSON(w, http.StatusCreated, map[string]string{"ok": "true"})

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}

	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["ok"] != "true" {
		t.Errorf(`body["ok"] = %q, want "true"`, body["ok"])
	}
}
