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

func (s *Service) packs_onNewEvent(e *inter.Event) {
	epoch := s.engine.CurrentSuperFrameN()
	packIdx := s.store.GetPacksNum(epoch)
	packInfo := s.store.GetPackInfo(s.engine.CurrentSuperFrameN(), packIdx)

	s.store.AddToPack(epoch, packIdx, e.Hash())

	packInfo.NumOfEvents += 1
	packInfo.Size += uint32(len(s.store.GetEventRLP(e.Hash()))) // TODO optimize (no need to do DB access)
	if packInfo.NumOfEvents >= maxPackEventsNum || packInfo.Size >= maxPackSize {
		// pin the s.store.GetHeads()
		packInfo.Heads = s.store.GetHeads()
		s.store.SetPacksNum(epoch, packIdx + 1)

		_ = s.mux.Post(packIdx + 1) // notify about new pack
	}
	s.store.SetPackInfo(epoch, packIdx, packInfo)
}
