package fp

import (
	"testing"

	"github.com/shopspring/decimal"
)

func mkShortcut(id string, st Status) OrderShortcut {
	return OrderShortcut{ID: id, Status: st, Price: decimal.Zero}
}

func TestDiffOrderSnapshotsEmpty(t *testing.T) {
	got := diffOrderSnapshots(map[string]OrderShortcut{}, nil)
	if len(got) != 0 {
		t.Fatalf("empty: got %d events, want 0", len(got))
	}
}

func TestDiffOrderSnapshotsNewOrder(t *testing.T) {
	prev := map[string]OrderShortcut{}
	current := []OrderShortcut{mkShortcut("111", StatusNew)}
	got := diffOrderSnapshots(prev, current)
	if len(got) != 1 {
		t.Fatalf("new: got %d events, want 1", len(got))
	}
	if got[0].Kind != OrderEventNew {
		t.Errorf("kind: got %q, want new", got[0].Kind)
	}
	if got[0].Order.ID != "111" {
		t.Errorf("order id: got %q, want 111", got[0].Order.ID)
	}
	if got[0].ToStatus != StatusNew {
		t.Errorf("toStatus: got %q, want new", got[0].ToStatus)
	}
}

func TestDiffOrderSnapshotsNoChange(t *testing.T) {
	prev := map[string]OrderShortcut{"111": mkShortcut("111", StatusNew)}
	current := []OrderShortcut{mkShortcut("111", StatusNew)}
	got := diffOrderSnapshots(prev, current)
	if len(got) != 0 {
		t.Fatalf("no change: got %d events, want 0", len(got))
	}
}

func TestDiffOrderSnapshotsCompleted(t *testing.T) {
	prev := map[string]OrderShortcut{"111": mkShortcut("111", StatusNew)}
	current := []OrderShortcut{mkShortcut("111", StatusCompleted)}
	got := diffOrderSnapshots(prev, current)
	if len(got) != 1 {
		t.Fatalf("completed: got %d events, want 1", len(got))
	}
	if got[0].Kind != OrderEventCompleted {
		t.Errorf("kind: got %q, want completed", got[0].Kind)
	}
	if got[0].FromStatus != StatusNew {
		t.Errorf("fromStatus: got %q, want new", got[0].FromStatus)
	}
	if got[0].ToStatus != StatusCompleted {
		t.Errorf("toStatus: got %q, want completed", got[0].ToStatus)
	}
}

func TestDiffOrderSnapshotsCancelled(t *testing.T) {
	prev := map[string]OrderShortcut{"111": mkShortcut("111", StatusNew)}
	current := []OrderShortcut{mkShortcut("111", StatusCancelled)}
	got := diffOrderSnapshots(prev, current)
	if len(got) != 1 {
		t.Fatalf("cancelled: got %d events, want 1", len(got))
	}
	if got[0].Kind != OrderEventCancelled {
		t.Errorf("kind: got %q, want cancelled", got[0].Kind)
	}
}

func TestDiffOrderSnapshotsMissingInCurrentIgnored(t *testing.T) {
	prev := map[string]OrderShortcut{"111": mkShortcut("111", StatusNew)}
	current := []OrderShortcut{}
	got := diffOrderSnapshots(prev, current)
	if len(got) != 0 {
		t.Fatalf("missing ignored: got %d events, want 0", len(got))
	}
}
