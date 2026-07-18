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
