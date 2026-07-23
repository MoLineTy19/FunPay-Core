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

type RefundedResult struct {
	OrderID string `json:"orderId"`
}

type OrderRefunder interface {
	RefundOrder(ctx context.Context, orderID string) (RefundedResult, error)
}

type ChatMessager interface {
	SendChatMessage(ctx context.Context, node, text string) (MessageSentResult, error)
}

// @Summary      Список продаж
// @Description  Возвращает список заказов продавца. Пустой список — [] (не null).
// @Tags         orders
// @Produce      json
// @Success      200  {object}  OrdersListResponse
// @Failure      401  {object}  EngineError  "missing or invalid token"
// @Failure      503  {object}  EngineError  "auth_lost / service_unavailable"
// @Failure      500  {object}  EngineError  "internal (retryable)"
// @Security     ApiKeyAuth
// @Router       /orders [get]
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
	writeJSON(w, http.StatusOK, OrdersListResponse{Orders: items})
}

// @Summary      Детали заказа
// @Description  Полная информация по одному заказу по его ID.
// @Tags         orders
// @Produce      json
// @Param        id   path      string  true  "ID заказа (например WMBY8JNK)"
// @Success      200  {object}  OrderDetail
// @Failure      400  {object}  EngineError  "bad_request"
// @Failure      401  {object}  EngineError  "missing or invalid token"
// @Failure      404  {object}  EngineError  "order_not_found"
// @Failure      503  {object}  EngineError  "auth_lost / service_unavailable"
// @Failure      500  {object}  EngineError  "internal (retryable)"
// @Security     ApiKeyAuth
// @Router       /orders/{id} [get]
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

// @Summary      Вернуть заказ
// @Description  Инициирует возврат заказа. Тела нет.
// @Tags         orders
// @Produce      json
// @Param        id   path      string  true  "ID заказа"
// @Success      200  {object}  OrderRefundResponse
// @Failure      400  {object}  EngineError  "bad_request"
// @Failure      401  {object}  EngineError  "missing or invalid token"
// @Failure      404  {object}  EngineError  "order_not_found"
// @Failure      503  {object}  EngineError  "auth_lost / service_unavailable"
// @Failure      500  {object}  EngineError  "internal (retryable)"
// @Security     ApiKeyAuth
// @Router       /orders/{id}/refund [post]
func (s *Server) handleOrderRefund(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeEngineError(w, http.StatusBadRequest, "bad_request", "order id required", false)
		return
	}
	if s.orderRefunder == nil {
		writeEngineError(w, http.StatusServiceUnavailable, "service_unavailable", "order refunder not configured", false)
		return
	}
	res, err := s.orderRefunder.RefundOrder(r.Context(), id)
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
	writeJSON(w, http.StatusOK, OrderRefundResponse{Ok: true, OrderID: res.OrderID})
}

// @Summary      Ответить в чат
// @Description  Отправляет сообщение покупателю в чат. Параметр {id} — это node чата вида users-{buyerId}-{sellerId} (берётся из chatId заказа или события), а не числовой chatId.
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        id       path      string              true  "node чата (users-{buyerId}-{sellerId})"
// @Param        request  body      ChatMessageRequest  true  "Текст сообщения"
// @Success      200      {object}  ChatMessageResponse
// @Failure      400      {object}  EngineError  "bad_request"
// @Failure      401      {object}  EngineError  "missing or invalid token"
// @Failure      404      {object}  EngineError  "chat_not_found"
// @Failure      503      {object}  EngineError  "auth_lost / service_unavailable"
// @Failure      500      {object}  EngineError  "internal (retryable)"
// @Security     ApiKeyAuth
// @Router       /chats/{id}/messages [post]
func (s *Server) handleChatMessage(w http.ResponseWriter, r *http.Request) {
	node := r.PathValue("id")
	if node == "" {
		writeEngineError(w, http.StatusBadRequest, "bad_request", "chat node required", false)
		return
	}
	var req ChatMessageRequest
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
	res, err := s.chatMessager.SendChatMessage(r.Context(), node, req.Text)
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
	writeJSON(w, http.StatusOK, ChatMessageResponse{Ok: true, MessageID: res.MessageID})
}
