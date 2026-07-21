package rest

import (
	"net/http/httptest"
	"testing"
)

func TestHandleControlResumeOK(t *testing.T) {
	srv := NewServer(nil, "secret")
	srv.SetState("auth_lost")
	ch := make(chan struct{}, 1)
	srv.SetResumeCh(ch)

	req := httptest.NewRequest("POST", "/control/resume", nil)
	w := httptest.NewRecorder()
	srv.handleControlResume(w, req)

	if w.Code != 202 {
		t.Fatalf("status: got %d, want 202 (body=%s)", w.Code, w.Body.String())
	}

	select {
	case <-ch:
	default:
		t.Errorf("resume signal not sent to chan")
	}
}

func TestHandleControlResumeConflictWhenHealthy(t *testing.T) {
	srv := NewServer(nil, "secret")
	ch := make(chan struct{}, 1)
	srv.SetResumeCh(ch)

	req := httptest.NewRequest("POST", "/control/resume", nil)
	w := httptest.NewRecorder()
	srv.handleControlResume(w, req)

	if w.Code != 409 {
		t.Fatalf("status: got %d, want 409 (body=%s)", w.Code, w.Body.String())
	}

	select {
	case <-ch:
		t.Errorf("signal should not be sent when state != auth_lost")
	default:
	}
}

func TestHandleControlResumeNoChannel(t *testing.T) {
	srv := NewServer(nil, "secret")
	srv.SetState("auth_lost")

	req := httptest.NewRequest("POST", "/control/resume", nil)
	w := httptest.NewRecorder()
	srv.handleControlResume(w, req)

	if w.Code != 503 {
		t.Fatalf("status: got %d, want 503 (body=%s)", w.Code, w.Body.String())
	}
}

// TestHandleControlResumeNonBlocking: повторный POST без потребителя не блокирует handler.
func TestHandleControlResumeNonBlocking(t *testing.T) {
	srv := NewServer(nil, "secret")
	srv.SetState("auth_lost")

	ch := make(chan struct{}, 1)
	srv.SetResumeCh(ch)

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("POST", "/control/resume", nil)
		w := httptest.NewRecorder()
		srv.handleControlResume(w, req)
		if w.Code != 202 {
			t.Fatalf("POST #%d: status %d, want 202 (body=%s)", i, w.Code, w.Body.String())
		}
	}

	if len(ch) != 1 {
		t.Errorf("chan should have at most 1 signal, got %d", len(ch))
	}
}
