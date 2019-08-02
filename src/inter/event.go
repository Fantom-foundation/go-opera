package inter

import (
	"fmt"
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/cryptoaddr"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/inter/wire"
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"
)

type EventHeaderData struct {
	Version uint32

	Epoch idx.SuperFrame
	Seq   idx.Event

	Frame  idx.Frame
	IsRoot bool

	Creator hash.Peer // TODO common.Address

	GenesisHash hash.Hash
	Parents     hash.Events

	GasLeft uint64
	GasUsed uint64

	Lamport     idx.Lamport
	ClaimedTime Timestamp
	MedianTime  Timestamp

	TxHash hash.Hash

	Extra []byte

	hash hash.Event `rlp:"-"` // cache for .Hash()
}

type EventHeader struct {
	EventHeaderData

	Sig []byte
}

func (e *EventHeaderData) HashToSign() hash.Hash {
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

func (e *EventHeaderData) SelfParentEqualTo(hash hash.Event) bool {
	if e.SelfParent() == nil {
		return false
	}
	return *e.SelfParent() == hash
}

type Event struct {
	EventHeader
	InternalTransactions []*InternalTransaction
	ExternalTransactions ExtTxns
}

// SignBy signs event by private key.
func (e *Event) SignBy(priv *crypto.PrivateKey) error {
	sig, err := priv.Sign(e.HashToSign().Bytes())
	if err != nil {
		return err
	}

	e.Sig = sig
	return nil
}

// Verify sign event by public key.
func (e *Event) VerifySignature() bool {
	return cryptoaddr.VerifySignature(e.Creator, e.HashToSign(), e.Sig)
}

// Hash calcs hash of event.
func (e *EventHeaderData) Hash() hash.Event {
	if e.hash.IsZero() {
		hasher := sha3.New256()
		err := rlp.Encode(hasher, e)
		if err != nil {
			panic("can't encode: " + err.Error())
		}
		// TODO return  epoch | lamport | 24 bytes hash
		e.hash = hash.BytesToEvent(hasher.Sum(nil))
	}
	return e.hash
}

// FindInternalTxn find transaction in event's internal transactions list.
// TODO: use map
func (e *Event) FindInternalTxn(idx hash.Transaction) *InternalTransaction {
	for _, txn := range e.InternalTransactions {
		if TransactionHashOf(e.Creator, txn.Nonce) == idx {
			return txn
		}
	}
	return nil
}

// String returns string representation.
func (e *Event) String() string {
	return fmt.Sprintf("Event{%s, %s, t=%d}", e.Hash().String(), e.Parents.String(), e.Lamport)
}

// TODO erase
// ToWire converts to proto.Message.
func (e *Event) ToWire() (*wire.Event, *wire.Event_ExtTxnsValue) {
	return nil, nil
}

// TODO erase
// WireToEvent converts from wire.
func WireToEvent(w *wire.Event) *Event {
	return nil
}

/*
 * Utils:
 */

// FakeFuzzingEvents generates random independent events for test purpose.
func FakeFuzzingEvents() (res []*Event) {
	creators := []hash.Peer{
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
	extTxns := [][][]byte{
		nil,
		[][]byte{
			[]byte("fake external transaction 1"),
			[]byte("fake external transaction 2"),
		},
	}
	i := 0
	for c := 0; c < len(creators); c++ {
		for p := 0; p < len(parents); p++ {
			e := &Event{
				EventHeader: EventHeader{
					EventHeaderData: EventHeaderData{
						Seq:     idx.Event(p),
						Creator: creators[c],
						Parents: parents[p],
					},
				},
				InternalTransactions: []*InternalTransaction{
					{
						Amount:   999,
						Receiver: creators[c],
					},
				},
				ExternalTransactions: ExtTxns{
					Value: extTxns[i%len(extTxns)],
				},
			}

			res = append(res, e)
			i++
		}
	}
	return
}
