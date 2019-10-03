package inter

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

type EventHeaderData struct {
	Version uint32

	Epoch idx.Epoch
	Seq   idx.Event

	Frame  idx.Frame
	IsRoot bool

	Creator common.Address

	PrevEpochHash common.Hash
	Parents       hash.Events

	GasPowerLeft uint64
	GasPowerUsed uint64

	Lamport     idx.Lamport
	ClaimedTime Timestamp
	MedianTime  Timestamp

	TxHash common.Hash

	Extra []byte

	// caches
	hash atomic.Value
}

type EventHeader struct {
	EventHeaderData

	Sig []byte
}

type Event struct {
	EventHeader
	Transactions types.Transactions

	// caches
	size atomic.Value
}

// NewEvent constructs empty event
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
func (e *EventHeaderData) String() string {
	if e.IsRoot {
		return fmt.Sprintf("{id=%s, p=%s, seq=%d, f=%d, root}", e.Hash().String(), e.Parents.String(), e.Seq, e.Frame)
	}
	return fmt.Sprintf("{id=%s, p=%s, seq=%d, f=%d}", e.Hash().String(), e.Parents.String(), e.Seq, e.Frame)
}

func (e *EventHeaderData) DataToSign() []byte {
	buf := bytes.NewBuffer([]byte{})
	buf.Write([]byte("Lachesis: I'm signing the Event"))
	buf.Write(e.Hash().Bytes())
	return buf.Bytes()
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
	signer := func(data []byte) ([]byte, error) {
		data = crypto.Keccak256(data)
		sig, err := crypto.Sign(data, priv)
		return sig, err
	}

	return e.Sign(signer)
}

// Sign event by signer.
func (e *Event) Sign(signer func([]byte) ([]byte, error)) error {
	e.RecacheHash() // because HashToSign uses .Hash
	sig, err := signer(e.DataToSign())
	if err != nil {
		return err
	}

	e.Sig = sig
	return nil
}

// VerifySignature checks the signature against e.Creator.
func (e *Event) VerifySignature() bool {
	// NOTE: Keccak256 because of AccountManager
	signedHash := crypto.Keccak256(e.DataToSign())
	pk, err := crypto.SigToPub(signedHash, e.Sig)
	if err != nil {
		return false
	}
	return crypto.PubkeyToAddress(*pk) == e.Creator
}

/*
 * Event ID (hash):
 */

// CalcHash calcs hash of event (not cached).
func (e *EventHeaderData) CalcHash() hash.Event {
	hasher := sha3.NewLegacyKeccak256()
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

func (e *EventHeaderData) RecacheHash() hash.Event {
	id := e.CalcHash()
	e.hash.Store(id)
	return id
}

// Hash calcs hash of event (cached).
func (e *EventHeaderData) Hash() hash.Event {
	if cached := e.hash.Load(); cached != nil {
		return cached.(hash.Event)
	}
	return e.RecacheHash()
}

/*
 * Event size:
 */

type writeCounter int

func (c *writeCounter) Write(b []byte) (int, error) {
	*c += writeCounter(len(b))
	return len(b), nil
}

func (e *Event) CalcSize() int {
	c := writeCounter(0)
	_ = rlp.Encode(&c, e)
	return int(c)
}

func (e *Event) RecacheSize() int {
	size := e.CalcSize()
	e.size.Store(size)
	return size
}

func (e *Event) Size() int {
	if cached := e.size.Load(); cached != nil {
		return cached.(int)
	}
	return e.RecacheSize()
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
