package rest

import (
	"context"
	"net/http"

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

func (s *Server) handleOrdersList(w http.ResponseWriter, r *http.Request)  {}
func (s *Server) handleOrderDetail(w http.ResponseWriter, r *http.Request) {}
func (s *Server) handleOrderDeliver(w http.ResponseWriter, r *http.Request) {
}
func (s *Server) handleOrderReady(w http.ResponseWriter, r *http.Request) {}
func (s *Server) handleChatMessage(w http.ResponseWriter, r *http.Request) {}