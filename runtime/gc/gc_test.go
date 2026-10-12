package gc

import "testing"

func TestTrackAllocAccumulatesTotals(t *testing.T) {
	m := New()
	m.TrackAlloc(128)
	m.TrackAlloc(256)

	stats := m.GetStats()
	if stats.AllocatedObjects != 2 {
		t.Fatalf("AllocatedObjects = %d, want 2", stats.AllocatedObjects)
	}
	if stats.AllocatedBytes != 384 {
		t.Fatalf("AllocatedBytes = %d, want 384", stats.AllocatedBytes)
	}
	if stats.LiveObjects != 2 {
		t.Fatalf("LiveObjects = %d, want 2", stats.LiveObjects)
	}
}

func TestTrackFreeDecrementsLiveObjects(t *testing.T) {
	m := New()
	m.TrackAlloc(64)
	m.TrackAlloc(64)
	m.TrackFree()

	if got := m.GetStats().LiveObjects; got != 1 {
		t.Fatalf("LiveObjects = %d, want 1", got)
	}
	// Bytes are a cumulative total and must not be reduced by a free.
	if got := m.GetStats().AllocatedBytes; got != 128 {
		t.Fatalf("AllocatedBytes = %d, want 128", got)
	}
}

func TestCollectIncrementsCycleCount(t *testing.T) {
	m := New()
	m.Collect()
	m.Collect()
	if got := m.GetStats().GCCycles; got != 2 {
		t.Fatalf("GCCycles = %d, want 2", got)
	}
}

func TestNewStartsWithZeroedStats(t *testing.T) {
	stats := New().GetStats()
	if stats != (Stats{}) {
		t.Fatalf("new manager stats = %+v, want zero value", stats)
	}
}
