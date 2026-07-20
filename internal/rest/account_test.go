package rest

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandleAccountDefault(t *testing.T) {
	srv := NewServer(nil, "secret")
	req := httptest.NewRequest("GET", "/account", nil)
	w := httptest.NewRecorder()
	srv.handleAccount(w, req)

	if w.Code != 200 {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	var got AccountSnapshot
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Balance != "" {
		t.Errorf("balance: got %q, want empty (no SetAccount called)", got.Balance)
	}
	if !got.LoadedAt.IsZero() {
		t.Errorf("loadedAt: got %v, want zero (no SetAccount called)", got.LoadedAt)
	}
}

func TestHandleAccountSet(t *testing.T) {
	srv := NewServer(nil, "secret")
	want := AccountSnapshot{
		UserID:   12345,
		Login:    "seller_login",
		Balance:  "1234.56",
		LoadedAt: time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC),
	}
	srv.SetAccount(want)

	req := httptest.NewRequest("GET", "/account", nil)
	w := httptest.NewRecorder()
	srv.handleAccount(w, req)

	if w.Code != 200 {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	var got AccountSnapshot
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.UserID != want.UserID {
		t.Errorf("userId: got %d, want %d", got.UserID, want.UserID)
	}
	if got.Login != want.Login {
		t.Errorf("login: got %q, want %q", got.Login, want.Login)
	}
	if got.Balance != want.Balance {
		t.Errorf("balance: got %q, want %q", got.Balance, want.Balance)
	}
	if !got.LoadedAt.Equal(want.LoadedAt) {
		t.Errorf("loadedAt: got %v, want %v", got.LoadedAt, want.LoadedAt)
	}
}
