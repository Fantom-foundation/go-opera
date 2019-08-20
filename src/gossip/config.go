package cfg_gossip

import "github.com/ethereum/go-ethereum/eth/downloader"

type Config struct {
	// Protocol options
	SyncMode downloader.SyncMode

	NoPruning       bool // Whether to disable pruning and flush everything to disk
	NoPrefetch      bool // Whether to disable prefetching and only load state on demand
	ForcedBroadcast bool
}
