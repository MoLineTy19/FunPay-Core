package engine

import (
	"testing"
	"time"

	"FunPay-Core/internal/fp"
)

func TestWrapEvents(t *testing.T) {
	in := fp.RunnerEvents{
		Messages: []fp.ChatMessage{
			{ID: "111", ChatID: "222", Text: "привет", Author: fp.AuthorBuyer, CreatedAt: time.Now()},
			{ID: "333", ChatID: "444", Text: "как дела", Author: fp.AuthorBuyer, CreatedAt: time.Now()},
		},
	}

	out := WrapEvents(in)

	if len(out) != 2 {
		t.Fatalf("len: got %d, want 2", len(out))
	}

	for i, e := range out {
		if e.Type != ChatMessage {
			t.Errorf("[%d] Type: got %q, want %q", i, e.Type, ChatMessage)
		}
		if e.EventID != 0 {
			t.Errorf("[%d] EventID: got %d, want 0 (decoder must not assign ids)", i, e.EventID)
		}
		msg, ok := e.Payload.(fp.ChatMessage)
		if !ok {
			t.Errorf("[%d] Payload is not fp.ChatMessage: %T", i, e.Payload)
			continue
		}
		if msg.Text != in.Messages[i].Text {
			t.Errorf("[%d] Text: got %q, want %q", i, msg.Text, in.Messages[i].Text)
		}
	}
}

func TestWrapEventsOrders(t *testing.T) {
	in := fp.RunnerEvents{
		Orders: []fp.OrderEvent{
			{Kind: fp.OrderEventNew, Order: fp.OrderShortcut{ID: "111"}, ToStatus: fp.StatusNew},
			{Kind: fp.OrderEventCompleted, Order: fp.OrderShortcut{ID: "222"}, FromStatus: fp.StatusNew, ToStatus: fp.StatusCompleted},
			{Kind: fp.OrderEventCancelled, Order: fp.OrderShortcut{ID: "333"}, FromStatus: fp.StatusNew, ToStatus: fp.StatusCancelled},
		},
	}

	out := WrapEvents(in)

	if len(out) != 3 {
		t.Fatalf("len: got %d, want 3", len(out))
	}
	wantTypes := []EventType{OrderNew, OrderCompleted, OrderCancelled}
	for i, e := range out {
		if e.Type != wantTypes[i] {
			t.Errorf("[%d] Type: got %q, want %q", i, e.Type, wantTypes[i])
		}
		os, ok := e.Payload.(fp.OrderShortcut)
		if !ok {
			t.Errorf("[%d] Payload is not fp.OrderShortcut: %T", i, e.Payload)
			continue
		}
		wantID := []string{"111", "222", "333"}[i]
		if os.ID != wantID {
			t.Errorf("[%d] ID: got %q, want %q", i, os.ID, wantID)
		}
	}
}
