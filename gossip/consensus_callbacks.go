package gossip

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-lachesis/eventcheck"
	"github.com/Fantom-foundation/go-lachesis/evmcore"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/tracing"
)

// processEvent extends the engine.ProcessEvent with gossip-specific actions on each event processing
func (s *Service) processEvent(realEngine Consensus, e *inter.Event) error {
	// s.engineMu is locked here

	if s.store.HasEventHeader(e.Hash()) { // sanity check
		return eventcheck.ErrAlreadyConnectedEvent
	}

	// Trace arrival time of events
	if s.config.EventLocalTimeIndex {
		s.store.SetEventReceivingTime(e.Hash(), inter.Timestamp(time.Now().UnixNano()))
	}
	if s.config.DecisiveEventsIndex {
		s.currentEvent = e.Hash()
	}

	oldEpoch := e.Epoch

	s.store.SetEvent(e)
	if realEngine != nil {
		err := realEngine.ProcessEvent(e)
		if err != nil { // TODO make it possible to write only on success
			s.store.DeleteEvent(e.Epoch, e.Hash())
			return err
		}
	}
	_ = s.occurredTxs.CollectNotConfirmedTxs(e.Transactions)

	// set validator's last event. we don't care about forks, because this index is used only for emitter
	s.store.SetLastEvent(e.Epoch, e.Creator, e.Hash())

	// track events with no descendants, i.e. heads
	for _, parent := range e.Parents {
		s.store.DelHead(e.Epoch, parent)
	}
	s.store.AddHead(e.Epoch, e.Hash())

	s.packsOnNewEvent(e, e.Epoch)
	s.emitter.OnNewEvent(e)

	newEpoch := oldEpoch
	if realEngine != nil {
		newEpoch = realEngine.GetEpoch()
	}

	if newEpoch != oldEpoch {
		// notify event checkers about new validation data
		s.heavyCheckReader.Addrs.Store(ReadEpochPubKeys(s.app, newEpoch))
		s.gasPowerCheckReader.Ctx.Store(ReadGasPowerContext(s.store, s.app, s.engine.GetValidators(), newEpoch, &s.config.Net.Economy))

		// sealings/prunings
		s.packsOnNewEpoch(oldEpoch, newEpoch)
		s.store.delEpochStore(oldEpoch)
		s.store.getEpochStore(newEpoch)
		s.occurredTxs.Clear()

		// notify about new epoch after event connection
		s.emitter.OnNewEpoch(s.engine.GetValidators(), newEpoch)
		s.feed.newEpoch.Send(newEpoch)
	}

	immediately := (newEpoch != oldEpoch)

	err := s.app.Commit(e.Hash().Bytes(), immediately)
	if err != nil {
		return err
	}
	return s.store.Commit(e.Hash().Bytes(), immediately)
}

