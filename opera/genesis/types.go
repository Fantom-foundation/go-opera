package genesis

import (
	"github.com/Fantom-foundation/lachesis-base/hash"

	"github.com/Fantom-foundation/go-opera/inter/ibr"
	"github.com/Fantom-foundation/go-opera/inter/ier"
)

type (
	Hashes struct {
		Blocks      hash.Hashes
		Epochs      hash.Hashes
		RawEvmItems hash.Hashes
	}
	Header struct {
		GenesisID   hash.Hash
		NetworkID   uint64
		NetworkName string
	}
	Blocks interface {
		ForEach(fn func(ibr.LlrIdxFullBlockRecord) bool)
	}
	Epochs interface {
		ForEach(fn func(ier.LlrIdxFullEpochRecord) bool)
	}
	EvmItems interface {
		ForEach(fn func(key, value []byte) bool)
	}
	Genesis struct {
		Header

		Blocks      Blocks
		Epochs      Epochs
		RawEvmItems EvmItems
	}
)

func includes(base, hh hash.Hashes) bool {
	baseSet := base.Set()
	for _, h := range hh {
		if !baseSet.Contains(h) {
			return false
		}
	}
	return true
}

func (h Hashes) Includes(h2 Hashes) bool {
	if !includes(h.Epochs, h2.Epochs) {
		return false
	}
	if !includes(h.Blocks, h2.Blocks) {
		return false
	}
	if !includes(h.RawEvmItems, h2.RawEvmItems) {
		return false
	}
	return true
}

func (h Hashes) Equal(h2 Hashes) bool {
	return h.Includes(h2) && h2.Includes(h)
}

func add(base, hh hash.Hashes) hash.Hashes {
	baseSet := base.Set()
	for _, h := range hh {
		if !baseSet.Contains(h) {
			base.Add(h)
		}
	}
	return base
}

func (h *Hashes) Add(h2 Hashes) {
	h.Blocks = add(h.Blocks, h2.Blocks)
	h.Epochs = add(h.Epochs, h2.Epochs)
	h.RawEvmItems = add(h.RawEvmItems, h2.RawEvmItems)
}

func (h Header) Equal(h2 Header) bool {
	return h == h2
}
