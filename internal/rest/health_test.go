package rest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"FunPay-Core/internal/engine"
)

func TestHandleHealth(t *testing.T) {
	srv := NewServer(nil, "")
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	srv.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status code: got %d, want 200", w.Code)
	}
	var got healthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Status != "healthy" {
		t.Fatalf("status: got %q, want \"healthy\"", got.Status)
	}
}

func TestHandleHealthAuthLost(t *testing.T) {
	srv := NewServer(nil, "")
	srv.SetState("auth_lost")

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	srv.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status code: got %d, want 200", w.Code)
	}
	var got healthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Status != "auth_lost" {
		t.Fatalf("status: got %q, want \"auth_lost\"", got.Status)
	}
}

func TestHandleHealthEnriched(t *testing.T) {
	buf := engine.NewBuffer()
	buf.Push([]engine.Event{{Type: engine.ChatMessage}})
	buf.Push([]engine.Event{{Type: engine.ChatMessage}})

	srv := NewServer(buf, "")
	// Искусственно состарим startedAt, чтобы uptime был ненулевым и предсказуемым.
	srv.startedAt = time.Now().Add(-90 * time.Second)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	srv.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status code: got %d, want 200", w.Code)
	}
	var got healthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.EventsBuffered != 2 {
		t.Errorf("eventsBuffered: got %d, want 2", got.EventsBuffered)
	}
	if got.Uptime == "" {
		t.Errorf("uptime: got empty, want non-empty")
	}
}
