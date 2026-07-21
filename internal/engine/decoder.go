package engine

import (
	"time"

	"FunPay-Core/internal/fp"
)

func WrapEvents(ev fp.RunnerEvents) []Event {
	events := make([]Event, 0, len(ev.Messages)+len(ev.Orders))
	for _, m := range ev.Messages {
		events = append(events, Event{
			Type:    ChatMessage,
			Payload: m,
			At:      m.CreatedAt,
		})
	}
	for _, o := range ev.Orders {
		var t EventType
		switch o.Kind {
		case fp.OrderEventNew:
			t = OrderNew
		case fp.OrderEventCompleted:
			t = OrderCompleted
		case fp.OrderEventCancelled:
			t = OrderCancelled
		}
		events = append(events, Event{
			Type:    t,
			Payload: o.Order,
			At:      time.Now(),
		})
	}
	return events
}
