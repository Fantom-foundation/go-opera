package gossip

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"

	"github.com/Fantom-foundation/go-lachesis/eventcheck"
	"github.com/Fantom-foundation/go-lachesis/eventcheck/gaspowercheck"
	"github.com/Fantom-foundation/go-lachesis/evmcore"
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/tracing"
)

// processEvent extends the engine.ProcessEvent with gossip-specific actions on each event processing
func (s *Service) processEvent(realEngine Consensus, e *inter.Event) error {
	// s.engineMu is locked here

	if s.store.HasEvent(e.Hash()) { // sanity check
		return eventcheck.ErrAlreadyConnectedEvent
	}

	// we check gas power here, because engineMu is locked here
	if err := s.gasPowerCheck(e); err != nil {
		return err
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
		if s.store.IsHead(e.Epoch, parent) {
			s.store.DelHead(e.Epoch, parent)
		}
	}
	s.store.AddHead(e.Epoch, e.Hash())

	s.packsOnNewEvent(e, e.Epoch)
	s.emitter.OnNewEvent(e)

	newEpoch := oldEpoch
	if realEngine != nil {
		newEpoch = realEngine.GetEpoch()
	}

	if newEpoch != oldEpoch {
		s.packsOnNewEpoch(oldEpoch, newEpoch)
		s.store.delEpochStore(oldEpoch)
		s.store.getEpochStore(newEpoch)
		s.feed.newEpoch.Send(newEpoch)
		s.occurredTxs.Clear()
	}

	immediately := (newEpoch != oldEpoch)
	return s.store.Commit(e.Hash().Bytes(), immediately)
}

// applyNewState moves the state according to new block (txs execution, SFC logic, epoch sealing)
func (s *Service) applyNewState(
	block *inter.Block,
	sealEpoch bool,
	cheaters []common.Address,
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
	statedb := s.store.StateDB(stateHash)

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
	s.updateOriginationScores(block, receipts, txPositions, sealEpoch)
	s.updateValidationScores(block, sealEpoch)
	s.updateUsersPOI(block, evmBlock, receipts, sealEpoch)
	s.updateStakersPOI(block, sealEpoch)

	// Process SFC contract transactions
	s.processSfc(block, receipts, totalFee, sealEpoch, cheaters, statedb)

	// Process new epoch
	var newEpochHash common.Hash
	if sealEpoch {
		newEpochHash = s.onEpochSealed(block, cheaters)
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
	var newAppHash common.Hash
	if sealEpoch {
		newAppHash = hash.Of(newStateHash.Bytes(), newEpochHash.Bytes())
	} else {
		newAppHash = newStateHash
	}

	log.Info("New block", "index", block.Index, "atropos", block.Atropos, "fee", totalFee, "gasUsed",
		evmBlock.GasUsed, "txs", len(evmBlock.Transactions), "skipped_txs", len(block.SkippedTxs), "elapsed", time.Since(start))

	return block, evmBlock, receipts, txPositions, newAppHash
}

// assembleEvmBlock converts inter.Block to evmcore.EvmBlock (without skipped transactions)
func (s *Service) assembleEvmBlock(
	block *inter.Block,
) (*evmcore.EvmBlock, inter.Events) {
	// s.engineMu is locked here
	if len(block.SkippedTxs) != 0 {
		log.Crit("Building with SkippedTxs isn't supported")
	}
	blockEvents := make(inter.Events, 0, len(block.Events))

	// Assemble block data
	evmBlock := &evmcore.EvmBlock{
		EvmHeader:    *evmcore.ToEvmHeader(block),
		Transactions: make(types.Transactions, 0, len(block.Events)*10),
	}
	for _, id := range block.Events {
		e := s.store.GetEvent(id)
		if e == nil {
			s.Log.Crit("Event not found", "event", id.String())
		}

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

	evmProcessor := evmcore.NewStateProcessor(params.AllEthashProtocolChanges, s.GetEvmStateReader())

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

	return block, evmBlock, totalFee, receipts
}

// onEpochSealed applies the new epoch sealing state
func (s *Service) onEpochSealed(block *inter.Block, cheaters inter.Cheaters) (newEpochHash common.Hash) {
	// s.engineMu is locked here

	epoch := s.engine.GetEpoch()

	// update last headers
	for _, cheater := range cheaters {
		s.store.DelLastHeader(epoch, cheater) // for cheaters, it's uncertain which event is "last confirmed"
	}
	hh := s.store.GetLastHeaders(epoch)
	// After sealing, AppHash includes last confirmed headers in this epoch from each honest validator and cheaters list
	// TODO use transparent state hashing (i.e. store state in a trie)
	newEpochHash = hash.Of(newEpochHash.Bytes(), hash.Of(hh.Bytes()).Bytes(), types.DeriveSha(cheaters).Bytes())
	// prune not needed last headers
	s.store.DelLastHeaders(epoch - 1)

	return newEpochHash
}

// applyBlock execs ordered txns of new block on state, and fills the block DB indexes.
func (s *Service) applyBlock(block *inter.Block, decidedFrame idx.Frame, cheaters inter.Cheaters) (newAppHash common.Hash, sealEpoch bool) {
	// s.engineMu is locked here

	confirmBlocksMeter.Inc(1)
	sealEpoch = decidedFrame == s.config.Net.Dag.EpochLen

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
			s.store.SetReceipts(block.Index, receipts)
		}
	}

	// Notify about new block
	s.feed.newBlock.Send(evmcore.ChainHeadNotify{Block: evmBlock})

	// trace confirmed transactions
	for _, tx := range evmBlock.Transactions {
		tracing.FinishTx(tx.Hash(), "Service.onNewBlock()")
		if latency, err := txLatency.Finish(tx.Hash()); err == nil {
			confirmTxLatencyMeter.Update(latency.Milliseconds())
		}
	}

	s.blockParticipated = make(map[common.Address]bool) // reset map of participated validators

	return newAppHash, sealEpoch
}

