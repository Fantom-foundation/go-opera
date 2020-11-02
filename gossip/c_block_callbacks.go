package gossip

import (
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/core"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/Fantom-foundation/lachesis-base/lachesis"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc"
	"github.com/Fantom-foundation/go-opera/gossip/evmstore"
	"github.com/Fantom-foundation/go-opera/inter"
)

func (s *Service) GetConsensusCallbacks() lachesis.ConsensusCallbacks {
	return lachesis.ConsensusCallbacks{
		BeginBlock: func(cBlock *lachesis.Block) lachesis.BlockCallbacks {
			start := time.Now()

			bs := s.store.GetBlockState()
			es := s.store.GetEpochState()

			// Get stateDB
			stateHash := s.store.GetBlock(bs.LastBlock).Root
			statedb := s.store.evm.StateDB(stateHash)

			bs.LastBlock++
			bs.EpochBlocks++

			eventProcessor := s.blockProc.EventsModule.Start(bs, es)

			var atropos inter.EventI
			confirmedEvents := make(hash.OrderedEvents, 0, 3*es.Validators.Len())

			return lachesis.BlockCallbacks{
				ApplyEvent: func(_e dag.Event) {
					e := _e.(inter.EventI)
					if cBlock.Atropos == e.ID() {
						atropos = e
					}
					if !e.NoTxs() {
						// non-empty events only
						confirmedEvents = append(confirmedEvents, e.ID())
						s.occurredTxs.CollectConfirmedTxs(s.store.GetEventPayload(e.ID()).Txs())
					}
					eventProcessor.ProcessConfirmedEvent(e)
				},
				EndBlock: func() (newValidators *pos.Validators) {
					// Note: it's possible that i'th Atropos observes i+1's Atropos,
					// hence ApplyEvent may be not called at all
					if atropos == nil {
						atropos = s.store.GetEvent(cBlock.Atropos)
					}

					blockCtx := blockproc.BlockCtx{
						Idx:    bs.LastBlock,
						Time:   atropos.MedianTime(),
						CBlock: *cBlock,
					}

					bs = eventProcessor.Finalize(blockCtx)

					sealer := s.blockProc.SealerModule.Start(blockCtx, bs, es)
					sealing := sealer.EpochSealing()
					txListener := s.blockProc.TxListenerModule.Start(blockCtx, bs, es, statedb)
					evmProcessor := s.blockProc.EVMModule.Start(blockCtx, statedb, s.GetEvmStateReader(), txListener.OnNewLog)

					// Execute pre-internal transactions
					preInternalTxs := s.blockProc.PreTxTransactor.PopInternalTxs(blockCtx, bs, es, sealing, statedb)
					evmProcessor.Execute(preInternalTxs, true)
					bs = txListener.Finalize()

					// Seal epoch if requested
					if sealing {
						sealer.Update(bs, es)
						bs, es = sealer.SealEpoch()
						newValidators = es.Validators
						s.store.SetEpochState(es)
						txListener.Update(bs, es)
					}
					// At this point, newValidators may be returned and the rest of the code may be executed in a parallel thread

					// Execute post-internal transactions
					internalTxs := s.blockProc.PostTxTransactor.PopInternalTxs(blockCtx, bs, es, sealing, statedb)
					evmProcessor.Execute(internalTxs, true)

					// sort events by Lamport time
					sort.Sort(confirmedEvents)

					// new block
					var block = &inter.Block{
						Time:    blockCtx.Time,
						Atropos: cBlock.Atropos,
						Events:  hash.Events(confirmedEvents),
					}
					for _, tx := range append(preInternalTxs, internalTxs...) {
						block.InternalTxs = append(block.InternalTxs, tx.Hash())
					}

					block, blockEvents := s.spillBlockEvents(block)
					txs := make(types.Transactions, 0, blockEvents.Len()*10)
					for _, e := range blockEvents {
						txs = append(txs, e.Txs()...)
						blockEvents = append(blockEvents, e)
					}

					evmProcessor.Execute(txs, false)
					evmBlock, skippedTxs, receipts := evmProcessor.Finalize()

					block.SkippedTxs = skippedTxs
					block.Root = hash.Hash(evmBlock.Root)

					// memorize event position of each tx
					txPositions := make(map[common.Hash]evmstore.TxPosition)
					for _, e := range blockEvents {
						for i, tx := range e.Txs() {
							// If tx was met in multiple events, then assign to first ordered event
							if _, ok := txPositions[tx.Hash()]; ok {
								continue
							}
							txPositions[tx.Hash()] = evmstore.TxPosition{
								Event:       e.ID(),
								EventOffset: uint32(i),
							}
						}
					}
					// memorize block position of each tx
					for i, tx := range evmBlock.Transactions {
						// not skipped txs only
						position := txPositions[tx.Hash()]
						position.Block = bs.LastBlock
						position.BlockOffset = uint32(i)
						txPositions[tx.Hash()] = position
					}

					// call OnNewReceipt
					for i, r := range receipts {
						txEventPos := txPositions[r.TxHash]
						var creator idx.ValidatorID
						if !txEventPos.Event.IsZero() {
							txEvent := s.store.GetEvent(txEventPos.Event)
							creator = txEvent.Creator()
							if es.Validators.Get(creator) == 0 {
								creator = 0
							}
						}
						txListener.OnNewReceipt(evmBlock.Transactions[i], r, creator)
					}
					bs = txListener.Finalize()
					// At this point, block state is finalized

					s.store.SetBlock(bs.LastBlock, block)
					s.store.SetBlockState(bs)

					// Notify about new block and txs
					s.feed.newBlock.Send(evmcore.ChainHeadNotify{Block: evmBlock})
					s.feed.newTxs.Send(core.NewTxsEvent{Txs: evmBlock.Transactions})
					var logs []*types.Log
					for _, r := range receipts {
						for _, l := range r.Logs {
							logs = append(logs, l)
						}
					}
					s.feed.newLogs.Send(logs)

					// Build index for not skipped txs
					if s.config.TxIndex {
						for _, tx := range evmBlock.Transactions {
							// not skipped txs only
							s.store.evm.SetTxPosition(tx.Hash(), txPositions[tx.Hash()])
						}

						// Index receipts
						if receipts.Len() != 0 {
							s.store.evm.SetReceipts(bs.LastBlock, receipts)

							for _, r := range receipts {
								s.store.evm.IndexLogs(r.Logs...)
							}
						}
					}
					for _, tx := range append(preInternalTxs, internalTxs...) {
						s.store.evm.SetTx(tx.Hash(), tx)
					}

					log.Info("New block", "index", bs.LastBlock, "atropos", block.Atropos, "gas_used",
						evmBlock.GasUsed, "skipped_txs", len(block.SkippedTxs), "txs", len(evmBlock.Transactions), "t", time.Since(start))

					return newValidators
				},
			}
		},
	}
}

// spillBlockEvents excludes first events which exceed BlockGasHardLimit
func (s *Service) spillBlockEvents(block *inter.Block) (*inter.Block, inter.EventPayloads) {
	fullEvents := make(inter.EventPayloads, len(block.Events))
	if len(block.Events) == 0 {
		return block, fullEvents
	}
	gasPowerUsedSum := uint64(0)
	// iterate in reversed order
	for i := len(block.Events) - 1; ; i-- {
		id := block.Events[i]
		e := s.store.GetEventPayload(id)
		if e == nil {
			s.Log.Crit("Block event not found", "event", id.String())
		}
		fullEvents[i] = e
		gasPowerUsedSum += e.GasPowerUsed()
		// stop if limit is exceeded, erase [:i] events
		if gasPowerUsedSum > s.config.Net.Blocks.BlockGasHardLimit {
			// spill
			block.Events = block.Events[i+1:]
			fullEvents = fullEvents[i+1:]
			break
		}
		if i == 0 {
			break
		}
	}
	return block, fullEvents
}
