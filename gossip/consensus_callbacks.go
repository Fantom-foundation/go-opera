package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/eventcheck"
	"github.com/Fantom-foundation/go-lachesis/tracing"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"

	"github.com/Fantom-foundation/go-lachesis/evmcore"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
)

// processEvent extends the engine.ProcessEvent with gossip-specific actions on each event processing
func (s *Service) processEvent(realEngine Consensus, e *inter.Event) error {
	// s.engineMu is locked here

	if s.store.HasEvent(e.Hash()) { // sanity check
		return eventcheck.ErrAlreadyConnectedEvent
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

// executeTransactions execs ordered txns of new block on state.
func (s *Service) executeTransactions(arg inter.ApplyBlockArgs, forEachEvent func(*inter.Event)) (*inter.Block, *evmcore.EvmBlock, types.Receipts) {
	evmProcessor := evmcore.NewStateProcessor(params.AllEthashProtocolChanges, s.GetEvmStateReader())

	// Assemble block data
	evmBlock := &evmcore.EvmBlock{
		EvmHeader:    *evmcore.ToEvmHeader(arg.Block),
		Transactions: make(types.Transactions, 0, len(arg.Block.Events)*10),
	}
	for _, id := range arg.Block.Events {
		e := s.store.GetEvent(id)
		if e == nil {
			s.Log.Crit("Event not found", "event", id.String())
		}

		evmBlock.Transactions = append(evmBlock.Transactions, e.Transactions...)
		if forEachEvent != nil {
			forEachEvent(e)
		}
	}

	// Process txs
	statedb := s.store.StateDB(arg.StateHash)
	receipts, _, gasUsed, totalFee, skipped, err := evmProcessor.Process(evmBlock, statedb, vm.Config{}, false)
	if err != nil {
		s.Log.Crit("Shouldn't happen ever because it's not strict", "err", err)
	}
	arg.Block.SkippedTxs = skipped
	arg.Block.GasUsed = gasUsed

	// finalize
	newStateHash, err := statedb.Commit(true)
	if err != nil {
		s.Log.Crit("Failed to commit state", "err", err)
	}
	arg.Block.Root = newStateHash

	log.Info("New block", "index", arg.Block.Index, "hash", arg.Block.Hash().String(), "fee", totalFee, "txs", len(evmBlock.Transactions), "skipped_txs", len(skipped))

	// Filter skipped transactions. Receipts are filtered already
	skipCount := 0
	filteredTxs := make(types.Transactions, 0, len(evmBlock.Transactions))
	for i, tx := range evmBlock.Transactions {
		if skipCount < len(arg.Block.SkippedTxs) && arg.Block.SkippedTxs[skipCount] == uint(i) {
			skipCount++
		} else {
			filteredTxs = append(filteredTxs, tx)
		}
	}

	// Calc Merkle root of transactions
	arg.Block.TxHash = types.DeriveSha(filteredTxs)

	*evmBlock = evmcore.EvmBlock{
		EvmHeader:    *evmcore.ToEvmHeader(arg.Block),
		Transactions: filteredTxs,
	}
	return arg.Block, evmBlock, receipts
}

// applyBlock execs ordered txns of new block on state, and fills the block DB indexes.
func (s *Service) applyBlock(arg inter.ApplyBlockArgs) (newStateHash common.Hash, sealEpoch bool) {
	confirmBlocksMeter.Inc(1)

	// memorize position of each tx, for later indexing
	var txPositions map[common.Hash]TxPosition
	var forEachBlockEvent func(e *inter.Event)
	if s.config.TxIndex {
		txPositions = make(map[common.Hash]TxPosition)
		forEachBlockEvent = func(e *inter.Event) {
			for i, tx := range e.Transactions {
				// we don't care if tx was met in multiple events, any valid position will work
				txPositions[tx.Hash()] = TxPosition{
					Event:       e.Hash(),
					EventOffset: uint32(i),
				}
			}
		}
	}

	block, evmBlock, receipts := s.executeTransactions(arg, forEachBlockEvent)

	s.store.SetBlock(block)
	s.store.SetBlockIndex(block.Hash(), arg.Block.Index)
	newStateHash = block.Root
	sealEpoch = arg.DecidedFrame == s.config.Net.Dag.EpochLen

	// Build index for not skipped txs
	if s.config.TxIndex {
		for i, tx := range evmBlock.Transactions {
			// not skipped txs only
			position := txPositions[tx.Hash()]
			position.Block = block.Index
			position.BlockOffset = uint32(i)
			s.store.SetTxPosition(tx.Hash(), &position)
		}

		if receipts.Len() != 0 {
			s.store.SetReceipts(block.Index, receipts)
		}
	}

	// Notify about new block
	s.feed.newBlock.Send(evmcore.ChainHeadNotify{evmBlock})

	// trace confirmed transactions
	for _, tx := range evmBlock.Transactions {
		tracing.FinishTx(tx.Hash(), "Service.onNewBlock()")
		if latency, err := txLatency.Finish(tx.Hash()); err == nil {
			confirmTxLatencyMeter.Update(latency.Milliseconds())
		}
	}

	return newStateHash, sealEpoch
}

// selectValidatorsGroup is a callback type to select new validators group
func (s *Service) selectValidatorsGroup(oldEpoch, newEpoch idx.Epoch) (newValidators pos.Validators) {
	// new validators calculation
	// TODO replace with SFC transactions for changing validators state
	// TODO the schema below doesn't work in all the cases, and intended only for testing
	{
		newValidators = s.engine.GetValidators().Copy()
		statedb := s.store.StateDB(s.GetEvmStateReader().CurrentHeader().Root)
		for addr := range newValidators.Iterate() {
			stake := pos.BalanceToStake(statedb.GetBalance(addr))
			newValidators.Set(addr, stake)
		}
	}
	return newValidators
}

// onEventConfirmed is callback type to notify about event confirmation
func (s *Service) onEventConfirmed(header *inter.EventHeaderData) {
	if !header.NoTransactions() {
		// erase confirmed txs from originated-but-non-confirmed
		event := s.store.GetEvent(header.Hash())
		s.occurredTxs.CollectConfirmedTxs(event.Transactions)
	}
}

// isEventAllowedIntoBlock is callback type to check is event may be within block or not
func (s *Service) isEventAllowedIntoBlock(header *inter.EventHeaderData, highestCreatorSeq idx.Event) bool {
	if header.NoTransactions() {
		return false // block contains only non-empty events to speed up block retrieving and processing
	}
	if header.Seq+s.config.Net.Dag.MaxValidatorEventsInBlock <= highestCreatorSeq {
		return false // block contains only MaxValidatorEventsInBlock highest events from a creator to prevent huge blocks
	}
	return true
}
