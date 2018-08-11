package net

import "github.com/andrecronje/lachesis/hashgraph"

type SyncRequest struct {
	FromID int
	Known  map[int]int
}

type SyncResponse struct {
	FromID    int
	SyncLimit bool
	Events    []hashgraph.WireEvent
	Known     map[int]int
}

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

type EagerSyncRequest struct {
	FromID int
	Events []hashgraph.WireEvent
}

type EagerSyncResponse struct {
	FromID  int
	Success bool
}

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

type SubmitTxRequest struct {
	Transactions []string
	ID int
}

type SubmitTxResponse struct {
	FromID  int
	Success bool
}
