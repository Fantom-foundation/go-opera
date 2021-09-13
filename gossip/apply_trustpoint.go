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

	// export of blocks
	// EVM needs last 256 blocks only, see core/vm.opBlockhash() instruction
	const history idx.Block = 256
	var firstBlockN idx.Block
	if lastBlockN > history {
		firstBlockN = lastBlockN - history
	}
	evm := s.EvmStore()
	for index := firstBlockN; index <= lastBlockN; index++ {
		block := s.GetBlock(index)
		g.SetBlock(index, block)

		for _, e := range block.Events {
			event := s.GetEventPayload(e)
			g.SetEvent(event)
		}

		for _, id := range append(block.InternalTxs, block.Txs...) {
			tx := evm.GetTx(id)
			g.SetTx(tx.Hash(), tx)
		}

		receipts := evm.GetReceipts(index)
		g.SetReceipts(index, receipts)
	}

	// check events
	// TODO: rm
	for i, val := range es.ValidatorStates {
		if val.PrevEpochEvent != hash.ZeroEvent {
			if s.GetEventPayload(val.PrevEpochEvent) == nil {
				log.Crit("PrevEpochEvent not found", "hash", val.PrevEpochEvent, "validator", i)
			}
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
		s.Log.Error("ApplyTrustpoint", "genesis", genesisHash, "trustpoint", g.GenesisHash, "err", err)
	}

	bs, es := g.GetBlockEpochState()
	s.SetBlockEpochState(*bs, *es)

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
