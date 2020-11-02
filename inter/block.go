package inter

import (
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/ethereum/go-ethereum/common"
)

type Block struct {
	Time        Timestamp
	Atropos     hash.Event
	Events      hash.Events
	Txs         []common.Hash
	InternalTxs []common.Hash
	SkippedTxs  []uint32 // indexes of skipped txs, starting from first tx of first event, ending with last tx of last event
	GasUsed     uint64
	Root        hash.Hash
}
