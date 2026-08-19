package goninja

import (
	"encoding/json"
	"errors"
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
		{"unknown error", errors.New("boom"), http.StatusInternalServerError, "INTERNAL"},
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
