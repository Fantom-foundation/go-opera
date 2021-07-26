package gossip

import (
	"fmt"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/lachesis"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-opera/gossip/blockproc"
	"github.com/Fantom-foundation/go-opera/gossip/evmstore"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/drivertype"
	"github.com/Fantom-foundation/go-opera/opera/genesis"
	"github.com/Fantom-foundation/go-opera/opera/trustpoint"
)

func (s *Store) firstEpochBlock() (b *inter.Block, n idx.Block, err error) {
	bs, es := s.GetBlockEpochState()

	for n = bs.LastBlock.Idx; n > 0; n-- {
		b = s.GetBlock(n)
		if b == nil || (b.Time == es.EpochStart) {
			break
		}
	}
	if b == nil {
		err = fmt.Errorf("1st block of epoch %d is not found", es.Epoch)
		return
	}

	return
}

// SaveTrustpoint saves initial state of current epoch.
func (s *Store) SaveTrustpoint(g *trustpoint.Store) (err error) {
	g.GenesisHash = *s.GetGenesisHash()

	bs, es := s.GetBlockEpochState()
	bs, es = bs.Copy(), es.Copy()

	lastBlock, lastBlockN, err := s.firstEpochBlock()
	if err != nil {
		return err
	}

	s.Log.Info("Save trustpoint", "block", lastBlockN)

	// reset to epoch start
	bs.FinalizedStateRoot = lastBlock.Root
	bs.LastBlock = blockproc.BlockCtx{
		Idx:     lastBlockN,
		Time:    lastBlock.Time,
		Atropos: lastBlock.Atropos,
	}

	bs.EpochGas = 0
	bs.EpochCheaters = lachesis.Cheaters{}
	bs.NextValidatorProfiles = make(map[idx.ValidatorID]drivertype.Validator)
	for i, vs := range bs.ValidatorStates {
		vs.DirtyGasRefund = 0
		vs.Uptime = 0
		bs.ValidatorStates[i] = vs
	}
	g.SetBlockEpochState(&bs, &es)

	// events
	for _, val := range es.ValidatorStates {
		if val.PrevEpochEvent != hash.ZeroEvent {
			e := s.GetEventPayload(val.PrevEpochEvent)
			g.SetEvent(e)
		}
	}

	// EVM needs last 256 blocks only, see core/vm.opBlockhash() instruction
	const history idx.Block = 256
	var firstBlockN idx.Block
	if lastBlockN > history {
		firstBlockN = lastBlockN - history
	}
	// export of blocks
	evm := s.EvmStore()
	for index := firstBlockN; index <= lastBlockN; index++ {
		block := s.GetBlock(index)
		internalTxs := make([]*types.Transaction, len(block.InternalTxs))
		for i, txid := range block.InternalTxs {
			internalTxs[i] = evm.GetTx(txid)
		}
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
			InternalTxs: internalTxs,
			Txs:         txs,
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
	genesisHash := *s.GetGenesisHash()
	if genesisHash != g.GenesisHash {
		err = fmt.Errorf("Genesis mismatch")
		s.Log.Error("ApplyTrustpoint", "genesis", genesisHash, "trustpoint", g.GenesisHash, "err", err)
	}

	bs, es := g.GetBlockEpochState()
	s.SetBlockEpochState(*bs, *es)

	g.ForEachEvent(func(e *inter.EventPayload) bool {
		s.SetEvent(e)
		return true
	})

	evm := s.EvmStore()
	g.ForEachBlock(func(blockIdx idx.Block, block genesis.Block) {
		internalTxHashes := make([]common.Hash, len(block.InternalTxs))
		for i, tx := range block.InternalTxs {
			internalTxHashes[i] = tx.Hash()
		}
		txHashes := make([]common.Hash, len(block.Txs))
		for i, tx := range block.Txs {
			txHashes[i] = tx.Hash()
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
