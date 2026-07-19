package engine

import (
	"errors"
	"testing"
	"time"
)

func TestBufferPushAssignsMonotonicIDs(t *testing.T) {
	b := NewBuffer()

	in := []Event{
		{Type: ChatMessage, Payload: "a"},
		{Type: ChatMessage, Payload: "b"},
		{Type: ChatMessage, Payload: "c"},
	}
	b.Push(in)

	// 1. Сколько событий в буфере?
	if got := len(b.events); got != 3 {
		t.Fatalf("len = %d, want 3", got)
	}

	// 2. EventID должны быть 1, 2, 3 по порядку.
	wantIDs := []int64{1, 2, 3}
	for i, want := range wantIDs {
		if got := b.events[i].event.EventID; got != want {
			t.Errorf("events[%d].EventID = %d, want %d", i, got, want)
		}
	}

	// 3. Счётчик после трёх Push должен указывать на следующий — 4.
	if b.next != 4 {
		t.Errorf("next = %d, want 4", b.next)
	}
}

func TestBufferSince(t *testing.T) {
	b := NewBuffer()
	b.Push([]Event{{}, {}, {}})

	got, err := b.Since(2)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].EventID != 3 {
		t.Fatalf("EventID = %d, want 3", got[0].EventID)
	}

	bEmpty := NewBuffer()
	got, err = bEmpty.Since(0)
	if err != nil || got != nil {
		t.Fatalf("empty: got=%v, err=%v", got, err)
	}

	bAll := NewBuffer()
	bAll.Push([]Event{{}, {}, {}})
	got, err = bAll.Since(0)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}

	bGrow := NewBuffer()
	bGrow.Push([]Event{{}, {}, {}})
	bGrow.Push([]Event{{}, {}})
	got, err = bGrow.Since(2)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("len=%d, want 3", len(got))
	}

	bAhead := NewBuffer()
	bAhead.Push([]Event{{}, {}, {}})
	got, err = bAhead.Since(10)
	if err != nil || got != nil {
		t.Fatalf("ahead: got=%v, err=%v", got, err)
	}
}

func TestBufferEvictExpired(t *testing.T) {
	b := NewBuffer()
	b.ttl = time.Minute         // короткий TTL для теста
	t0 := time.Unix(1000, 0)    // фиксированная точка во времени
	b.Push([]Event{{}, {}, {}}) // события 1, 2, 3
	for i := range b.events {
		b.events[i].insertedAt = t0 // ← ручная установка времени вставки
	}
	count := b.EvictExpired(t0.Add(5 * time.Minute))

	if count != 3 {
		t.Fatalf("count = %d, want 3", count)
	}

	if len(b.events) != 0 {
		t.Fatalf("len = %d, want 0", len(b.events))
	}

	b1 := NewBuffer()
	b1.ttl = time.Minute
	b1.Push([]Event{{}, {}, {}})
	b1.events[0].insertedAt = t0
	b1.events[1].insertedAt = t0
	b1.events[2].insertedAt = t0.Add(5 * time.Minute)

	count = b1.EvictExpired(t0.Add(5 * time.Minute))
	if count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}
	if len(b1.events) != 1 {
		t.Fatalf("len = %d, want 1", len(b1.events))
	}

	got, err := b1.Since(1)
	if !errors.Is(err, ErrCursorTooOld) {
		t.Fatalf("err = %v, want ErrCursorTooOld", err)
	}
	if got != nil {
		t.Fatalf("got = %v, want nil", got)
	}
}
