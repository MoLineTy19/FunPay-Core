package fp

import (
	"testing"
	"time"
)

func TestClientUpdateAuth(t *testing.T) {
	c := NewClient("old-key", "old-session", "old-seal", "old-csrf", 1*time.Millisecond, 1*time.Millisecond)

	gk, sid, seal := c.SnapshotAuth()
	if gk != "old-key" || sid != "old-session" || seal != "old-seal" {
		t.Fatalf("snapshot before: got key=%q session=%q seal=%q", gk, sid, seal)
	}

	prev := c.UpdateAuth("new-key", "new-session", "new-seal")
	if prev != "old-seal" {
		t.Errorf("previous seal: got %q, want old-seal", prev)
	}

	gk, sid, seal = c.SnapshotAuth()
	if gk != "new-key" {
		t.Errorf("golden key after: got %q, want new-key", gk)
	}
	if sid != "new-session" {
		t.Errorf("session after: got %q, want new-session", sid)
	}
	if seal != "new-seal" {
		t.Errorf("seal after: got %q, want new-seal", seal)
	}
}
