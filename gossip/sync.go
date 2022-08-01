package gossip

import (
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

type syncStage uint32

type syncStatus struct {
	stage       uint32
	maybeSynced uint32
}

const (
	ssUnknown syncStage = iota
	ssSnaps
	ssEvmSnapGen
	ssEvents
)

const (
	snapsyncMinEndAge   = 14 * 24 * time.Hour
	snapsyncMaxStartAge = 6 * time.Hour
)

func (ss *syncStatus) Is(s ...syncStage) bool {
	self := &ss.stage
	for _, v := range s {
		if atomic.LoadUint32(self) == uint32(v) {
			return true
		}
	}
	return false
}

func (ss *syncStatus) Set(s syncStage) {
	atomic.StoreUint32(&ss.stage, uint32(s))
}

func (ss *syncStatus) MaybeSynced() bool {
	return atomic.LoadUint32(&ss.maybeSynced) != 0
}

func (ss *syncStatus) MarkMaybeSynced() {
	atomic.StoreUint32(&ss.maybeSynced, uint32(1))
}

func (ss *syncStatus) AcceptEvents() bool {
	return ss.Is(ssEvents)
}

func (ss *syncStatus) AcceptBlockRecords() bool {
	return !ss.Is(ssEvents)
}

func (ss *syncStatus) AcceptTxs() bool {
	return ss.MaybeSynced() && ss.Is(ssEvents)
}

func (ss *syncStatus) RequestLLR() bool {
	return !ss.Is(ssEvents) || ss.MaybeSynced()
}

type txsync struct {
	p     *peer
	txids []common.Hash
}

// syncTransactions starts sending all currently pending transactions to the given peer.
func (h *handler) syncTransactions(p *peer, txids []common.Hash) {
	if len(txids) == 0 {
		return
	}
	select {
	case h.txsyncCh <- &txsync{p, txids}:
	case <-h.quitSync:
	}
}

// txsyncLoop takes care of the initial transaction sync for each new
// connection. When a new peer appears, we relay all currently pending
// transactions. In order to minimise egress bandwidth usage, we send
// the transactions in small packs to one peer at a time.
func (h *handler) txsyncLoop() {
	var (
		pending = make(map[enode.ID]*txsync)
		sending = false               // whether a send is active
		pack    = new(txsync)         // the pack that is being sent
		done    = make(chan error, 1) // result of the send
	)

	// send starts a sending a pack of transactions from the sync.
	send := func(s *txsync) {
		// Fill pack with transactions up to the target size.
		pack.p = s.p
		pack.txids = pack.txids[:0]
		for i := 0; i < len(s.txids) && len(pack.txids) < softLimitItems; i++ {
			pack.txids = append(pack.txids, s.txids[i])
		}
		// Remove the transactions that will be sent.
		s.txids = s.txids[len(pack.txids):]
		if len(s.txids) == 0 {
			delete(pending, s.p.ID())
		}
		// Send the pack in the background.
		s.p.Log().Trace("Sending batch of transaction hashes", "count", len(pack.txids))
		sending = true
		go func() {
			if len(pack.txids) != 0 {
				done <- pack.p.SendTransactionHashes(pack.txids)
			} else {
				done <- nil
			}
		}()
	}

	// pick chooses the next pending sync.
	pick := func() *txsync {
		if len(pending) == 0 {
			return nil
		}
		n := rand.Intn(len(pending)) + 1
		for _, s := range pending {
			if n--; n == 0 {
				return s
			}
		}
		return nil
	}

	for {
		select {
		case s := <-h.txsyncCh:
			pending[s.p.ID()] = s
			if !sending {
				send(s)
			}
		case err := <-done:
			sending = false
			// Stop tracking peers that cause send failures.
			if err != nil {
				pack.p.Log().Debug("Transaction send failed", "err", err)
				delete(pending, pack.p.ID())
			}
			// Schedule the next send.
			if s := pick(); s != nil {
				send(s)
			}
		case <-h.quitSync:
			return
		}
	}
}

func (h *handler) updateSnapsyncStage() {
	// never allow fullsync while EVM snap is still generating, as it may lead to a race condition
	snapGenOngoing, _ := h.store.evm.Snaps.Generating()
	fullsyncPossibleEver := h.store.evm.HasStateDB(h.store.GetBlockState().FinalizedStateRoot)
	fullsyncPossibleNow := fullsyncPossibleEver && !snapGenOngoing
	// never allow to stop fullsync as it may lead to a race condition due to overwritten EVM snapshot by snapsync
	snapsyncPossible := h.config.AllowSnapsync && (h.syncStatus.Is(ssUnknown) || h.syncStatus.Is(ssSnaps))
	snapsyncNeeded := !fullsyncPossibleEver || time.Since(h.store.GetEpochState().EpochStart.Time()) > snapsyncMinEndAge

	if snapsyncPossible && snapsyncNeeded {
		h.syncStatus.Set(ssSnaps)
	} else if snapGenOngoing {
		h.syncStatus.Set(ssEvmSnapGen)
	} else if fullsyncPossibleNow {
		if !h.syncStatus.Is(ssEvents) {
			h.Log.Info("Start/Switch to fullsync mode...")
		}
		h.syncStatus.Set(ssEvents)
	}
}

func (h *handler) snapsyncStageTick() {
	// check if existing snapsync process can be resulted
	h.updateSnapsyncStage()
	llrs := h.store.GetLlrState()
	if h.syncStatus.Is(ssSnaps) {
		for i := 0; i < 3; i++ {
			epoch := llrs.LowestEpochToFill - 1 - idx.Epoch(i)
			if epoch <= h.store.GetEpoch() {
				continue
			}
			bs, _ := h.store.GetHistoryBlockEpochState(epoch)
			if bs == nil {
				continue
			}
			if !h.store.evm.HasStateDB(bs.FinalizedStateRoot) {
				continue
			}
			if llrs.LowestBlockToFill <= bs.LastBlock.Idx {
				continue
			}
			if time.Since(bs.LastBlock.Time.Time()) > snapsyncMinEndAge {
				continue
			}
			// cancel snapsync activity to prevent race condition
			done := make(chan struct{})
			h.snapState.updatesCh <- snapsyncStateUpd{
				snapsyncCancelCmd: &snapsyncCancelCmd{done},
			}
			<-done
			// finalize snapsync
			if err := h.process.SwitchEpochTo(epoch); err != nil {
				h.Log.Error("Failed to result snapsync", "epoch", epoch, "block", bs.LastBlock.Idx, "err", err)
			} else {
				h.Log.Info("Snapsync is finalized at", "epoch", epoch, "block", bs.LastBlock.Idx, "root", bs.FinalizedStateRoot)
				// switch state to non-snapsync and thus not allow ssSnaps ever again
				h.syncStatus.Set(ssEvmSnapGen)
			}
		}
	}
	// push new data into an existing snapsync process
	if h.syncStatus.Is(ssSnaps) {
		lastEpoch := llrs.LowestEpochToFill - 1
		lastBs, _ := h.store.GetHistoryBlockEpochState(lastEpoch)
		if lastBs != nil && time.Since(lastBs.LastBlock.Time.Time()) < snapsyncMaxStartAge {
			h.snapState.updatesCh <- snapsyncStateUpd{
				snapsyncEpochUpd: &snapsyncEpochUpd{
					epoch: lastEpoch,
					root:  common.Hash(lastBs.FinalizedStateRoot),
				},
			}
		}
	}
	// resume events downloading if events sync is enabled
	if h.syncStatus.Is(ssEvents) {
		h.dagLeecher.Resume()
		h.brLeecher.Pause()
	} else {
		h.dagLeecher.Pause()
		h.brLeecher.Resume()
	}
}

func (h *handler) snapsyncStageLoop() {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	defer h.loopsWg.Done()
	for {
		select {
		case <-ticker.C:
			h.snapsyncStageTick()
		case <-h.snapState.quit:
			return
		}
	}
}

// mayCancel cancels existing snapsync process if any
func (ss *snapsyncState) mayCancel() error {
	if ss.cancel != nil {
		err := ss.cancel()
		ss.cancel = nil
		return err
	}
	return nil
}

func (h *handler) snapsyncStateLoop() {
	defer h.loopsWg.Done()
	for {
		select {
		case cmd := <-h.snapState.updatesCh:
			if cmd.snapsyncEpochUpd != nil {
				upd := cmd.snapsyncEpochUpd
				// check if epoch has advanced
				if h.snapState.epoch >= upd.epoch {
					continue
				}
				h.snapState.epoch = upd.epoch
				_ = h.snapState.mayCancel()
				// start new snapsync state
				h.Log.Info("Update snapsync epoch", "epoch", upd.epoch, "root", upd.root)
				h.process.PauseEvmSnapshot()
				ss := h.snapLeecher.SyncState(upd.root)
				h.snapState.cancel = ss.Cancel
			}
			if cmd.snapsyncCancelCmd != nil {
				_ = h.snapState.mayCancel()
				cmd.snapsyncCancelCmd.done <- struct{}{}
			}
		case <-h.snapState.quit:
			_ = h.snapState.mayCancel()
			return
		}
	}
}