// applyNewState moves the state according to new block (txs execution, SFC logic, epoch sealing)
func (s *Service) applyNewState(
	block *inter.Block,
	sealEpoch bool,
	cheaters inter.Cheaters,
) (
	*inter.Block,
	*evmcore.EvmBlock,
	types.Receipts,
	map[common.Hash]TxPosition,
	common.Hash,
) {
	// s.engineMu is locked here

	start := time.Now()

	// Assemble block data
	evmBlock, blockEvents := s.assembleEvmBlock(block)

	// memorize position of each tx, for indexing and origination scores
	txPositions := make(map[common.Hash]TxPosition)
	for _, e := range blockEvents {
		for i, tx := range e.Transactions {
			// If tx was met in multiple events, then assign to first ordered event
			if _, ok := txPositions[tx.Hash()]; ok {
				continue
			}
			txPositions[tx.Hash()] = TxPosition{
				Event:       e.Hash(),
				EventOffset: uint32(i),
			}
		}
	}

	// Get stateDB
	stateHash := s.store.GetBlock(block.Index - 1).Root
	statedb := s.app.StateDB(stateHash)

	// Process EVM txs
	block, evmBlock, totalFee, receipts := s.executeEvmTransactions(block, evmBlock, statedb)

	// memorize block position of each tx, for indexing and origination scores
	for i, tx := range evmBlock.Transactions {
		// not skipped txs only
		position := txPositions[tx.Hash()]
		position.Block = block.Index
		position.BlockOffset = uint32(i)
		txPositions[tx.Hash()] = position
	}

	// Process PoI/score changes
	s.updateOriginationScores(block, evmBlock, receipts, txPositions, sealEpoch)
	s.updateValidationScores(block, sealEpoch)
	s.updateUsersPOI(block, evmBlock, receipts, totalFee, sealEpoch)
	s.updateStakersPOI(block, sealEpoch)

	// Process SFC contract transactions
	s.processSfc(block, receipts, totalFee, sealEpoch, cheaters, statedb)

	// Process new epoch
	if sealEpoch {
		s.onEpochSealed(block, cheaters)
	}

	// Get state root
	newStateHash, err := statedb.Commit(true)
	if err != nil {
		s.Log.Crit("Failed to commit state", "err", err)
	}
	block.Root = newStateHash
	*evmBlock = evmcore.EvmBlock{
		EvmHeader:    *evmcore.ToEvmHeader(block),
		Transactions: evmBlock.Transactions,
	}

	// calc appHash
	appHash := block.TxHash

	log.Info("New block", "index", block.Index, "atropos", block.Atropos, "fee", totalFee, "gasUsed",
		evmBlock.GasUsed, "skipped_txs", len(block.SkippedTxs), "txs", len(evmBlock.Transactions), "t", time.Since(start))

	return block, evmBlock, receipts, txPositions, appHash
}

