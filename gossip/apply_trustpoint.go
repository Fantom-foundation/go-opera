package gossip

import (
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-opera/gossip/evmstore"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/opera/genesis"
	"github.com/Fantom-foundation/go-opera/opera/trustpoint"
)

// SaveTrustpoint saves initial state.
func (s *Store) SaveTrustpoint(g *trustpoint.Store) (err error) {
	g.GenesisHash = *s.GetGenesisHash()

	bs, es := s.GetBlockEpochState()
	g.SetBlockEpochState(&bs, &es)

	// EVM needs last 256 blocks only, see core/vm.opBlockhash() instruction
	const history idx.Block = 256
	var firstBlockIdx idx.Block
	if bs.LastBlock.Idx > history {
		firstBlockIdx = bs.LastBlock.Idx - history
	}
	// export of blocks
	evm := s.EvmStore()
	for index := firstBlockIdx; index <= bs.LastBlock.Idx; index++ {
		block := s.GetBlock(index)
		txs := make([]*types.Transaction, len(block.Txs))
		for i, txid := range block.Txs {
			txs[i] = evm.GetTx(txid)
		}
		receipts := evm.GetReceipts(index)
		receiptsForStorage := make([]*types.ReceiptForStorage, len(receipts))
		for i, r := range receipts {
			receiptsForStorage[i] = (*types.ReceiptForStorage)(r)
		}
		g.SetBlock(index, genesis.Block{
			Time:        block.Time,
			Atropos:     block.Atropos,
			Txs:         txs,
			InternalTxs: types.Transactions{},
			Root:        block.Root,
			Receipts:    receiptsForStorage,
		})
	}

	// export of EVM state
	log.Info("Exporting EVM state", "root", bs.FinalizedStateRoot)
	it := evm.EvmDb.NewIterator(nil, nil)
	for it.Next() {
		// TODO: skip unnecessary states after root
		g.SetRawEvmItem(it.Key(), it.Value())
	}
	it.Release()

	return nil
}

// ApplyTrustpoint applies initial state.
func (s *Store) ApplyTrustpoint(g *trustpoint.Store) (err error) {
	s.SetGenesisHash(g.GenesisHash)

	bs, es := g.GetBlockEpochState()
	s.SetBlockEpochState(*bs, *es)

	evm := s.EvmStore()
	g.ForEachBlock(func(blockIdx idx.Block, block genesis.Block) {
		txHashes := make([]common.Hash, len(block.Txs))
		internalTxHashes := make([]common.Hash, len(block.InternalTxs))
		for i, tx := range block.Txs {
			txHashes[i] = tx.Hash()
		}
		for i, tx := range block.InternalTxs {
			internalTxHashes[i] = tx.Hash()
		}
		for i, tx := range append(block.InternalTxs, block.Txs...) {
			evm.SetTxPosition(tx.Hash(), evmstore.TxPosition{
				Block:       blockIdx,
				BlockOffset: uint32(i),
			})
			evm.SetTx(tx.Hash(), tx)
		}
		gasUsed := uint64(0)
		if len(block.Receipts) != 0 {
			gasUsed = block.Receipts[len(block.Receipts)-1].CumulativeGasUsed
			evm.SetRawReceipts(blockIdx, block.Receipts)
			allTxs := block.Txs
			if block.InternalTxs.Len() > 0 {
				allTxs = append(block.InternalTxs, block.Txs...)
			}
			logIdx := uint(0)
			for i, r := range block.Receipts {
				for _, l := range r.Logs {
					l.BlockNumber = uint64(blockIdx)
					l.TxHash = allTxs[i].Hash()
					l.Index = logIdx
					l.TxIndex = uint(i)
					l.BlockHash = common.Hash(block.Atropos)
					logIdx++
				}
				evm.IndexLogs(r.Logs...)
			}
		}

		s.SetBlock(blockIdx, &inter.Block{
			Time:        block.Time,
			Atropos:     block.Atropos,
			Events:      hash.Events{},
			Txs:         txHashes,
			InternalTxs: internalTxHashes,
			SkippedTxs:  []uint32{},
			GasUsed:     gasUsed,
			Root:        block.Root,
		})
		s.SetBlockIndex(block.Atropos, blockIdx)
	})

	// apply EVM genesis
	log.Info("Importing EVM state", "root", bs.FinalizedStateRoot)
	err = evm.ApplyRawEvmItems(g.GetRawEvmItem())
	if err != nil {
		return err
	}

	return
}
