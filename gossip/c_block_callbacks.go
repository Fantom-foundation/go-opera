package gossip

import (
	"errors"
	"math/big"
	"sort"
	"time"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/Fantom-foundation/lachesis-base/lachesis"
)

var (
	errStopped         = errors.New("service is stopped")
	errWrongMedianTime = errors.New("wrong event median time")
)

func (s *Service) GetConsensusCallbacks() lachesis.ConsensusCallbacks {
	return lachesis.ConsensusCallbacks{
		BeginBlock: func(cBlock *lachesis.Block) lachesis.BlockCallbacks {

			start := time.Now()

			bs := s.store.GetBlockState()
			es := s.store.GetEpochState()

			var atropos inter.EventI
			var validatorHighestEvents = make(inter.EventIs, es.Validators.Len())
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
					creatorIdx := es.Validators.GetIdx(e.Creator())
					prev := validatorHighestEvents[creatorIdx]
					if prev == nil || e.Seq() > prev.Seq() {
						validatorHighestEvents[creatorIdx] = e
					}
				},
				EndBlock: func() (newValidators *pos.Validators) {
					// sort events by Lamport time
					sort.Sort(confirmedEvents)

					// new block
					var block = &inter.Block{
						Time:    atropos.MedianTime(),
						Atropos: cBlock.Atropos,
						Events:  hash.Events(confirmedEvents),
					}
					bs.Block++
					bs.EpochBlocks++

					for _, v := range cBlock.Cheaters {
						creatorIdx := es.Validators.GetIdx(v)
						validatorHighestEvents[creatorIdx] = nil
					}
					for creatorIdx, e := range validatorHighestEvents {
						if e == nil {
							continue
						}
						info := bs.ValidatorStates[creatorIdx]
						if bs.Block-info.PrevBlock <= s.config.Net.Economy.BlockMissedLatency {
							info.Uptime += e.MedianTime() - info.PrevMedianTime
						}
						info.PrevGasPowerLeft = e.GasPowerLeft()
						info.PrevMedianTime = e.MedianTime()
						info.PrevBlock = bs.Block
						info.PrevEvent = e.ID()
						bs.ValidatorStates[creatorIdx] = info
					}

					sealEpoch := bs.EpochBlocks >= s.config.Net.Dag.MaxEpochBlocks
					sealEpoch = sealEpoch || (block.Time-es.EpochStart) >= inter.Timestamp(s.config.Net.Dag.MaxEpochDuration)
					sealEpoch = sealEpoch || cBlock.Cheaters.Len() != 0

					block, evmBlock, receipts, txPositions := s.applyNewState(&bs, &es, block, sealEpoch, cBlock.Cheaters)

					// Build index for not skipped txs
					if s.config.TxIndex {
						for _, tx := range evmBlock.Transactions {
							// not skipped txs only
							position := txPositions[tx.Hash()]
							s.store.SetTxPosition(tx.Hash(), &position)
						}

						if receipts.Len() != 0 {
							s.store.app.SetReceipts(bs.Block, receipts)
						}
					}

					if sealEpoch {
						// seal epoch
						bs.EpochBlocks = 0
						newEpoch := es.Epoch + 1
						es.Epoch = newEpoch

						s.store.SetEpochState(es)

						newValidators = es.Validators // es.Validators are updated by s.applyNewState
					}
					s.store.SetBlock(bs.Block, block)
					s.store.SetBlockState(bs)

					log.Info("New block", "index", bs.Block, "atropos", block.Atropos, "gas_used",
						evmBlock.GasUsed, "skipped_txs", len(block.SkippedTxs), "txs", len(evmBlock.Transactions), "t", time.Since(start))

					// meter
					return newValidators
				},
			}
		},
	}
}

