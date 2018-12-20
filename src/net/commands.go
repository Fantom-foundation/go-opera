package net

import "github.com/Fantom-foundation/go-lachesis/src/poset"

// SyncRequest initiates a synchronization request
type SyncRequest struct {
	FromID int64
	Known  map[int64]int64
}

// SyncResponse is a response to a SyncRequest request
type SyncResponse struct {
	FromID    int64
	SyncLimit bool
	Events    []poset.WireEvent
	Known     map[int64]int64
}

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

// EagerSyncRequest after an initial sync to quickly catch up
type EagerSyncRequest struct {
	FromID int64
	Events []poset.WireEvent
}

// EagerSyncResponse response to an EagerSyncRequest
type EagerSyncResponse struct {
	FromID  int64
	Success bool
}

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

// FastForwardRequest request to start a fast forward catch up
type FastForwardRequest struct {
	FromID int64
}

// FastForwardResponse response with the snapshot data for fast forward request
type FastForwardResponse struct {
	FromID   int64
	Block    poset.Block
	Frame    poset.Frame
	Snapshot []byte
}
