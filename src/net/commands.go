package net

import "github.com/andrecronje/lachesis/src/poset"

type SyncRequest struct {
	FromID int64
	Known  map[int64]int64
}

type SyncResponse struct {
	FromID    int64
	SyncLimit bool
	Events    []poset.WireEvent
	Known     map[int64]int64
}

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

type EagerSyncRequest struct {
	FromID int64
	Events []poset.WireEvent
}

type EagerSyncResponse struct {
	FromID  int64
	Success bool
}

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

type FastForwardRequest struct {
	FromID int64
}

type FastForwardResponse struct {
	FromID   int64
	Block    poset.Block
	Frame    poset.Frame
	Snapshot []byte
}
