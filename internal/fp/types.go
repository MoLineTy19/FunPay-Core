package fp

import (
	"time"

	"github.com/shopspring/decimal"
)

type Account struct {
	UserID  int64
	Login   string
	Balance decimal.Decimal
}

type Offer struct {
	ID       string
	NodeID   string
	Title    string
	Price    decimal.Decimal
	Currency string
	Active   bool
	Stock    int
}

type Status string

const (
	StatusNew       Status = "new"
	StatusCompleted Status = "completed"
	StatusCancelled Status = "cancelled"
)

type Order struct {
	ID        string
	OfferID   string
	NodeID    string
	BuyerName string
	BuyerID   int64
	Amount    decimal.Decimal
	Currency  string
	Status    Status
	CreatedAt time.Time
	ChatID    string
}
type Chat struct {
	ID         string
	BuyerID    int64
	NodeID     string
	LastReadAt time.Time
}

type Author string

const (
	AuthorBuyer  Author = "buyer"
	AuthorSeller Author = "seller"
	AuthorSystem Author = "system"
)

type ChatMessage struct {
	ID        string
	ChatID    string
	Author    Author
	Text      string
	CreatedAt time.Time
}