// applyNewState moves the state according to new block (txs execution, SFC logic, epoch sealing)
func (s *Service) applyNewState(
	bs *BlockState,
	es *EpochState,
	block *inter.Block,
	sealEpoch bool,
	cheaters lachesis.Cheaters,
) (
	*inter.Block,
	*evmcore.EvmBlock,
	types.Receipts,
	map[common.Hash]TxPosition,
) {
	// s.engineMu is locked here

	// Assemble block data
	evmBlock, blockEvents := s.assembleEvmBlock(block, bs.Block)

	// memorize position of each tx, for indexing and origination scores
	txPositions := make(map[common.Hash]TxPosition)
	for _, e := range blockEvents {
		for i, tx := range e.Txs() {
			// If tx was met in multiple events, then assign to first ordered event
			if _, ok := txPositions[tx.Hash()]; ok {
				continue
			}
			txPositions[tx.Hash()] = TxPosition{
				Event:       e.ID(),
				EventOffset: uint32(i),
			}
		}
	}

	// Get stateDB
	stateHash := s.store.GetBlock(bs.Block - 1).Root
	statedb := s.store.app.StateDB(stateHash)

	// Process EVM txs
	block, evmBlock, totalFee, receipts := s.executeEvmTransactions(block, evmBlock, statedb)

	// memorize block position of each tx, for indexing and origination scores
	for i, tx := range evmBlock.Transactions {
		// not skipped txs only
		position := txPositions[tx.Hash()]
		position.Block = bs.Block
		position.BlockOffset = uint32(i)
		txPositions[tx.Hash()] = position
	}

	// Process PoI/score changes
	s.updateOriginationScores(bs, es, evmBlock, receipts, txPositions)

	// Process SFC contract transactions
	s.processSfc(bs, es, block, receipts, totalFee, sealEpoch, cheaters, statedb)

	// Get state root
	newStateHash, err := statedb.Commit(true)
	if err != nil {
		s.Log.Crit("Failed to commit state", "err", err)
	}
	block.Root = hash.Hash(newStateHash)
	*evmBlock = evmcore.EvmBlock{
		EvmHeader:    *evmcore.ToEvmHeader(block, idx.Block(evmBlock.Number.Uint64()), hash.Event(evmBlock.ParentHash)),
		Transactions: evmBlock.Transactions,
	}

	return block, evmBlock, receipts, txPositions
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
			s.Log.Crit("Event not found", "event", id.String())
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

// assembleEvmBlock converts inter.Block to evmcore.EvmBlock
func (s *Service) assembleEvmBlock(
	block *inter.Block,
	index idx.Block,
) (*evmcore.EvmBlock, inter.EventPayloads) {
	// s.engineMu is locked here
	if len(block.SkippedTxs) != 0 {
		log.Crit("Building with SkippedTxs isn't supported")
	}
	block, blockEvents := s.spillBlockEvents(block)

	var prev hash.Event
	if index != 0 {
		prev = s.store.GetBlock(index - 1).Atropos
	}
	// Assemble block data
	evmBlock := &evmcore.EvmBlock{
		EvmHeader:    *evmcore.ToEvmHeader(block, index, prev),
		Transactions: make(types.Transactions, 0, len(block.Events)*10),
	}
	for _, e := range blockEvents {
		evmBlock.Transactions = append(evmBlock.Transactions, e.Txs()...)
		blockEvents = append(blockEvents, e)
	}

	return evmBlock, blockEvents
}

func filterSkippedTxs(block *inter.Block, evmBlock *evmcore.EvmBlock) *evmcore.EvmBlock {
	// Filter skipped transactions. Receipts are filtered already
	skipCount := 0
	filteredTxs := make(types.Transactions, 0, len(evmBlock.Transactions))
	for i, tx := range evmBlock.Transactions {
		if skipCount < len(block.SkippedTxs) && block.SkippedTxs[skipCount] == uint32(i) {
			skipCount++
		} else {
			filteredTxs = append(filteredTxs, tx)
		}
	}
	evmBlock.Transactions = filteredTxs
	return evmBlock
}

// executeTransactions execs ordered txns of new block on state.
func (s *Service) executeEvmTransactions(
	block *inter.Block,
	evmBlock *evmcore.EvmBlock,
	statedb *state.StateDB,
) (
	*inter.Block,
	*evmcore.EvmBlock,
	*big.Int,
	types.Receipts,
) {
	// s.engineMu is locked here

	evmProcessor := evmcore.NewStateProcessor(s.config.Net.EvmChainConfig(), s.GetEvmStateReader())

	// Process txs
	receipts, _, gasUsed, totalFee, skipped, err := evmProcessor.Process(evmBlock, statedb, vm.Config{}, false)
	if err != nil {
		s.Log.Crit("Shouldn't happen ever because it's not strict", "err", err)
	}
	block.SkippedTxs = skipped
	block.GasUsed = gasUsed

	// Filter skipped transactions
	evmBlock = filterSkippedTxs(block, evmBlock)

	*evmBlock = evmcore.EvmBlock{
		EvmHeader:    *evmcore.ToEvmHeader(block, idx.Block(evmBlock.Number.Uint64()), hash.Event(evmBlock.ParentHash)),
		Transactions: evmBlock.Transactions,
	}

	for _, r := range receipts {
		s.store.app.IndexLogs(r.Logs...)
	}

	return block, evmBlock, totalFee, receipts
}