// selectValidatorsGroup is a callback type to select new validators group
func (s *Service) selectValidatorsGroup(oldEpoch, newEpoch idx.Epoch) (newValidators *pos.Validators) {
	// s.engineMu is locked here

	newValidators = pos.NewValidators()
	for _, it := range s.store.GetEpochValidators(newEpoch) {
		newValidators.Set(it.Staker.Address, pos.BalanceToStake(it.Staker.CalcTotalStake()))
	}

	return newValidators
}

// onEventConfirmed is callback type to notify about event confirmation
func (s *Service) onEventConfirmed(header *inter.EventHeaderData, seqDepth idx.Event) {
	// s.engineMu is locked here

	if !header.NoTransactions() {
		// erase confirmed txs from originated-but-non-confirmed
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

/*
 * Calling gaspowercheck
 */

func (s *Service) gasPowerCheck(e *inter.Event) error {
	// s.engineMu is locked here

	gasPowerChecker := gaspowercheck.New(&s.config.Net.Dag.GasPower, &GasPowerCheckReader{
		Consensus: s.engine,
		store:     s.store,
	})
	var selfParent *inter.EventHeaderData
	if e.SelfParent() != nil {
		selfParent = s.store.GetEventHeader(e.Epoch, *e.SelfParent())
	}
	return gasPowerChecker.Validate(e, selfParent)
}

// GasPowerCheckReader is a helper to run gas power check
type GasPowerCheckReader struct {
	Consensus
	store *Store
}

// GetPrevEpochLastHeaders isn't safe for concurrent use
func (r *GasPowerCheckReader) GetPrevEpochLastHeaders() (inter.HeadersByCreator, idx.Epoch) {
	// engineMu is locked here
	epoch := r.GetEpoch() - 1
	return r.store.GetLastHeaders(epoch), epoch
}

// GetPrevEpochEndTime isn't safe for concurrent use
func (r *GasPowerCheckReader) GetPrevEpochEndTime() (inter.Timestamp, idx.Epoch) {
	// engineMu is locked here
	epoch := r.GetEpoch() - 1
	return r.store.GetEpochStats(epoch).End, epoch
}
