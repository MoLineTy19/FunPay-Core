package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"FunPay-Core/internal/fp"

	"github.com/shopspring/decimal"
)

type stubOfferCreator struct {
	result OfferCreated
	err    error
}

func (s stubOfferCreator) CreateOffer(ctx context.Context, csrf, nodeID string, fields map[string]string, price decimal.Decimal, amount int, active bool) (OfferCreated, error) {
	return s.result, s.err
}

func TestHandleOffersCreateOK(t *testing.T) {
	srv := NewServer(nil, "secret")
	srv.SetOfferCreator(stubOfferCreator{
		result: OfferCreated{NodeID: "791", OfferID: "73311257", URL: "https://funpay.com/lots/791/trade"},
	}, "csrf-token")

	body := map[string]any{
		"nodeId": "791",
		"fields": map[string]string{"summary": "Test Lot", "level": "111"},
		"price":  "111",
		"active": true,
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/offers", bytes.NewReader(b))
	w := httptest.NewRecorder()
	srv.handleOffersCreate(w, req)

	if w.Code != 201 {
		t.Fatalf("status: got %d, want 201", w.Code)
	}
	var got createOfferResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.OfferID != "73311257" {
		t.Errorf("offerId: got %q", got.OfferID)
	}
	if got.NodeID != "791" {
		t.Errorf("nodeId: got %q", got.NodeID)
	}
	if !got.Created {
		t.Errorf("created: got false")
	}
}

func TestHandleOffersCreateBadInput(t *testing.T) {
	srv := NewServer(nil, "secret")
	srv.SetOfferCreator(stubOfferCreator{}, "csrf")

	cases := []map[string]any{
		{"fields": map[string]string{"summary": "x"}, "price": "1"},
		{"nodeId": "791", "price": "1"},
		{"nodeId": "791", "fields": map[string]string{"level": "1"}, "price": "1"},
		{"nodeId": "791", "fields": map[string]string{"summary": "x"}, "price": "-1"},
	}
	for i, body := range cases {
		b, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/offers", bytes.NewReader(b))
		w := httptest.NewRecorder()
		srv.handleOffersCreate(w, req)
		if w.Code != 400 {
			t.Errorf("case %d: status: got %d, want 400 (body=%s)", i, w.Code, b)
		}
	}
}

func TestHandleOffersCreateAuthLost(t *testing.T) {
	srv := NewServer(nil, "secret")
	srv.SetOfferCreator(stubOfferCreator{err: fp.ErrAuthLost}, "csrf")

	body, _ := json.Marshal(map[string]any{"nodeId": "791", "fields": map[string]string{"summary": "x"}, "price": "1"})
	req := httptest.NewRequest("POST", "/offers", bytes.NewReader(body))
	w := httptest.NewRecorder()
	srv.handleOffersCreate(w, req)

	if w.Code != 503 {
		t.Fatalf("status: got %d, want 503", w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("auth_lost")) {
		t.Errorf("body: want 'auth_lost', got %q", w.Body.String())
	}
}

func TestHandleOffersCreateNotConfigured(t *testing.T) {
	srv := NewServer(nil, "secret")

	body, _ := json.Marshal(map[string]any{"nodeId": "791", "fields": map[string]string{"summary": "x"}, "price": "1"})
	req := httptest.NewRequest("POST", "/offers", bytes.NewReader(body))
	w := httptest.NewRecorder()
	srv.handleOffersCreate(w, req)

	if w.Code != 503 {
		t.Fatalf("status: got %d, want 503", w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("service_unavailable")) {
		t.Errorf("body: want 'service_unavailable', got %q", w.Body.String())
	}
}
