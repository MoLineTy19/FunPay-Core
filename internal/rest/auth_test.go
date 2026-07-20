package rest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthMiddlewareNoToken(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	h := authMiddleware("secret", next)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if called {
		t.Fatal("next handler called without token")
	}
	if w.Code != 401 {
		t.Fatalf("status: got %d, want 401", w.Code)
	}
	if !strings.Contains(w.Body.String(), "unauthorized") {
		t.Fatalf("body: want code 'unauthorized', got %q", w.Body.String())
	}
}

func TestAuthMiddlewareWrongToken(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	h := authMiddleware("secret", next)

	req := httptest.NewRequest("GET", "/health", nil)
	req.Header.Set("X-Engine-Token", "wrong")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if called {
		t.Fatal("next handler called with wrong token")
	}
	if w.Code != 401 {
		t.Fatalf("status: got %d, want 401", w.Code)
	}
}

func TestAuthMiddlewareOK(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(200)
	})
	h := authMiddleware("secret", next)

	req := httptest.NewRequest("GET", "/health", nil)
	req.Header.Set("X-Engine-Token", "secret")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if !called {
		t.Fatal("next handler not called with valid token")
	}
	if w.Code != 200 {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
}
