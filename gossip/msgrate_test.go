package gossip

import "testing"

func TestCapacityOverflow(t *testing.T) {
	tracker := NewTracker(nil)
	tracker.Update(EventsMsg, 100000*10000000)
	cap := tracker.Capacity(EventsMsg)
	if int32(cap) < 0 {
		t.Fatalf("Negative: %v", int32(cap))
	}
}
