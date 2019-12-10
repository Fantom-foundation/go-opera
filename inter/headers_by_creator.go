package inter

import (
	"bytes"
	"io"
	"sort"

	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

type (
	// HeadersByCreator is a event headers grouped by creator.
	// ( creator --> event header )
	HeadersByCreator map[idx.StakerID]*EventHeaderData
)

type headersByCreatorPair struct {
	Creator idx.StakerID
	Header  *EventHeaderData
}

// EncodeRLP implements rlp.Encoder interface.
func (hh HeadersByCreator) EncodeRLP(w io.Writer) error {
	arr := make([]headersByCreatorPair, 0, len(hh))
	creators := make([]idx.StakerID, 0, len(hh))
	for creator := range hh {
		creators = append(creators, creator)
	}
	// for determinism
	sort.Slice(creators, func(i, j int) bool {
		a, b := creators[i], creators[j]
		return bytes.Compare(a.Bytes(), b.Bytes()) < 0
	})

	for _, creator := range creators {
		header := hh[creator]
		arr = append(arr, headersByCreatorPair{
			Creator: creator,
			Header:  header,
		})
	}
	return rlp.Encode(w, arr)
}

// DecodeRLP is for RLP deserialization.
func (hh *HeadersByCreator) DecodeRLP(s *rlp.Stream) error {
	if *hh == nil {
		*hh = HeadersByCreator{}
	}
	var arr []headersByCreatorPair
	if err := s.Decode(&arr); err != nil {
		return err
	}

	for _, w := range arr {
		(*hh)[w.Creator] = w.Header
	}

	return nil
}

// Bytes gets the byte representation of the headers map.
func (hh HeadersByCreator) Bytes() []byte {
	b, _ := rlp.EncodeToBytes(hh)
	return b
}
