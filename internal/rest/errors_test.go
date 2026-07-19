package rest

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	body := map[string]string{"status": "healthy"}
	writeJSON(w, 200, body)

	if w.Code != 200 {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type: got %q, want application/json", ct)
	}
	var got map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["status"] != "healthy" {
		t.Fatalf("body.status: got %q, want healthy", got["status"])
	}
}

func TestWriteEngineError(t *testing.T) {
	w := httptest.NewRecorder()
	writeEngineError(w, 409, "cursor_too_old", "events evicted", false)

	if w.Code != 409 {
		t.Fatalf("status: got %d, want 409", w.Code)
	}
	var got engineErrorJSON
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Error.Code != "cursor_too_old" {
		t.Errorf("code: got %q, want cursor_too_old", got.Error.Code)
	}
	if got.Error.Message != "events evicted" {
		t.Errorf("message: got %q, want 'events evicted'", got.Error.Message)
	}
	if got.Error.Retryable != false {
		t.Errorf("retryable: got %v, want false", got.Error.Retryable)
	}
}
