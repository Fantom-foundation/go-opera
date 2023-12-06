package gossip

import (
	"fmt"
	"github.com/Fantom-foundation/go-opera/integration/xenblocks/reporter"
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
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc/verwatcher"
	"github.com/Fantom-foundation/go-opera/gossip/emitter"
	"github.com/Fantom-foundation/go-opera/gossip/evmstore"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/iblockproc"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/utils"
)

var (
	// Ethereum compatible metrics set (see go-ethereum/core)

	headBlockGauge     = metrics.GetOrRegisterGauge("chain/head/block", nil)
	headHeaderGauge    = metrics.GetOrRegisterGauge("chain/head/header", nil)
	headFastBlockGauge = metrics.GetOrRegisterGauge("chain/head/receipt", nil)

	accountReadTimer   = metrics.GetOrRegisterTimer("chain/account/reads", nil)
	accountHashTimer   = metrics.GetOrRegisterTimer("chain/account/hashes", nil)
	accountUpdateTimer = metrics.GetOrRegisterTimer("chain/account/updates", nil)
	accountCommitTimer = metrics.GetOrRegisterTimer("chain/account/commits", nil)

	storageReadTimer   = metrics.GetOrRegisterTimer("chain/storage/reads", nil)
	storageHashTimer   = metrics.GetOrRegisterTimer("chain/storage/hashes", nil)
	storageUpdateTimer = metrics.GetOrRegisterTimer("chain/storage/updates", nil)
	storageCommitTimer = metrics.GetOrRegisterTimer("chain/storage/commits", nil)

	snapshotAccountReadTimer = metrics.GetOrRegisterTimer("chain/snapshot/account/reads", nil)
	snapshotStorageReadTimer = metrics.GetOrRegisterTimer("chain/snapshot/storage/reads", nil)
	snapshotCommitTimer      = metrics.GetOrRegisterTimer("chain/snapshot/commits", nil)

	blockInsertTimer    = metrics.GetOrRegisterTimer("chain/inserts", nil)
	blockExecutionTimer = metrics.GetOrRegisterTimer("chain/execution", nil)
	blockWriteTimer     = metrics.GetOrRegisterTimer("chain/write", nil)
	blockAgeGauge       = metrics.GetOrRegisterGauge("chain/block/age", nil)
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
			&s.emitters,
			s.verWatcher,
			&s.bootstrapping,
			s.reporter,
		),
	}
}

