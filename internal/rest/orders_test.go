package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"FunPay-Core/internal/fp"
)

type stubOrderLister struct {
	result []OrderListItem
	err    error
}

func (s stubOrderLister) ListOrders(ctx context.Context) ([]OrderListItem, error) {
	return s.result, s.err
}

type stubOrderGetter struct {
	result OrderDetail
	err    error
}

func (s stubOrderGetter) GetOrder(ctx context.Context, orderID string) (OrderDetail, error) {
	return s.result, s.err
}

type stubOrderRefunder struct {
	calledOrderID string
	result        RefundedResult
	err           error
}

func (s *stubOrderRefunder) RefundOrder(ctx context.Context, orderID string) (RefundedResult, error) {
	s.calledOrderID = orderID
	return s.result, s.err
}

type stubChatMessager struct {
	calledChatID string
	calledText   string
	result       MessageSentResult
	err          error
}

func (s *stubChatMessager) SendChatMessage(ctx context.Context, node, text string) (MessageSentResult, error) {
	s.calledChatID = node
	s.calledText = text
	return s.result, s.err
}

func TestHandleOrdersListOK(t *testing.T) {
	srv := NewServer(nil, "secret")
	srv.SetOrderLister(stubOrderLister{result: []OrderListItem{{ID: "111", Status: "new"}}})

	req := httptest.NewRequest("GET", "/orders", nil)
	w := httptest.NewRecorder()
	srv.handleOrdersList(w, req)

	if w.Code != 200 {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	var got struct {
		Orders []OrderListItem `json:"orders"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got.Orders) != 1 || got.Orders[0].ID != "111" {
		t.Errorf("orders: got %+v", got.Orders)
	}
}

func TestHandleOrdersListEmptyIsArray(t *testing.T) {
	srv := NewServer(nil, "secret")
	srv.SetOrderLister(stubOrderLister{result: nil})

	req := httptest.NewRequest("GET", "/orders", nil)
	w := httptest.NewRecorder()
	srv.handleOrdersList(w, req)

	if w.Code != 200 {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"orders":[]`)) {
		t.Errorf("body should contain orders:[]; got %s", w.Body.String())
	}
}

func TestHandleOrdersListAuthLost(t *testing.T) {
	srv := NewServer(nil, "secret")
	srv.SetOrderLister(stubOrderLister{err: fp.ErrAuthLost})

	req := httptest.NewRequest("GET", "/orders", nil)
	w := httptest.NewRecorder()
	srv.handleOrdersList(w, req)

	if w.Code != 503 {
		t.Fatalf("status: got %d, want 503", w.Code)
	}
}

func TestHandleOrderDetailOK(t *testing.T) {
	srv := NewServer(nil, "secret")
	srv.SetOrderGetter(stubOrderGetter{result: OrderDetail{ID: "111", Status: "new", ChatID: "c222"}})

	req := httptest.NewRequest("GET", "/orders/111", nil)
	req.SetPathValue("id", "111")
	w := httptest.NewRecorder()
	srv.handleOrderDetail(w, req)

	if w.Code != 200 {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	var got OrderDetail
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.ID != "111" || got.ChatID != "c222" {
		t.Errorf("detail: got %+v", got)
	}
}

func TestHandleOrderDetailNotFound(t *testing.T) {
	srv := NewServer(nil, "secret")
	srv.SetOrderGetter(stubOrderGetter{err: fp.ErrOrderNotFound})

	req := httptest.NewRequest("GET", "/orders/999", nil)
	req.SetPathValue("id", "999")
	w := httptest.NewRecorder()
	srv.handleOrderDetail(w, req)

	if w.Code != 404 {
		t.Fatalf("status: got %d, want 404", w.Code)
	}
}

func TestHandleOrderRefundOK(t *testing.T) {
	srv := NewServer(nil, "secret")
	stub := &stubOrderRefunder{result: RefundedResult{OrderID: "WMBY8JNK"}}
	srv.SetOrderRefunder(stub)

	req := httptest.NewRequest("POST", "/orders/WMBY8JNK/refund", nil)
	req.SetPathValue("id", "WMBY8JNK")
	w := httptest.NewRecorder()
	srv.handleOrderRefund(w, req)

	if w.Code != 200 {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	if stub.calledOrderID != "WMBY8JNK" {
		t.Errorf("refund called with id=%q", stub.calledOrderID)
	}
	var got struct {
		Ok      bool   `json:"ok"`
		OrderID string `json:"orderId"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !got.Ok || got.OrderID != "WMBY8JNK" {
		t.Errorf("response: %+v", got)
	}
}

func TestHandleOrderRefundNotFound(t *testing.T) {
	srv := NewServer(nil, "secret")
	srv.SetOrderRefunder(&stubOrderRefunder{err: fp.ErrOrderNotFound})

	req := httptest.NewRequest("POST", "/orders/UNKNOWN/refund", nil)
	req.SetPathValue("id", "UNKNOWN")
	w := httptest.NewRecorder()
	srv.handleOrderRefund(w, req)

	if w.Code != 404 {
		t.Fatalf("status: got %d, want 404", w.Code)
	}
}

func TestHandleChatMessageOK(t *testing.T) {
	srv := NewServer(nil, "secret")
	stub := &stubChatMessager{result: MessageSentResult{MessageID: "m999"}}
	srv.SetChatMessager(stub)

	body, _ := json.Marshal(map[string]any{"text": "hello"})
	req := httptest.NewRequest("POST", "/chats/c222/messages", bytes.NewReader(body))
	req.SetPathValue("id", "c222")
	w := httptest.NewRecorder()
	srv.handleChatMessage(w, req)

	if w.Code != 200 {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	if stub.calledChatID != "c222" || stub.calledText != "hello" {
		t.Errorf("chat called with id=%q text=%q", stub.calledChatID, stub.calledText)
	}
}

func TestHandleChatMessageNoText(t *testing.T) {
	srv := NewServer(nil, "secret")
	srv.SetChatMessager(&stubChatMessager{})

	body, _ := json.Marshal(map[string]any{})
	req := httptest.NewRequest("POST", "/chats/c222/messages", bytes.NewReader(body))
	req.SetPathValue("id", "c222")
	w := httptest.NewRecorder()
	srv.handleChatMessage(w, req)

	if w.Code != 400 {
		t.Fatalf("status: got %d, want 400", w.Code)
	}
}

func TestHandleChatMessageChatNotFound(t *testing.T) {
	srv := NewServer(nil, "secret")
	srv.SetChatMessager(&stubChatMessager{err: fp.ErrChatNotFound})

	body, _ := json.Marshal(map[string]any{"text": "x"})
	req := httptest.NewRequest("POST", "/chats/unknown/messages", bytes.NewReader(body))
	req.SetPathValue("id", "unknown")
	w := httptest.NewRecorder()
	srv.handleChatMessage(w, req)

	if w.Code != 404 {
		t.Fatalf("status: got %d, want 404", w.Code)
	}
}
