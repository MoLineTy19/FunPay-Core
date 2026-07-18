package fp

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

type Throttler struct {
	mu        sync.Mutex
	lastTime  time.Time
	minDelay  time.Duration
	maxJitter time.Duration
	rng       *rand.Rand
}

func NewThrottler(minDelay, maxJitter time.Duration) *Throttler {
	return &Throttler{
		minDelay:  minDelay,
		maxJitter: maxJitter,
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (t *Throttler) Wait(ctx context.Context) error {
	t.mu.Lock()

	elapsed := time.Since(t.lastTime)
	delay := t.minDelay - elapsed
	if delay < 0 {
		delay = 0
	}

	if t.maxJitter > 0 {
		delay += time.Duration(t.rng.Int63n(int64(t.maxJitter)))
	}

	t.lastTime = time.Now()

	t.mu.Unlock()

	if delay == 0 {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
		return nil
	}
}
