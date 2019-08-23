package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

type PackInfo struct {
	Index       idx.Pack
	Size        uint32
	NumOfEvents uint32
	Heads       hash.Events
}

const (
	maxPackSize      = softResponseLimitSize
	maxPackEventsNum = softLimitItems
)

func (s *Service) packs_onNewEvent(e *inter.Event, epoch idx.SuperFrame) {
	// due to default values, we don't need to explicitly set values at a start of an epoch
	packIdx := s.store.GetPacksNumOrDefault(epoch)
	packInfo := s.store.GetPackInfoOrDefault(s.engine.CurrentSuperFrameN(), packIdx)

	s.store.AddToPack(epoch, packIdx, e.Hash())

	packInfo.Index = packIdx
	packInfo.NumOfEvents += 1
	packInfo.Size += uint32(len(s.store.GetEventRLP(e.Hash()))) // TODO optimize (no need to do DB access)
	if packInfo.NumOfEvents >= maxPackEventsNum || packInfo.Size >= maxPackSize {
		// pin the s.store.GetHeads()
		packInfo.Heads = s.store.GetHeads(epoch)
		s.store.SetPacksNum(epoch, packIdx+1)

		_ = s.mux.Post(packIdx + 1) // notify about new pack
	}
	s.store.SetPackInfo(epoch, packIdx, packInfo)
}

func (s *Service) packs_onNewEpoch(oldEpoch, newEpoch idx.SuperFrame) {
	// pin the last pack
	packIdx := s.store.GetPacksNumOrDefault(oldEpoch)
	packInfo := s.store.GetPackInfoOrDefault(s.engine.CurrentSuperFrameN(), packIdx)

	packInfo.Heads = s.store.GetHeads(oldEpoch)
	s.store.SetPackInfo(oldEpoch, packIdx, packInfo)

	s.store.SetPacksNum(oldEpoch, packIdx+1) // the last pack is always not pinned, so create not pinned one

	_ = s.mux.Post(packIdx + 1)
}
