package genesisstore

import (
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-opera/inter/ibr"
	"github.com/Fantom-foundation/go-opera/inter/ier"
	"github.com/Fantom-foundation/go-opera/opera/genesis"
	"github.com/Fantom-foundation/go-opera/utils/iodb"
)

type (
	Blocks struct {
		fMap FilesMap
	}
	Epochs struct {
		fMap FilesMap
	}
	RawEvmItems struct {
		fMap FilesMap
	}
)

func (s *Store) Genesis() genesis.Genesis {
	return genesis.Genesis{
		Header:      s.head,
		Blocks:      s.Blocks(),
		Epochs:      s.Epochs(),
		RawEvmItems: s.RawEvmItems(),
	}
}

func getSectionName(base string, i int) string {
	if i == 0 {
		return base
	}
	return fmt.Sprintf("%s-%d", base, i)
}

func (s Store) Header() genesis.Header {
	return s.head
}

func (s *Store) Blocks() genesis.Blocks {
	return Blocks{s.fMap}
}

func (s Blocks) ForEach(fn func(ibr.LlrIdxFullBlockRecord) bool) {
	for i := 1000; i >= 0; i-- {
		f, err := s.fMap(BlocksSection(i))
		if err != nil {
			continue
		}
		stream := rlp.NewStream(f, 0)
		for {
			br := ibr.LlrIdxFullBlockRecord{}
			err = stream.Decode(&br)
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Crit("Failed to decode Blocks genesis section", "err", err)
			}
			if !fn(br) {
				break
			}
		}
	}
}

func (s *Store) Epochs() genesis.Epochs {
	return Epochs{s.fMap}
}

func (s Epochs) ForEach(fn func(ier.LlrIdxFullEpochRecord) bool) {
	for i := 1000; i >= 0; i-- {
		f, err := s.fMap(EpochsSection(i))
		if err != nil {
			continue
		}
		stream := rlp.NewStream(f, 0)
		for {
			er := ier.LlrIdxFullEpochRecord{}
			err = stream.Decode(&er)
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Crit("Failed to decode Epochs genesis section", "err", err)
			}
			if !fn(er) {
				break
			}
		}
	}
}

func (s *Store) RawEvmItems() genesis.EvmItems {
	return RawEvmItems{s.fMap}
}

func (s RawEvmItems) ForEach(fn func(key, value []byte) bool) {
	for i := 1000; i >= 0; i-- {
		f, err := s.fMap(EvmSection(i))
		if err != nil {
			continue
		}
		it := iodb.NewIterator(f)
		for it.Next() {
			if !fn(it.Key(), it.Value()) {
				break
			}
		}
		if it.Error() != nil {
			log.Crit("Failed to decode RawEvmItems genesis section", "err", it.Error())
		}
	}
}
