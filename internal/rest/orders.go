package rest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"FunPay-Core/internal/fp"

	"github.com/shopspring/decimal"
)

type OrderListItem struct {
	ID        string          `json:"id"`
	Status    string          `json:"status"`
	BuyerName string          `json:"buyerName,omitempty"`
	Summary   string          `json:"summary,omitempty"`
	Price     decimal.Decimal `json:"price"`
	ChatID    string          `json:"chatId,omitempty"`
}

type OrderLister interface {
	ListOrders(ctx context.Context) ([]OrderListItem, error)
}

type OrderDetail struct {
	ID        string          `json:"id"`
	OfferID   string          `json:"offerId,omitempty"`
	NodeID    string          `json:"nodeId,omitempty"`
	BuyerID   int64           `json:"buyerId,omitempty"`
	BuyerName string          `json:"buyerName,omitempty"`
	Amount    decimal.Decimal `json:"amount,omitempty"`
	Currency  string          `json:"currency,omitempty"`
	Status    string          `json:"status"`
	CreatedAt string          `json:"createdAt,omitempty"`
	ChatID    string          `json:"chatId,omitempty"`
}

type OrderGetter interface {
	GetOrder(ctx context.Context, orderID string) (OrderDetail, error)
}

type MessageSentResult struct {
	MessageID string `json:"messageId,omitempty"`
}

type OrderDeliverer interface {
	DeliverOrder(ctx context.Context, orderID, text string) (MessageSentResult, error)
}

type MarkedReadyResult struct {
	OrderID string `json:"orderId"`
}

type OrderReadier interface {
	MarkOrderReady(ctx context.Context, orderID string) (MarkedReadyResult, error)
}

type ChatMessager interface {
	SendChatMessage(ctx context.Context, chatID, text string) (MessageSentResult, error)
}

func (s *Server) handleOrdersList(w http.ResponseWriter, r *http.Request) {
	if s.orderLister == nil {
		writeEngineError(w, http.StatusServiceUnavailable, "service_unavailable", "order lister not configured", false)
		return
	}
	items, err := s.orderLister.ListOrders(r.Context())
	if err != nil {
		if errors.Is(err, fp.ErrAuthLost) {
			writeEngineError(w, http.StatusServiceUnavailable, "auth_lost", err.Error(), false)
			return
		}
		writeEngineError(w, http.StatusInternalServerError, "internal", err.Error(), true)
		return
	}
	if items == nil {
		items = []OrderListItem{}
	}
	writeJSON(w, http.StatusOK, struct {
		Orders []OrderListItem `json:"orders"`
	}{Orders: items})
}

func (s *Server) handleOrderDetail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeEngineError(w, http.StatusBadRequest, "bad_request", "order id required", false)
		return
	}
	if s.orderGetter == nil {
		writeEngineError(w, http.StatusServiceUnavailable, "service_unavailable", "order getter not configured", false)
		return
	}
	d, err := s.orderGetter.GetOrder(r.Context(), id)
	if err != nil {
		if errors.Is(err, fp.ErrAuthLost) {
			writeEngineError(w, http.StatusServiceUnavailable, "auth_lost", err.Error(), false)
			return
		}
		if errors.Is(err, fp.ErrOrderNotFound) {
			writeEngineError(w, http.StatusNotFound, "order_not_found", err.Error(), false)
			return
		}
		writeEngineError(w, http.StatusInternalServerError, "internal", err.Error(), true)
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func (s *Server) handleOrderDeliver(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeEngineError(w, http.StatusBadRequest, "bad_request", "order id required", false)
		return
	}
	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeEngineError(w, http.StatusBadRequest, "bad_request", err.Error(), false)
		return
	}
	if req.Text == "" {
		writeEngineError(w, http.StatusBadRequest, "bad_request", "text required", false)
		return
	}
	if s.orderDeliverer == nil {
		writeEngineError(w, http.StatusServiceUnavailable, "service_unavailable", "order deliverer not configured", false)
		return
	}
	res, err := s.orderDeliverer.DeliverOrder(r.Context(), id, req.Text)
	if err != nil {
		if errors.Is(err, fp.ErrAuthLost) {
			writeEngineError(w, http.StatusServiceUnavailable, "auth_lost", err.Error(), false)
			return
		}
		if errors.Is(err, fp.ErrOrderNotFound) {
			writeEngineError(w, http.StatusNotFound, "order_not_found", err.Error(), false)
			return
		}
		if errors.Is(err, fp.ErrCannotDeliver) {
			writeEngineError(w, http.StatusUnprocessableEntity, "cannot_deliver", err.Error(), false)
			return
		}
		writeEngineError(w, http.StatusInternalServerError, "internal", err.Error(), true)
		return
	}
	writeJSON(w, http.StatusOK, struct {
		Ok        bool   `json:"ok"`
		MessageID string `json:"messageId,omitempty"`
	}{Ok: true, MessageID: res.MessageID})
}

func (s *Server) handleOrderReady(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeEngineError(w, http.StatusBadRequest, "bad_request", "order id required", false)
		return
	}
	if s.orderReadier == nil {
		writeEngineError(w, http.StatusServiceUnavailable, "service_unavailable", "order readier not configured", false)
		return
	}
	res, err := s.orderReadier.MarkOrderReady(r.Context(), id)
	if err != nil {
		if errors.Is(err, fp.ErrAuthLost) {
			writeEngineError(w, http.StatusServiceUnavailable, "auth_lost", err.Error(), false)
			return
		}
		if errors.Is(err, fp.ErrOrderNotFound) {
			writeEngineError(w, http.StatusNotFound, "order_not_found", err.Error(), false)
			return
		}
		writeEngineError(w, http.StatusInternalServerError, "internal", err.Error(), true)
		return
	}
	writeJSON(w, http.StatusOK, struct {
		Ok      bool   `json:"ok"`
		OrderID string `json:"orderId"`
	}{Ok: true, OrderID: res.OrderID})
}

func (s *Server) handleChatMessage(w http.ResponseWriter, r *http.Request) {
	chatID := r.PathValue("id")
	if chatID == "" {
		writeEngineError(w, http.StatusBadRequest, "bad_request", "chat id required", false)
		return
	}
	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeEngineError(w, http.StatusBadRequest, "bad_request", err.Error(), false)
		return
	}
	if req.Text == "" {
		writeEngineError(w, http.StatusBadRequest, "bad_request", "text required", false)
		return
	}
	if s.chatMessager == nil {
		writeEngineError(w, http.StatusServiceUnavailable, "service_unavailable", "chat messager not configured", false)
		return
	}
	res, err := s.chatMessager.SendChatMessage(r.Context(), chatID, req.Text)
	if err != nil {
		if errors.Is(err, fp.ErrAuthLost) {
			writeEngineError(w, http.StatusServiceUnavailable, "auth_lost", err.Error(), false)
			return
		}
		if errors.Is(err, fp.ErrChatNotFound) {
			writeEngineError(w, http.StatusNotFound, "chat_not_found", err.Error(), false)
			return
		}
		writeEngineError(w, http.StatusInternalServerError, "internal", err.Error(), true)
		return
	}
	writeJSON(w, http.StatusOK, struct {
		Ok        bool   `json:"ok"`
		MessageID string `json:"messageId,omitempty"`
	}{Ok: true, MessageID: res.MessageID})
}