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

var (
	EmptyTxHash = types.DeriveSha(types.Transactions{})
)

// EventHeaderData is the graph vertex in the Lachesis consensus algorithm
// Doesn't contain transactions, only their hash
// Doesn't contain event signature
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

// EventHeader is the graph vertex in the Lachesis consensus algorithm
// Doesn't contain transactions, only their hash
type EventHeader struct {
	EventHeaderData

	Sig []byte
}

// Event is the graph vertex in the Lachesis consensus algorithm
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

func (e *EventHeaderData) NoTransactions() bool {
	return e.TxHash == EmptyTxHash
}

// DataToSign returns data which must be signed to sign the event
func (e *EventHeaderData) DataToSign() []byte {
	buf := bytes.NewBuffer([]byte{})
	buf.Write([]byte("Lachesis: I'm signing the Event"))
	buf.Write(e.Hash().Bytes())
	return buf.Bytes()
}

// SelfParent returns event's self-parent, if any
func (e *EventHeaderData) SelfParent() *hash.Event {
	if e.Seq <= 1 || len(e.Parents) == 0 {
		return nil
	}
	return &e.Parents[0]
}

// IsSelfParent is true if specified ID is event's self-parent
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

// CalcHash re-calculates event's ID
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

// RecacheHash re-calculates event's ID and caches it
func (e *EventHeaderData) RecacheHash() hash.Event {
	id := e.CalcHash()
	e.hash.Store(id)
	return id
}

// Hash returns cached event ID
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

// Write only counts "written" bytes
func (c *writeCounter) Write(b []byte) (int, error) {
	*c += writeCounter(len(b))
	return len(b), nil
}

// CalcSize re-calculates event's size
func (e *Event) CalcSize() int {
	c := writeCounter(0)
	_ = rlp.Encode(&c, e)
	return int(c)
}

// RecacheSize re-calculates event's size and caches it
func (e *Event) RecacheSize() int {
	size := e.CalcSize()
	e.size.Store(size)
	return size
}

// Size returns cached event size
func (e *Event) Size() int {
	if cached := e.size.Load(); cached != nil {
		return cached.(int)
	}
	return e.RecacheSize()
}

/*
 * Utils:
 */

// FakeFuzzingEvents generates random independent events with the same epoch for testing purpose.
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
			e.Epoch = hash.FakeEpoch()
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
