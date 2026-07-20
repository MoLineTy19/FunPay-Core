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
	EngineStatus   EventType = "engine.status"
)

type Event struct {
	EventID   int64     `json:"eventId"`
	Type      EventType `json:"type"`
	At        time.Time `json:"at"`
	Payload   any       `json:"payload"`
	SourceRaw any       `json:"-"`
}

type EngineStatusPayload struct {
	State string `json:"state"`
	Error string `json:"error,omitempty"`
}

type AccountBalancePayload struct {
	UserID  int64  `json:"userId"`
	Login   string `json:"login"`
	Balance string `json:"balance"`
}
