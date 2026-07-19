package engine

import (
	"errors"
	"sync"
	"time"
)

type bufEntry struct {
	event      Event
	insertedAt time.Time
}
type Buffer struct {
	mu     sync.Mutex
	events []bufEntry
	next   int64
	ttl    time.Duration
}

func NewBuffer() *Buffer {
	return &Buffer{
		events: make([]bufEntry, 0, 128),
		next:   1,
		ttl:    10 * time.Minute,
	}
}

func (b *Buffer) Push(in []Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	for _, e := range in {
		e.EventID = b.next
		b.next++
		b.events = append(b.events, bufEntry{event: e, insertedAt: now})
	}
}

var ErrCursorTooOld = errors.New("cursor too old: events evicted from buffer")

func (b *Buffer) Since(last int64) ([]Event, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.events) == 0 {
		return nil, nil
	}

	if last+1 < b.events[0].event.EventID {
		return nil, ErrCursorTooOld
	}

	if last >= b.events[len(b.events)-1].event.EventID {
		return nil, nil
	}

	var startPos int
	for i := 0; i < len(b.events); i++ {
		if b.events[i].event.EventID > last {
			startPos = i
			break
		}
	}

	out := make([]Event, 0, len(b.events)-startPos)
	for _, entry := range b.events[startPos:] {
		out = append(out, entry.event)
	}

	return out, nil
}

func (b *Buffer) EvictExpired(now time.Time) int {
	b.mu.Lock()
	defer b.mu.Unlock()

	var count int

	for _, e := range b.events {
		if now.Sub(e.insertedAt) > b.ttl {
			count++
			continue
		}
		break
	}

	if count > 0 {
		b.events = b.events[count:]
	}

	return count
}
