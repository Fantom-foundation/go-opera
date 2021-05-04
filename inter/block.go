package inter

import (
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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

func (b *Block) EstimateSize() int {
	return (len(b.Events)+len(b.InternalTxs)+len(b.Txs)+1+1)*32 + len(b.SkippedTxs)*4 + 8 + 8
}

func (b *Block) NotSkippedTxs() []common.Hash {
	txs := append(b.InternalTxs, b.Txs...)

	if len(b.SkippedTxs) == 0 {
		// short circuit if nothing to skip
		return txs
	}
	skipCount := 0
	res := make([]common.Hash, 0, len(txs))
	for i, tx := range txs {
		if skipCount < len(b.SkippedTxs) && b.SkippedTxs[skipCount] == uint32(i) {
			skipCount++
		} else {
			res = append(res, tx)
		}
	}
	return res
}

func FilterSkippedTxs(txs types.Transactions, skippedTxs []uint32) types.Transactions {
	if len(skippedTxs) == 0 {
		// short circuit if nothing to skip
		return txs
	}
	skipCount := 0
	filteredTxs := make(types.Transactions, 0, len(txs))
	for i, tx := range txs {
		if skipCount < len(skippedTxs) && skippedTxs[skipCount] == uint32(i) {
			skipCount++
		} else {
			filteredTxs = append(filteredTxs, tx)
		}
	}

	return filteredTxs
}