// consensusCallbackBeginBlockFn takes only necessaries for block processing and
// makes lachesis.BeginBlockFn.
func consensusCallbackBeginBlockFn(
	parallelTasks *workers.Workers,
	wg *sync.WaitGroup,
	blockBusyFlag *uint32,
	store *Store,
	blockProc BlockProc,
	txIndex bool,
	feed *ServiceFeed,
	emitters *[]*emitter.Emitter,
	verWatcher *verwatcher.VerWarcher,
	bootstrapping *bool,
	xenblocks *reporter.Reporter,
) lachesis.BeginBlockFn {
	return func(cBlock *lachesis.Block) lachesis.BlockCallbacks {
		if *bootstrapping {
			// ignore block processing during bootstrapping
			return lachesis.BlockCallbacks{
				ApplyEvent: func(dag.Event) {},
				EndBlock: func() *pos.Validators {
					return nil
				},
			}
		}
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
		evmStateReader := &EvmStateReader{
			ServiceFeed: feed,
			store:       store,
		}

		eventProcessor := blockProc.EventsModule.Start(bs, es)

		atroposTime := bs.LastBlock.Time + 1
		atroposDegenerate := true
		// events with txs
		confirmedEvents := make(hash.OrderedEvents, 0, 3*es.Validators.Len())

		mpsCheatersMap := make(map[idx.ValidatorID]struct{})
		reportCheater := func(reporter, cheater idx.ValidatorID) {
			mpsCheatersMap[cheater] = struct{}{}
		}

		return lachesis.BlockCallbacks{
			ApplyEvent: func(_e dag.Event) {
				e := _e.(inter.EventI)
				if cBlock.Atropos == e.ID() {
					atroposTime = e.MedianTime()
					atroposDegenerate = false
				}
				if e.AnyTxs() {
					confirmedEvents = append(confirmedEvents, e.ID())
				}
				if e.AnyMisbehaviourProofs() {
					mps := store.GetEventPayload(e.ID()).MisbehaviourProofs()
					for _, mp := range mps {
						// self-contained parts of proofs are already checked by the checkers
						if proof := mp.BlockVoteDoublesign; proof != nil {
							reportCheater(e.Creator(), proof.Pair[0].Signed.Locator.Creator)
						}
						if proof := mp.EpochVoteDoublesign; proof != nil {
							reportCheater(e.Creator(), proof.Pair[0].Signed.Locator.Creator)
						}
						if proof := mp.EventsDoublesign; proof != nil {
							reportCheater(e.Creator(), proof.Pair[0].Locator.Creator)
						}
						if proof := mp.WrongBlockVote; proof != nil {
							// all other votes are the same, see MinAccomplicesForProof
							if proof.WrongEpoch {
								actualBlockEpoch := store.FindBlockEpoch(proof.Block)
								if actualBlockEpoch != 0 && actualBlockEpoch != proof.Pals[0].Val.Epoch {
									for _, pal := range proof.Pals {
										reportCheater(e.Creator(), pal.Signed.Locator.Creator)
									}
								}
							} else {
								actualRecordHash := store.GetBlockRecordHash(proof.Block)
								if actualRecordHash != nil && proof.GetVote(0) != *actualRecordHash {
									for _, pal := range proof.Pals {
										reportCheater(e.Creator(), pal.Signed.Locator.Creator)
									}
								}
							}
						}
						if proof := mp.WrongEpochVote; proof != nil {
							// all other votes are the same, see MinAccomplicesForProof
							vote := proof.Pals[0]
							actualRecord := store.GetFullEpochRecord(vote.Val.Epoch)
							if actualRecord == nil {
								continue
							}
							if vote.Val.Vote != actualRecord.Hash() {
								for _, pal := range proof.Pals {
									reportCheater(e.Creator(), pal.Signed.Locator.Creator)
								}
							}
						}
					}
				}
				eventProcessor.ProcessConfirmedEvent(e)
				for _, em := range *emitters {
					em.OnEventConfirmed(e)
				}
			},
			EndBlock: func() (newValidators *pos.Validators) {
				if atroposTime <= bs.LastBlock.Time {
					atroposTime = bs.LastBlock.Time + 1
				}
				blockCtx := iblockproc.BlockCtx{
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
				{                                                 // sort and merge MPs cheaters
					mpsCheaters := make(lachesis.Cheaters, 0, len(mpsCheatersMap))
					for vid := range mpsCheatersMap {
						mpsCheaters = append(mpsCheaters, vid)
					}
					sort.Slice(mpsCheaters, func(i, j int) bool {
						a, b := mpsCheaters[i], mpsCheaters[j]
						return a < b
					})
					bs.EpochCheaters = mergeCheaters(bs.EpochCheaters, mpsCheaters)
				}
				if skipBlock {
					// save the latest block state even if block is skipped
					store.SetBlockEpochState(bs, es)
					log.Debug("Frame is skipped", "atropos", cBlock.Atropos.String())
					return nil
				}

				sealer := blockProc.SealerModule.Start(blockCtx, bs, es)
				sealing := sealer.EpochSealing()
				txListener := blockProc.TxListenerModule.Start(blockCtx, bs, es, statedb)
				onNewLogAll := func(l *types.Log) {
					txListener.OnNewLog(l)
					// Note: it's possible for logs to get indexed twice by BR and block processing
					if verWatcher != nil {
						verWatcher.OnNewLog(l)
					}
				}

				// skip LLR block/epoch deciding if not activated
				if !es.Rules.Upgrades.Llr {
					store.ModifyLlrState(func(llrs *LlrState) {
						if llrs.LowestBlockToDecide == blockCtx.Idx {
							llrs.LowestBlockToDecide++
						}
						if sealing && es.Epoch+1 == llrs.LowestEpochToDecide {
							llrs.LowestEpochToDecide++
						}
					})
				}

				evmProcessor := blockProc.EVMModule.Start(blockCtx, statedb, evmStateReader, onNewLogAll, es.Rules, es.Rules.EvmChainConfig(store.GetUpgradeHeights()))
				executionStart := time.Now()

				// Execute pre-internal transactions
				preInternalTxs := blockProc.PreTxTransactor.PopInternalTxs(blockCtx, bs, es, sealing, statedb)
				preInternalReceipts := evmProcessor.Execute(preInternalTxs)
				bs = txListener.Finalize()
				for _, r := range preInternalReceipts {
					if r.Status == 0 {
						log.Warn("Pre-internal transaction reverted", "txid", r.TxHash.String())
					}
				}

				// Seal epoch if requested
				if sealing {
					sealer.Update(bs, es)
					prevUpg := es.Rules.Upgrades
					bs, es = sealer.SealEpoch() // TODO: refactor to not mutate the bs, it is unclear
					if es.Rules.Upgrades != prevUpg {
						store.AddUpgradeHeight(opera.UpgradeHeight{
							Upgrades: es.Rules.Upgrades,
							Height:   blockCtx.Idx + 1,
						})
					}
					store.SetBlockEpochState(bs, es)
					newValidators = es.Validators
					txListener.Update(bs, es)
				}

				// At this point, newValidators may be returned and the rest of the code may be executed in a parallel thread
				blockFn := func() {
					// Execute post-internal transactions
					internalTxs := blockProc.PostTxTransactor.PopInternalTxs(blockCtx, bs, es, sealing, statedb)
					internalReceipts := evmProcessor.Execute(internalTxs)
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
						block.Txs = append(block.Txs, tx.Hash())
					}

					block, blockEvents := spillBlockEvents(store, block, es.Rules)
					txs := make(types.Transactions, 0, blockEvents.Len()*10)
					for _, e := range blockEvents {
						txs = append(txs, e.Txs()...)
					}

					_ = evmProcessor.Execute(txs)

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
						// Note: it's possible for receipts to get indexed twice by BR and block processing
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
					bs.CheatersWritten = uint32(bs.EpochCheaters.Len())
					if sealing {
						store.SetHistoryBlockEpochState(es.Epoch, bs, es)
						store.SetEpochBlock(blockCtx.Idx+1, es.Epoch)
					}
					store.SetBlock(blockCtx.Idx, block)
					store.SetBlockIndex(block.Atropos, blockCtx.Idx)
					store.SetBlockEpochState(bs, es)
					store.EvmStore().SetCachedEvmBlock(blockCtx.Idx, evmBlock)
					updateLowestBlockToFill(blockCtx.Idx, store)
					updateLowestEpochToFill(es.Epoch, store)

					// Update the metrics touched during block processing
					accountReadTimer.Update(statedb.AccountReads)
					storageReadTimer.Update(statedb.StorageReads)
					accountUpdateTimer.Update(statedb.AccountUpdates)
					storageUpdateTimer.Update(statedb.StorageUpdates)
					snapshotAccountReadTimer.Update(statedb.SnapshotAccountReads)
					snapshotStorageReadTimer.Update(statedb.SnapshotStorageReads)
					accountHashTimer.Update(statedb.AccountHashes)
					storageHashTimer.Update(statedb.StorageHashes)
					triehash := statedb.AccountHashes + statedb.StorageHashes
					trieproc := statedb.SnapshotAccountReads + statedb.AccountReads + statedb.AccountUpdates
					trieproc += statedb.SnapshotStorageReads + statedb.StorageReads + statedb.StorageUpdates
					blockExecutionTimer.Update(time.Since(executionStart) - trieproc - triehash)

					// Update the metrics touched by new block
					headBlockGauge.Update(int64(blockCtx.Idx))
					headHeaderGauge.Update(int64(blockCtx.Idx))
					headFastBlockGauge.Update(int64(blockCtx.Idx))

					// Notify about new block
					if feed != nil {
						feed.newBlock.Send(evmcore.ChainHeadNotify{Block: evmBlock})
						var logs []*types.Log
						for _, r := range allReceipts {
							for _, l := range r.Logs {
								logs = append(logs, l)
							}
						}
						feed.newLogs.Send(logs)
					}

					commitStart := time.Now()
					store.commitEVM(false)

					// Update the metrics touched during block commit
					accountCommitTimer.Update(statedb.AccountCommits)
					storageCommitTimer.Update(statedb.StorageCommits)
					snapshotCommitTimer.Update(statedb.SnapshotCommits)
					blockWriteTimer.Update(time.Since(commitStart) - statedb.AccountCommits - statedb.StorageCommits - statedb.SnapshotCommits)
					blockInsertTimer.UpdateSince(start)

					now := time.Now()
					blockAge := now.Sub(block.Time.Time())
					log.Info("New block", "index", blockCtx.Idx, "id", block.Atropos, "gas_used",
						evmBlock.GasUsed, "txs", fmt.Sprintf("%d/%d", len(evmBlock.Transactions), len(block.SkippedTxs)),
						"age", utils.PrettyDuration(blockAge), "t", utils.PrettyDuration(now.Sub(start)))
					blockAgeGauge.Update(int64(blockAge.Nanoseconds()))

					if xenblocks.Enabled {
						xenblocks.Send(
							fmt.Sprintf("%d", blockCtx.Idx),
							fmt.Sprintf("%s", block.Atropos),
							fmt.Sprintf("%s", utils.PrettyDuration(now.Sub(start))))
					}
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

func (s *Service) ReexecuteBlocks(from, to idx.Block) {
	blockProc := s.blockProcModules
	upgradeHeights := s.store.GetUpgradeHeights()
	evmStateReader := s.GetEvmStateReader()
	prev := s.store.GetBlock(from)
	for b := from + 1; b <= to; b++ {
		block := s.store.GetBlock(b)
		blockCtx := iblockproc.BlockCtx{
			Idx:     b,
			Time:    block.Time,
			Atropos: block.Atropos,
		}
		statedb, err := s.store.evm.StateDB(prev.Root)
		if err != nil {
			log.Crit("Failue to re-execute blocks", "err", err)
		}
		es := s.store.GetHistoryEpochState(s.store.FindBlockEpoch(b))
		evmProcessor := blockProc.EVMModule.Start(blockCtx, statedb, evmStateReader, func(t *types.Log) {}, es.Rules, es.Rules.EvmChainConfig(upgradeHeights))
		txs := s.store.GetBlockTxs(b, block)
		evmProcessor.Execute(txs)
		evmProcessor.Finalize()
		_ = s.store.evm.Commit(b, block.Root, false)
		s.store.evm.Cap()
		s.mayCommit(false)
		prev = block
	}
}

func (s *Service) RecoverEVM() {
	start := s.store.GetLatestBlockIndex()
	for b := start; b >= 1 && b > start-20000; b-- {
		block := s.store.GetBlock(b)
		if block == nil {
			break
		}
		if s.store.evm.HasStateDB(block.Root) {
			if b != start {
				s.Log.Warn("Reexecuting blocks after abrupt stopping", "from", b, "to", start)
				s.ReexecuteBlocks(b, start)
			}
			break
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
