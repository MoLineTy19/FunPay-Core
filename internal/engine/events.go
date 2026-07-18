package engine

import (
	"time"
)

type EventType string

const (
	OrderNew       EventType = "order.new"
	OrderCompleted EventType = "order.completed"
	OrderCancelled EventType = "order.cancelled"
	ChatMessage    EventType = "chat.message"
	OfferChanged   EventType = "offer.changed"
	AccountBalance EventType = "account.balance"
)

type Event struct {
	EventID   int64     `json:"eventId"`
	Type      EventType `json:"type"`
	At        time.Time `json:"at"`
	Payload   any       `json:"payload"`
	SourceRaw any       `json:"-"`
}
