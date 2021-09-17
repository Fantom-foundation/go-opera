package gossip

import (
	"fmt"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-opera/gossip/evmstore"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/opera/trustpoint"
)

// SaveTrustpoint saves initial state of current epoch.
func (s *Store) SaveTrustpoint(g *trustpoint.Store) (err error) {
	g.GenesisHash = *s.GetGenesisHash()

	_, curr := s.GetBlockEpochState()
	bs, es := s.GetHistoryBlockEpochState(curr.Epoch - 1)
	g.SetBlockEpochState(bs, es)

	s.Log.Info("Creating trustpoint", "epoch", es.Epoch, "block", bs.LastBlock.Idx)

	// export of blocks
	// EVM needs last 256 blocks only, see core/vm.opBlockhash() instruction
	const history idx.Block = 256
	var firstBlockIdx idx.Block
	if bs.LastBlock.Idx > history {
		firstBlockIdx = bs.LastBlock.Idx - history
	}
	evm := s.EvmStore()
	for index := firstBlockIdx; index <= bs.LastBlock.Idx; index++ {
		block := s.GetBlock(index)
		g.SetBlock(index, block)

		for _, e := range block.Events {
			event := s.GetEventPayload(e)
			if event == nil {
				log.Crit("block event not found", "block", index, "event", e)
			}
			g.SetEvent(event)
		}

		for _, id := range append(block.InternalTxs, block.Txs...) {
			tx := evm.GetTx(id)
			g.SetTx(tx.Hash(), tx)
		}

		receipts := evm.GetReceipts(index)
		g.SetReceipts(index, receipts)
	}

	// prev epoch events
	for _, val := range es.ValidatorStates {
		if val.PrevEpochEvent != hash.ZeroEvent {
			e := s.GetEventPayload(val.PrevEpochEvent)
			if e == nil {
				log.Crit("validator prev epoch event not found", "event", val.PrevEpochEvent)
			}
			g.SetEvent(e)
		}
	}
	// last events
	for _, val := range bs.ValidatorStates {
		if val.LastEvent != hash.ZeroEvent {
			e := s.GetEventPayload(val.LastEvent)
			if e == nil {
				log.Crit("validator last event not found", "event", val.LastEvent)
			}
			g.SetEvent(e)
		}
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
	const txIndex = true // TODO: parametrize it
	evm := s.EvmStore()

	genesisHash := *s.GetGenesisHash()
	if genesisHash != g.GenesisHash {
		err = fmt.Errorf("Genesis mismatch")
		s.Log.Error("Applying trustpoint", "genesis", genesisHash, "trustpoint", g.GenesisHash, "err", err)
	}

	bs, es := g.GetBlockEpochState()
	s.Log.Info("Applying trustpoint", "genesis", genesisHash, "epoch", es.Epoch)
	s.SetBlockEpochState(*bs, *es)
	s.SetHistoryBlockEpochState(es.Epoch, *bs, *es)
	s.SetEpochBlock(bs.LastBlock.Idx, es.Epoch)

	// prev epoch events
	for _, val := range es.ValidatorStates {
		if val.PrevEpochEvent != hash.ZeroEvent {
			e := g.GetEvent(val.PrevEpochEvent)
			if e == nil {
				log.Crit("validator prev epoch event not found", "event", val.PrevEpochEvent)
			}
			s.SetEvent(e)
		}
	}
	// last events
	for _, val := range bs.ValidatorStates {
		if val.LastEvent != hash.ZeroEvent {
			e := g.GetEvent(val.LastEvent)
			if e == nil {
				log.Crit("validator last event not found", "event", val.LastEvent)
			}
			s.SetEvent(e)
		}
	}

	g.ForEachBlock(func(index idx.Block, block *inter.Block) {
		s.SetBlock(index, block)
		s.SetBlockIndex(block.Atropos, index)

		txcap := len(block.Events) * 10
		txs := make(types.Transactions, 0, txcap)
		txPositions := make(map[common.Hash]evmstore.TxPosition, txcap)

		for _, h := range append(block.InternalTxs, block.Txs...) {
			tx := g.GetTx(h)
			txs = append(txs, tx)
			evm.SetTx(h, tx)
		}

		for _, id := range block.Events {
			e := g.GetEvent(id)
			if e == nil {
				log.Crit("block event not found", "block", index, "event", id)
			}
			s.SetEvent(e)
			for i, tx := range e.Txs() {
				h := tx.Hash()
				txs = append(txs, tx)
				// If tx was met in multiple events, then assign to first ordered event
				if _, ok := txPositions[h]; ok {
					continue
				}
				txPositions[h] = evmstore.TxPosition{
					Event:       id,
					EventOffset: uint32(i),
				}
			}
		}

		if txIndex {
			txs = inter.FilterSkippedTxs(txs, block.SkippedTxs)
			for i, tx := range txs {
				h := tx.Hash()
				pos := txPositions[h]
				pos.Block = index
				pos.BlockOffset = uint32(i)
				evm.SetTxPosition(h, pos)
			}

			receipts := g.GetRawReceipts(index)
			if len(receipts) != len(txs) {
				log.Crit("Receipts mismatch", "block", index, "txs", len(txs), "receipts", receipts)
			}
			evm.SetRawReceipts(index, receipts)
			for _, r := range receipts {
				evm.IndexLogs(r.Logs...)
			}
		}
	})

	// apply EVM genesis
	log.Info("Importing EVM state", "root", bs.FinalizedStateRoot)
	item := g.GetRawEvmItem()
	err = evm.ApplyRawEvmItems(item)
	if err != nil {
		return err
	}

	return
}
