package gossip

import (
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/log"
)

type LlrState struct {
	LowestEpochToDecide idx.Epoch
	LowestEpochToFill   idx.Epoch

	LowestBlockToDecide idx.Block
	LowestBlockToFill   idx.Block
}

func (s *Store) setLlrState(llrs LlrState) {
	s.cache.LlrState.Store(&llrs)
}

func (s *Store) ModifyLlrState(f func(*LlrState)) {
	s.mutex.WriteLlrState.Lock()
	defer s.mutex.WriteLlrState.Unlock()
	llrs := s.GetLlrState()
	f(&llrs)
	s.setLlrState(llrs)
}

func (s *Store) GetLlrState() LlrState {
	if v := s.cache.LlrState.Load(); v != nil {
		return *v.(*LlrState)
	}
	v, ok := s.rlp.Get(s.table.LlrState, []byte{}, &LlrState{}).(*LlrState)
	if !ok {
		log.Crit("LLR state reading failed: genesis not applied")
	}
	s.cache.LlrState.Store(v)
	return *v
}

// FlushLlrState stores the LLR state in DB
func (s *Store) FlushLlrState() {
	s.rlp.Set(s.table.LlrState, []byte{}, s.GetLlrState())
}