// spillBlockEvents excludes first events which exceed BlockGasHardLimit
func (s *Service) spillBlockEvents(block *inter.Block) (*inter.Block, inter.Events) {
	fullEvents := make(inter.Events, len(block.Events))
	if len(block.Events) == 0 {
		return block, fullEvents
	}
	gasPowerUsedSum := uint64(0)
	// iterate in reversed order
	for i := len(block.Events) - 1; ; i-- {
		id := block.Events[i]
		e := s.store.GetEvent(id)
		if e == nil {
			s.Log.Crit("Event not found", "event", id.String())
		}
		fullEvents[i] = e
		gasPowerUsedSum += e.GasPowerUsed
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
) (*evmcore.EvmBlock, inter.Events) {
	// s.engineMu is locked here
	if len(block.SkippedTxs) != 0 {
		log.Crit("Building with SkippedTxs isn't supported")
	}
	block, blockEvents := s.spillBlockEvents(block)

	// Assemble block data
	evmBlock := &evmcore.EvmBlock{
		EvmHeader:    *evmcore.ToEvmHeader(block),
		Transactions: make(types.Transactions, 0, len(block.Events)*10),
	}
	for _, e := range blockEvents {
		evmBlock.Transactions = append(evmBlock.Transactions, e.Transactions...)
		blockEvents = append(blockEvents, e)
	}

	return evmBlock, blockEvents
}

func filterSkippedTxs(block *inter.Block, evmBlock *evmcore.EvmBlock) *evmcore.EvmBlock {
	// Filter skipped transactions. Receipts are filtered already
	skipCount := 0
	filteredTxs := make(types.Transactions, 0, len(evmBlock.Transactions))
	for i, tx := range evmBlock.Transactions {
		if skipCount < len(block.SkippedTxs) && block.SkippedTxs[skipCount] == uint(i) {
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

	block.TxHash = types.DeriveSha(evmBlock.Transactions)
	*evmBlock = evmcore.EvmBlock{
		EvmHeader:    *evmcore.ToEvmHeader(block),
		Transactions: evmBlock.Transactions,
	}

	for _, r := range receipts {
		s.app.IndexLogs(r.Logs...)
	}

	return block, evmBlock, totalFee, receipts
}

// onEpochSealed applies the new epoch sealing state
func (s *Service) onEpochSealed(block *inter.Block, cheaters inter.Cheaters) {
	// s.engineMu is locked here

	epoch := s.engine.GetEpoch()

	// delete last headers of cheaters
	for _, cheater := range cheaters {
		s.store.DelLastHeader(epoch, cheater) // for cheaters, it's uncertain which event is "last confirmed"
	}
	// prune not needed last headers
	s.store.DelLastHeaders(epoch - 1)
}

// applyBlock execs ordered txns of new block on state, and fills the block DB indexes.
func (s *Service) applyBlock(block *inter.Block, decidedFrame idx.Frame, cheaters inter.Cheaters) (newAppHash common.Hash, sealEpoch bool) {
	// s.engineMu is locked here

	confirmBlocksMeter.Inc(1)
	// if cheater is confirmed, seal epoch right away to prune them from of BFT validators list

	epochStart := s.store.GetEpochStats(pendingEpoch).Start
	sealEpoch = decidedFrame >= s.config.Net.Dag.MaxEpochBlocks
	sealEpoch = sealEpoch || block.Time-epochStart >= inter.Timestamp(s.config.Net.Dag.MaxEpochDuration)
	sealEpoch = sealEpoch || cheaters.Len() > 0

	block, evmBlock, receipts, txPositions, newAppHash := s.applyNewState(block, sealEpoch, cheaters)

	s.store.SetBlock(block)
	s.store.SetBlockIndex(block.Atropos, block.Index)

	// Build index for not skipped txs
	if s.config.TxIndex {
		for _, tx := range evmBlock.Transactions {
			// not skipped txs only
			position := txPositions[tx.Hash()]
			s.store.SetTxPosition(tx.Hash(), &position)
		}

		if receipts.Len() != 0 {
			s.app.SetReceipts(block.Index, receipts)
		}
	}

	var logs []*types.Log
	for _, r := range receipts {
		for _, l := range r.Logs {
			logs = append(logs, l)
		}
	}

	// Notify about new block and txs
	s.feed.newBlock.Send(evmcore.ChainHeadNotify{Block: evmBlock})
	s.feed.newTxs.Send(core.NewTxsEvent{Txs: evmBlock.Transactions})
	s.feed.newLogs.Send(logs)

	// Trace by which event this block was confirmed (only for API)
	if s.config.DecisiveEventsIndex {
		s.store.SetBlockDecidedBy(block.Index, s.currentEvent)
	}

	// trace confirmed transactions
	confirmTxnsMeter.Inc(int64(evmBlock.Transactions.Len()))
	for _, tx := range evmBlock.Transactions {
		tracing.FinishTx(tx.Hash(), "Service.onNewBlock()")
		if latency, err := txLatency.Finish(tx.Hash()); err == nil {
			txTtfMeter.Update(latency.Milliseconds())
		}
	}

	s.blockParticipated = make(map[idx.StakerID]bool) // reset map of participated validators

	return newAppHash, sealEpoch
}

// selectValidatorsGroup is a callback type to select new validators group
func (s *Service) selectValidatorsGroup(oldEpoch, newEpoch idx.Epoch) (newValidators *pos.Validators) {
	// s.engineMu is locked here

	builder := pos.NewBuilder()
	for _, it := range s.app.GetEpochValidators(newEpoch) {
		builder.Set(it.StakerID, pos.BalanceToStake(it.Staker.CalcTotalStake()))
	}

	return builder.Build()
}

// onEventConfirmed is callback type to notify about event confirmation
func (s *Service) onEventConfirmed(header *inter.EventHeaderData, seqDepth idx.Event) {
	// s.engineMu is locked here

	if !header.NoTransactions() {
		// erase confirmed txs from originated-but-non-confirmed
		// to allow to re-originate this transaction if it will get skipped or spilled
		event := s.store.GetEvent(header.Hash())
		s.occurredTxs.CollectConfirmedTxs(event.Transactions)
	}

	// track last confirmed events from each validator
	if seqDepth == 0 {
		s.store.AddLastHeader(header.Epoch, header)
	}

	// track validators who participated in the block
	s.blockParticipated[header.Creator] = true
}

// isEventAllowedIntoBlock is callback type to check is event may be within block or not
func (s *Service) isEventAllowedIntoBlock(header *inter.EventHeaderData, seqDepth idx.Event) bool {
	// s.engineMu is locked here

	if header.NoTransactions() {
		return false // block contains only non-empty events to speed up block retrieving and processing
	}
	if seqDepth > s.config.Net.Dag.MaxValidatorEventsInBlock {
		return false // block contains only MaxValidatorEventsInBlock highest events from a creator to prevent huge blocks
	}
	return true
}
