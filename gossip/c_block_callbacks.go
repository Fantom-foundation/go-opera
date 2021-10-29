package gossip

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
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
	"github.com/Fantom-foundation/go-opera/gossip/sfcapi"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/utils"
)

type ExtendedTxPosition struct {
	evmstore.TxPosition
	EventCreator idx.ValidatorID
}

// GetConsensusCallbacks returns single (for Service) callback instance.
func (s *Service) GetConsensusCallbacks() lachesis.ConsensusCallbacks {
	return lachesis.ConsensusCallbacks{
		BeginBlock: consensusCallbackBeginBlockFn(
			s.blockProcTasks,
			&s.blockProcWg,
			&s.blockBusyFlag,
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
	blockBusyFlag *uint32,
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

		// Note: take copies to avoid race conditions with API calls
		bs := store.GetBlockState().Copy()
		es := store.GetEpochState().Copy()

		// merge cheaters to ensure that every cheater will get punished even if only previous (not current) Atropos observed a doublesign
		// this feature is needed because blocks may be skipped even if cheaters list isn't empty
		// otherwise cheaters would get punished after a first block where cheaters were observed
		bs.EpochCheaters = mergeCheaters(bs.EpochCheaters, cBlock.Cheaters)

		// Get stateDB
		statedb, err := store.evm.StateDB(bs.FinalizedStateRoot)
		if err != nil {
			log.Crit("Failed to open StateDB", "err", err)
		}

		eventProcessor := blockProc.EventsModule.Start(bs, es)

		atroposTime := bs.LastBlock.Time + 1
		atroposDegenerate := true
		confirmedEvents := make(hash.OrderedEvents, 0, 3*es.Validators.Len())

		return lachesis.BlockCallbacks{
			ApplyEvent: func(_e dag.Event) {
				e := _e.(inter.EventI)
				if cBlock.Atropos == e.ID() {
					atroposTime = e.MedianTime()
					atroposDegenerate = false
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
				if atroposTime <= bs.LastBlock.Time {
					atroposTime = bs.LastBlock.Time + 1
				}
				blockCtx := blockproc.BlockCtx{
					Idx:     bs.LastBlock.Idx + 1,
					Time:    atroposTime,
					Atropos: cBlock.Atropos,
				}
				// Note:
				// it's possible that a previous Atropos observes current Atropos (1)
				// (even stronger statement is true - it's possible that current Atropos is equal to a previous Atropos).
				// (1) is true when and only when ApplyEvent wasn't called.
				// In other words, we should assume that every non-cheater root may be elected as an Atropos in any order,
				// even if typically every previous Atropos happened-before current Atropos
				// We have to skip block in case (1) to ensure that every block ID is unique.
				// If Atropos ID wasn't used as a block ID, it wouldn't be required.
				skipBlock := atroposDegenerate
				// Check if empty block should be pruned
				emptyBlock := confirmedEvents.Len() == 0 && cBlock.Cheaters.Len() == 0
				skipBlock = skipBlock || (emptyBlock && blockCtx.Time < bs.LastBlock.Time+es.Rules.Blocks.MaxEmptyBlockSkipPeriod)
				// Finalize the progress of eventProcessor
				bs = eventProcessor.Finalize(blockCtx, skipBlock) // TODO: refactor to not mutate the bs, it is unclear
				if skipBlock {
					// save the latest block state even if block is skipped
					store.SetBlockEpochState(bs, es)
					log.Debug("Frame is skipped", "atropos", cBlock.Atropos.String())
					return nil
				}

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
					sfcapi.OnNewLog(store.sfcapi, l)
				}
				evmProcessor := blockProc.EVMModule.Start(blockCtx, statedb, evmStateReader, onNewLogAll, es.Rules)

				// Execute pre-internal transactions
				preInternalTxs := blockProc.PreTxTransactor.PopInternalTxs(blockCtx, bs, es, sealing, statedb)
				preInternalReceipts := evmProcessor.Execute(preInternalTxs, true)
				bs = txListener.Finalize()
				for _, r := range preInternalReceipts {
					if r.Status == 0 {
						log.Warn("Pre-internal transaction reverted", "txid", r.TxHash.String())
					}
				}

				// Seal epoch if requested
				if sealing {
					sealer.Update(bs, es)
					bs, es = sealer.SealEpoch() // TODO: refactor to not mutate the bs, it is unclear
					store.SetBlockEpochState(bs, es)
					newValidators = es.Validators
					txListener.Update(bs, es)
				}

				// At this point, newValidators may be returned and the rest of the code may be executed in a parallel thread
				blockFn := func() {
					// Execute post-internal transactions
					internalTxs := blockProc.PostTxTransactor.PopInternalTxs(blockCtx, bs, es, sealing, statedb)
					internalReceipts := evmProcessor.Execute(internalTxs, true)
					for _, r := range internalReceipts {
						if r.Status == 0 {
							log.Warn("Internal transaction reverted", "txid", r.TxHash.String())
						}
					}

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
					}

					externalReceipts := evmProcessor.Execute(txs, false)

					evmBlock, skippedTxs, allReceipts := evmProcessor.Finalize()
					block.SkippedTxs = skippedTxs
					block.Root = hash.Hash(evmBlock.Root)
					block.GasUsed = evmBlock.GasUsed

					// memorize event position of each tx
					txPositions := make(map[common.Hash]ExtendedTxPosition)
					for _, e := range blockEvents {
						for i, tx := range e.Txs() {
							// If tx was met in multiple events, then assign to first ordered event
							if _, ok := txPositions[tx.Hash()]; ok {
								continue
							}
							txPositions[tx.Hash()] = ExtendedTxPosition{
								TxPosition: evmstore.TxPosition{
									Event:       e.ID(),
									EventOffset: uint32(i),
								},
								EventCreator: e.Creator(),
							}
						}
					}
					// memorize block position of each tx
					for i, tx := range evmBlock.Transactions {
						// not skipped txs only
						position := txPositions[tx.Hash()]
						position.Block = blockCtx.Idx
						position.BlockOffset = uint32(i)
						txPositions[tx.Hash()] = position
					}

					// call OnNewReceipt
					for i, r := range allReceipts {
						creator := txPositions[r.TxHash].EventCreator
						if creator != 0 && es.Validators.Get(creator) == 0 {
							creator = 0
						}
						txListener.OnNewReceipt(evmBlock.Transactions[i], r, creator)
					}
					bs = txListener.Finalize() // TODO: refactor to not mutate the bs
					bs.FinalizedStateRoot = block.Root
					// At this point, block state is finalized

					// Build index for not skipped txs
					if txIndex {
						for _, tx := range evmBlock.Transactions {
							// not skipped txs only
							store.evm.SetTxPosition(tx.Hash(), txPositions[tx.Hash()].TxPosition)
						}

						// Index receipts
						if allReceipts.Len() != 0 {
							store.evm.SetReceipts(blockCtx.Idx, allReceipts)
							for _, r := range allReceipts {
								store.evm.IndexLogs(r.Logs...)
							}
						}
					}
					for _, tx := range append(preInternalTxs, internalTxs...) {
						store.evm.SetTx(tx.Hash(), tx)
					}

					bs.LastBlock = blockCtx
					store.SetHistoryBlockEpochState(es.Epoch, bs, es)
					store.SetEpochBlock(blockCtx.Idx+1, es.Epoch)
					store.SetBlock(blockCtx.Idx, block)
					store.SetBlockIndex(block.Atropos, blockCtx.Idx)
					store.SetBlockEpochState(bs, es)
					store.EvmStore().SetCachedEvmBlock(blockCtx.Idx, evmBlock)

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

					store.commitEVM()

					now := time.Now()
					log.Info("New block", "index", blockCtx.Idx, "id", block.Atropos, "gas_used",
						evmBlock.GasUsed, "txs", fmt.Sprintf("%d/%d", len(evmBlock.Transactions), len(block.SkippedTxs)),
						"age", utils.PrettyDuration(now.Sub(block.Time.Time())), "t", utils.PrettyDuration(now.Sub(start)))
				}
				if confirmedEvents.Len() != 0 {
					atomic.StoreUint32(blockBusyFlag, 1)
					wg.Add(1)
					err := parallelTasks.Enqueue(func() {
						defer atomic.StoreUint32(blockBusyFlag, 0)
						defer wg.Done()
						blockFn()
					})
					if err != nil {
						panic(err)
					}
				} else {
					blockFn()
				}

				return newValidators
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

func mergeCheaters(a, b lachesis.Cheaters) lachesis.Cheaters {
	if len(b) == 0 {
		return a
	}
	if len(a) == 0 {
		return b
	}
	aSet := a.Set()
	merged := make(lachesis.Cheaters, 0, len(b)+len(a))
	for _, v := range a {
		merged = append(merged, v)
	}
	for _, v := range b {
		if _, ok := aSet[v]; !ok {
			merged = append(merged, v)
		}
	}
	return merged
}
