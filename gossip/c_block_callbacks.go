package gossip

import (
	"sort"
	"sync"
	"time"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/Fantom-foundation/lachesis-base/lachesis"
	"github.com/Fantom-foundation/lachesis-base/utils/workers"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc/verwatcher"
	"github.com/Fantom-foundation/go-opera/gossip/emitter"
	"github.com/Fantom-foundation/go-opera/gossip/evmstore"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/opera"
)

// GetConsensusCallbacks returns single (for Service) callback instance.
func (s *Service) GetConsensusCallbacks() lachesis.ConsensusCallbacks {
	return lachesis.ConsensusCallbacks{
		BeginBlock: consensusCallbackBeginBlockFn(
			s.blockProcTasks,
			&s.blockProcWg,
			s.store,
			s.blockProcModules,
			s.config.TxIndex,
			&s.feed,
			s.emitter,
			s.verWatcher,
			nil,
		),
	}
}

// consensusCallbackBeginBlockFn takes only necessaries for block processing and
// makes lachesis.BeginBlockFn.
// Note that onBlockEnd would be run async.
func consensusCallbackBeginBlockFn(
	parallelTasks *workers.Workers,
	wg *sync.WaitGroup,
	store *Store,
	blockProc BlockProc,
	txIndex bool,
	feed *ServiceFeed,
	emitter *emitter.Emitter,
	verWatcher *verwatcher.VerWarcher,
	onBlockEnd func(block *inter.Block, preInternalReceipts, internalReceipts, externalReceipts types.Receipts),
) lachesis.BeginBlockFn {
	return func(cBlock *lachesis.Block) lachesis.BlockCallbacks {
		wg.Wait()
		start := time.Now()

		bs := store.GetBlockState()
		es := store.GetEpochState()

		// Get stateDB
		statedb, err := store.evm.StateDB(bs.LastStateRoot)
		if err != nil {
			log.Crit("Failed to open StateDB", "err", err)
		}

		bs.LastBlock++

		eventProcessor := blockProc.EventsModule.Start(bs, es)

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
				}
				eventProcessor.ProcessConfirmedEvent(e)
				if emitter != nil {
					emitter.OnEventConfirmed(e)
				}
			},
			EndBlock: func() (newValidators *pos.Validators) {
				// Note: it's possible that i'th Atropos observes i+1's Atropos,
				// hence ApplyEvent may be not called at all
				if atropos == nil {
					atropos = store.GetEvent(cBlock.Atropos)
				}

				blockCtx := blockproc.BlockCtx{
					Idx:    bs.LastBlock,
					Time:   atropos.MedianTime(),
					CBlock: *cBlock,
				}

				bs = eventProcessor.Finalize(blockCtx) // TODO: refactor to don't mutate the bs, it is unclear

				sealer := blockProc.SealerModule.Start(blockCtx, bs, es)
				sealing := sealer.EpochSealing()
				txListener := blockProc.TxListenerModule.Start(blockCtx, bs, es, statedb)
				evmStateReader := &EvmStateReader{
					ServiceFeed: feed,
					store:       store,
				}
				onNewLogAll := func(l *types.Log) {
					txListener.OnNewLog(l)
					if verWatcher != nil {
						verWatcher.OnNewLog(l)
					}
				}
				evmProcessor := blockProc.EVMModule.Start(blockCtx, statedb, evmStateReader, onNewLogAll, es.Rules)

				// Execute pre-internal transactions
				preInternalTxs := blockProc.PreTxTransactor.PopInternalTxs(blockCtx, bs, es, sealing, statedb)
				preInternalReceipts := evmProcessor.Execute(preInternalTxs, true)
				bs = txListener.Finalize()

				// Seal epoch if requested
				if sealing {
					sealer.Update(bs, es)
					bs, es = sealer.SealEpoch() // TODO: refactor to don't mutate the bs, it is unclear
					newValidators = es.Validators
					store.SetEpochState(es)
					txListener.Update(bs, es)
				}

				// At this point, newValidators may be returned and the rest of the code may be executed in a parallel thread
				wg.Add(1)
				err := parallelTasks.Enqueue(func() {
					defer wg.Done()

					// Execute post-internal transactions
					internalTxs := blockProc.PostTxTransactor.PopInternalTxs(blockCtx, bs, es, sealing, statedb)
					internalReceipts := evmProcessor.Execute(internalTxs, true)

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

					block, blockEvents := spillBlockEvents(store, block, es.Rules)
					txs := make(types.Transactions, 0, blockEvents.Len()*10)
					for _, e := range blockEvents {
						txs = append(txs, e.Txs()...)
						blockEvents = append(blockEvents, e)
					}

					externalReceipts := evmProcessor.Execute(txs, false)
					evmBlock, skippedTxs, allReceipts := evmProcessor.Finalize()

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
					for i, r := range allReceipts {
						txEventPos := txPositions[r.TxHash]
						var creator idx.ValidatorID
						if !txEventPos.Event.IsZero() {
							txEvent := store.GetEvent(txEventPos.Event)
							creator = txEvent.Creator()
							if es.Validators.Get(creator) == 0 {
								creator = 0
							}
						}
						txListener.OnNewReceipt(evmBlock.Transactions[i], r, creator)
					}
					bs = txListener.Finalize() // TODO: refactor to don't mutate the bs
					bs.LastStateRoot = block.Root
					// At this point, block state is finalized

					// Build index for not skipped txs
					if txIndex {
						for _, tx := range evmBlock.Transactions {
							// not skipped txs only
							store.evm.SetTxPosition(tx.Hash(), txPositions[tx.Hash()])
						}

						// Index receipts
						if allReceipts.Len() != 0 {
							store.evm.SetReceipts(bs.LastBlock, allReceipts)

							for _, r := range allReceipts {
								store.evm.IndexLogs(r.Logs...)
							}
						}
					}
					for _, tx := range append(preInternalTxs, internalTxs...) {
						store.evm.SetTx(tx.Hash(), tx)
					}

					store.SetBlock(bs.LastBlock, block)
					store.SetBlockState(bs)

					// Notify about new block and txs
					if feed != nil {
						feed.newBlock.Send(evmcore.ChainHeadNotify{Block: evmBlock})
						feed.newTxs.Send(core.NewTxsEvent{Txs: evmBlock.Transactions})
						var logs []*types.Log
						for _, r := range allReceipts {
							for _, l := range r.Logs {
								logs = append(logs, l)
							}
						}
						feed.newLogs.Send(logs)
					}

					if onBlockEnd != nil {
						onBlockEnd(block, preInternalReceipts, internalReceipts, externalReceipts)
					}

					store.capEVM()

					log.Info("New block", "index", bs.LastBlock, "atropos", block.Atropos, "gas_used",
						evmBlock.GasUsed, "skipped_txs", len(block.SkippedTxs), "txs", len(evmBlock.Transactions), "t", time.Since(start))
				})
				if err != nil {
					panic(err)
				}

				return
			},
		}
	}
}

// spillBlockEvents excludes first events which exceed MaxBlockGas
func spillBlockEvents(store *Store, block *inter.Block, network opera.Rules) (*inter.Block, inter.EventPayloads) {
	fullEvents := make(inter.EventPayloads, len(block.Events))
	if len(block.Events) == 0 {
		return block, fullEvents
	}
	gasPowerUsedSum := uint64(0)
	// iterate in reversed order
	for i := len(block.Events) - 1; ; i-- {
		id := block.Events[i]
		e := store.GetEventPayload(id)
		if e == nil {
			log.Crit("Block event not found", "event", id.String())
		}
		fullEvents[i] = e
		gasPowerUsedSum += e.GasPowerUsed()
		// stop if limit is exceeded, erase [:i] events
		if gasPowerUsedSum > network.Blocks.MaxBlockGas {
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
