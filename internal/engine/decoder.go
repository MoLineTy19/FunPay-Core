package engine

import "FunPay-Core/internal/fp"

func WrapEvents(ev fp.RunnerEvents) []Event {
	events := make([]Event, 0, len(ev.Messages))
	for _, m := range ev.Messages {
		events = append(events, Event{
			Type:    ChatMessage,
			Payload: m,
			At:      m.CreatedAt,
		})
	}
	return events
}
