package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"FunPay-Core/internal/fp"

	"github.com/shopspring/decimal"
)

type stubOfferCreator struct {
	result OfferCreated
	err    error
}

func (s stubOfferCreator) CreateOffer(ctx context.Context, nodeID, serverID string, fields map[string]map[string]string, price decimal.Decimal, amount int, active bool) (OfferCreated, error) {
	return s.result, s.err
}

func TestHandleOffersCreateOK(t *testing.T) {
	srv := NewServer(nil, "secret")
	srv.SetOfferCreator(stubOfferCreator{
		result: OfferCreated{NodeID: "791", OfferID: "73311257", URL: "https://funpay.com/lots/791/trade"},
	})

	body := map[string]any{
		"nodeId":   "791",
		"serverId": "5188",
		"fields":   map[string]map[string]string{"summary": {"ru": "Test Lot"}, "level": {"ru": "111"}},
		"price":    "111",
		"active":   true,
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
	srv.SetOfferCreator(stubOfferCreator{})

	cases := []map[string]any{
		{"serverId": "5188", "fields": map[string]map[string]string{"summary": {"ru": "x"}}, "price": "1"}, // нет nodeId
		{"nodeId": "791", "price": "1"}, // нет serverId и fields
		{"nodeId": "791", "serverId": "5188", "fields": map[string]map[string]string{"level": {"ru": "1"}}, "price": "1"},    // нет fields.summary
		{"nodeId": "791", "serverId": "5188", "fields": map[string]map[string]string{"summary": {"ru": "x"}}, "price": "-1"}, // отрицательная цена
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
	srv.SetOfferCreator(stubOfferCreator{err: fp.ErrAuthLost})

	body, _ := json.Marshal(map[string]any{"nodeId": "791", "serverId": "5188", "fields": map[string]map[string]string{"summary": {"ru": "x"}}, "price": "1"})
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

	body, _ := json.Marshal(map[string]any{"nodeId": "791", "serverId": "5188", "fields": map[string]map[string]string{"summary": {"ru": "x"}}, "price": "1"})
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

// --- stubs for new interfaces ---

type stubOfferEditor struct {
	result OfferUpdated
	err    error
}

func (s stubOfferEditor) EditOffer(ctx context.Context, nodeID, offerID string, fields map[string]map[string]string, price *decimal.Decimal, amount *int, active *bool) (OfferUpdated, error) {
	return s.result, s.err
}

type stubOfferDeleter struct {
	result OfferDeleted
	err    error
}

func (s stubOfferDeleter) DeleteOffer(ctx context.Context, nodeID, offerID string) (OfferDeleted, error) {
	return s.result, s.err
}

type stubOfferLister struct {
	result []OfferListItem
	err    error
}

func (s stubOfferLister) ListOffers(ctx context.Context, nodeID string) ([]OfferListItem, error) {
	return s.result, s.err
}

type stubOfferFormGetter struct {
	result OfferForm
	err    error
}

func (s stubOfferFormGetter) GetOfferForm(ctx context.Context, nodeID string) (OfferForm, error) {
	return s.result, s.err
}

// --- PATCH /offers/{node}/{offer} ---

func TestHandleOffersUpdateOK(t *testing.T) {
	srv := NewServer(nil, "secret")
	srv.SetOfferEditor(stubOfferEditor{result: OfferUpdated{NodeID: "791", OfferID: "73311257", URL: "https://funpay.com/lots/791/trade"}})

	body, _ := json.Marshal(map[string]any{"fields": map[string]map[string]string{"summary": {"ru": "NEW"}}, "price": "200"})
	req := httptest.NewRequest("PATCH", "/offers/791/73311257", bytes.NewReader(body))
	req.SetPathValue("node", "791")
	req.SetPathValue("offer", "73311257")
	w := httptest.NewRecorder()
	srv.handleOffersUpdate(w, req)

	if w.Code != 200 {
		t.Fatalf("status: got %d, want 200 (body=%s)", w.Code, w.Body.String())
	}
	var got updateOfferResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !got.Updated || got.OfferID != "73311257" {
		t.Errorf("response: %+v", got)
	}
}

func TestHandleOffersUpdateNothingToUpdate(t *testing.T) {
	srv := NewServer(nil, "secret")
	srv.SetOfferEditor(stubOfferEditor{})

	body, _ := json.Marshal(map[string]any{})
	req := httptest.NewRequest("PATCH", "/offers/791/73311257", bytes.NewReader(body))
	req.SetPathValue("node", "791")
	req.SetPathValue("offer", "73311257")
	w := httptest.NewRecorder()
	srv.handleOffersUpdate(w, req)

	if w.Code != 400 {
		t.Fatalf("status: got %d, want 400", w.Code)
	}
}

func TestHandleOffersUpdateAuthLost(t *testing.T) {
	srv := NewServer(nil, "secret")
	srv.SetOfferEditor(stubOfferEditor{err: fp.ErrAuthLost})

	body, _ := json.Marshal(map[string]any{"price": "1"})
	req := httptest.NewRequest("PATCH", "/offers/791/73311257", bytes.NewReader(body))
	req.SetPathValue("node", "791")
	req.SetPathValue("offer", "73311257")
	w := httptest.NewRecorder()
	srv.handleOffersUpdate(w, req)

	if w.Code != 503 {
		t.Fatalf("status: got %d, want 503", w.Code)
	}
}

func TestHandleOffersUpdateNotFound(t *testing.T) {
	srv := NewServer(nil, "secret")
	srv.SetOfferEditor(stubOfferEditor{err: fp.ErrOfferNotFound})

	body, _ := json.Marshal(map[string]any{"price": "1"})
	req := httptest.NewRequest("PATCH", "/offers/791/999", bytes.NewReader(body))
	req.SetPathValue("node", "791")
	req.SetPathValue("offer", "999")
	w := httptest.NewRecorder()
	srv.handleOffersUpdate(w, req)

	if w.Code != 404 {
		t.Fatalf("status: got %d, want 404", w.Code)
	}
}

// --- DELETE /offers/{node}/{offer} ---

func TestHandleOffersDeleteOK(t *testing.T) {
	srv := NewServer(nil, "secret")
	srv.SetOfferDeleter(stubOfferDeleter{result: OfferDeleted{NodeID: "791", OfferID: "73311257"}})

	req := httptest.NewRequest("DELETE", "/offers/791/73311257", nil)
	req.SetPathValue("node", "791")
	req.SetPathValue("offer", "73311257")
	w := httptest.NewRecorder()
	srv.handleOffersDelete(w, req)

	if w.Code != 200 {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	var got deleteOfferResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !got.Deleted || got.OfferID != "73311257" {
		t.Errorf("response: %+v", got)
	}
}

func TestHandleOffersDeleteNotFound(t *testing.T) {
	srv := NewServer(nil, "secret")
	srv.SetOfferDeleter(stubOfferDeleter{err: fp.ErrOfferNotFound})

	req := httptest.NewRequest("DELETE", "/offers/791/999", nil)
	req.SetPathValue("node", "791")
	req.SetPathValue("offer", "999")
	w := httptest.NewRecorder()
	srv.handleOffersDelete(w, req)

	if w.Code != 404 {
		t.Fatalf("status: got %d, want 404", w.Code)
	}
}

// --- GET /offers/{node} ---

func TestHandleOffersListOK(t *testing.T) {
	srv := NewServer(nil, "secret")
	price, _ := decimal.NewFromString("111111")
	srv.SetOfferLister(stubOfferLister{result: []OfferListItem{
		{OfferID: "73311257", Summary: "lot 1", Server: "Android", Amount: "1", Price: price},
	}})

	req := httptest.NewRequest("GET", "/offers/791", nil)
	req.SetPathValue("node", "791")
	w := httptest.NewRecorder()
	srv.handleOffersList(w, req)

	if w.Code != 200 {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	var got listOffersResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got.Offers) != 1 || got.Offers[0].OfferID != "73311257" {
		t.Errorf("offers: %+v", got)
	}
}

func TestHandleOffersListEmpty(t *testing.T) {
	srv := NewServer(nil, "secret")
	srv.SetOfferLister(stubOfferLister{result: nil})

	req := httptest.NewRequest("GET", "/offers/791", nil)
	req.SetPathValue("node", "791")
	w := httptest.NewRecorder()
	srv.handleOffersList(w, req)

	if w.Code != 200 {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"offers":[]`) {
		t.Errorf("body should contain offers:[] (not null), got %s", body)
	}
}

// --- GET /offers/form?node=X ---

func TestHandleOffersFormOK(t *testing.T) {
	srv := NewServer(nil, "secret")
	srv.SetOfferFormGetter(stubOfferFormGetter{result: OfferForm{
		NodeID:  "791",
		Fields:  []OfferFormField{{ID: "summary", Type: 2}},
		Servers: []OfferServer{{ID: "5188", Name: "Android"}},
	}})

	req := httptest.NewRequest("GET", "/offers/form?node=791", nil)
	w := httptest.NewRecorder()
	srv.handleOffersForm(w, req)

	if w.Code != 200 {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	var got OfferForm
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got.Fields) != 1 || len(got.Servers) != 1 {
		t.Errorf("form: %+v", got)
	}
}

func TestHandleOffersFormMissingNode(t *testing.T) {
	srv := NewServer(nil, "secret")
	srv.SetOfferFormGetter(stubOfferFormGetter{})

	req := httptest.NewRequest("GET", "/offers/form", nil)
	w := httptest.NewRecorder()
	srv.handleOffersForm(w, req)

	if w.Code != 400 {
		t.Fatalf("status: got %d, want 400", w.Code)
	}
}
