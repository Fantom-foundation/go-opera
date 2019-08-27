package inter

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

type EventHeaderData struct {
	Version uint32

	Epoch idx.SuperFrame
	Seq   idx.Event

	Frame  idx.Frame
	IsRoot bool

	Creator common.Address

	PrevEpochHash common.Hash
	Parents       hash.Events

	GasLeft uint64
	GasUsed uint64

	Lamport     idx.Lamport
	ClaimedTime Timestamp
	MedianTime  Timestamp

	TxHash common.Hash

	Extra []byte

	hash *hash.Event `rlp:"-"` // cache for .Hash()
}

type EventHeader struct {
	EventHeaderData

	Sig []byte
}

type Event struct {
	EventHeader
	Transactions types.Transactions
}

func (e *EventHeaderData) HashToSign() common.Hash {
	hasher := sha3.New256()
	err := rlp.Encode(hasher, []interface{}{
		"Fantom signed event header",
		e,
	})
	if err != nil {
		panic("can't encode: " + err.Error())
	}
	return hash.FromBytes(hasher.Sum(nil))
}

func (e *EventHeaderData) SelfParent() *hash.Event {
	if e.Seq <= 1 || len(e.Parents) == 0 {
		return nil
	}
	return &e.Parents[0]
}

func (e *EventHeaderData) IsSelfParent(hash hash.Event) bool {
	if e.SelfParent() == nil {
		return false
	}
	return *e.SelfParent() == hash
}

// SignBy signs event by private key.
func (e *Event) SignBy(priv *ecdsa.PrivateKey) error {
	sig, err := crypto.Sign(e.HashToSign().Bytes(), priv)
	if err != nil {
		return err
	}

	e.Sig = sig
	return nil
}

// Verify sign event by public key.
func (e *Event) VerifySignature() bool {
	pk, err := crypto.SigToPub(e.HashToSign().Bytes(), e.Sig)
	if err != nil {
		return false
	}
	return crypto.PubkeyToAddress(*pk) == e.Creator
}

// Hash calcs hash of event (not cached).
func (e *EventHeaderData) CalcHash() hash.Event {
	hasher := sha3.New256()
	err := rlp.Encode(hasher, e)
	if err != nil {
		panic("can't encode: " + err.Error())
	}
	// return 24 bytes hash | epoch | lamport
	id := hash.BytesToEvent(hasher.Sum(nil))
	copy(id[0:4], e.Epoch.Bytes())
	copy(id[4:8], e.Lamport.Bytes())
	return id
}

func (e *EventHeaderData) RecacheHash() {
	id := e.CalcHash()
	e.hash = &id // TODO must be atomic
}

// Hash calcs hash of event (cached).
func (e *EventHeaderData) Hash() hash.Event {
	if e.hash == nil { // TODO must be atomic
		e.RecacheHash()
	}
	return *e.hash
}

// constructs empty event
func NewEvent() *Event {
	return &Event{
		EventHeader: EventHeader{
			EventHeaderData: EventHeaderData{
				Extra: []byte{},
			},
			Sig: []byte{},
		},
		Transactions: types.Transactions{},
	}
}

// String returns string representation.
func (e *Event) String() string {
	return fmt.Sprintf("Event{%s, %s, t=%d}", e.Hash().String(), e.Parents.String(), e.Lamport)
}

/*
 * Utils:
 */

// FakeFuzzingEvents generates random independent events for test purpose.
func FakeFuzzingEvents() (res []*Event) {
	creators := []common.Address{
		{},
		hash.FakePeer(),
		hash.FakePeer(),
		hash.FakePeer(),
	}
	parents := []hash.Events{
		hash.FakeEvents(1),
		hash.FakeEvents(2),
		hash.FakeEvents(8),
	}
	i := 0
	for c := 0; c < len(creators); c++ {
		for p := 0; p < len(parents); p++ {
			e := NewEvent()
			e.Seq = idx.Event(p)
			e.Creator = creators[c]
			e.Parents = parents[p]
			e.Extra = []byte{}
			e.Sig = []byte{}

			res = append(res, e)
			i++
		}
	}
	return
}
