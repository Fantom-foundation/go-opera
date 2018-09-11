package net

import "github.com/andrecronje/lachesis/poset"

type SyncRequest struct {
	FromID int
	Known  map[int]int
}

type SyncResponse struct {
	FromID    int
	SyncLimit bool
	Events    []poset.WireEvent
	Known     map[int]int
}

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

type EagerSyncRequest struct {
	FromID int
	Events []poset.WireEvent
}

type EagerSyncResponse struct {
	FromID  int
	Success bool
}

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

type FastForwardRequest struct {
	FromID int
}

type FastForwardResponse struct {
	FromID   int
	Block    poset.Block
	Frame    poset.Frame
	Snapshot []byte
}
