package inter

import (
	"crypto/sha256"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"
)

type EventI interface {
	dag.Event
	CreationTime() Timestamp
	MedianTime() Timestamp
	PrevEpochHash() *hash.Hash
	Extra() []byte
	TxHash() hash.Hash
	NoTxs() bool
	GasPowerLeft() GasPowerLeft
	GasPowerUsed() uint64
	HashToSign() hash.Hash
}

type EventPayloadI interface {
	EventI
	Sig() Signature
	Txs() types.Transactions
}

var (
	// EmptyTxHash is hash of empty transactions list. Used to check that event doesn't have transactions not having full event.
	EmptyTxHash = hash.Hash(types.DeriveSha(types.Transactions{}, new(trie.Trie)))
)

type baseEvent struct {
	dag.BaseEvent
}

type mutableBaseEvent struct {
	dag.MutableBaseEvent
}

type extEventData struct {
	creationTime  Timestamp
	medianTime    Timestamp
	prevEpochHash *hash.Hash
	txHash        hash.Hash
	gasPowerLeft  GasPowerLeft
	gasPowerUsed  uint64
	extra         []byte
}

type sigData struct {
	sig Signature
}

type payloadData struct {
	txs types.Transactions
}

type Event struct {
	baseEvent
	extEventData

	// cache
	_hash *hash.Hash
}

type SignedEvent struct {
	Event
	sigData
}

type EventPayload struct {
	SignedEvent
	payloadData

	// cache
	_size int
}

type MutableEventPayload struct {
	mutableBaseEvent
	extEventData
	sigData
	payloadData
}

func (e *Event) HashToSign() hash.Hash {
	return *e._hash
}

func (e *EventPayload) Size() int {
	return e._size
}

func (e *extEventData) CreationTime() Timestamp { return e.creationTime }

func (e *extEventData) MedianTime() Timestamp { return e.medianTime }

func (e *extEventData) PrevEpochHash() *hash.Hash { return e.prevEpochHash }

func (e *extEventData) Extra() []byte { return e.extra }

func (e *extEventData) TxHash() hash.Hash { return e.txHash }

func (e *extEventData) NoTxs() bool { return e.txHash == EmptyTxHash }

func (e *extEventData) GasPowerLeft() GasPowerLeft { return e.gasPowerLeft }

func (e *extEventData) GasPowerUsed() uint64 { return e.gasPowerUsed }

func (e *sigData) Sig() Signature { return e.sig }

func (e *payloadData) Txs() types.Transactions { return e.txs }

func (e *MutableEventPayload) SetCreationTime(v Timestamp) { e.creationTime = v }

func (e *MutableEventPayload) SetMedianTime(v Timestamp) { e.medianTime = v }

func (e *MutableEventPayload) SetPrevEpochHash(v *hash.Hash) { e.prevEpochHash = v }

func (e *MutableEventPayload) SetExtra(v []byte) { e.extra = v }

func (e *MutableEventPayload) SetTxHash(v hash.Hash) { e.txHash = v }

func (e *MutableEventPayload) SetGasPowerLeft(v GasPowerLeft) { e.gasPowerLeft = v }

func (e *MutableEventPayload) SetGasPowerUsed(v uint64) { e.gasPowerUsed = v }

func (e *MutableEventPayload) SetSig(v Signature) { e.sig = v }

func (e *MutableEventPayload) SetTxs(v types.Transactions) { e.txs = v }

func eventHash(eventSer []byte) hash.Hash {
	hasher := sha256.New()
	_, err := hasher.Write(eventSer)
	if err != nil {
		panic("can't hash: " + err.Error())
	}
	return hash.BytesToHash(hasher.Sum(nil))
}

func calcEventID(h hash.Hash) (id [24]byte) {
	copy(id[:], h[:24])
	return id
}

func (e *MutableEventPayload) hashToSign() hash.Hash {
	b, err := e.immutable().Event.MarshalBinary()
	if err != nil {
		panic("can't encode: " + err.Error())
	}
	return eventHash(b)
}

func (e *MutableEventPayload) size() int {
	b, err := e.immutable().MarshalBinary()
	if err != nil {
		panic("can't encode: " + err.Error())
	}
	return len(b)
}

func (e *MutableEventPayload) HashToSign() hash.Hash {
	return e.hashToSign()
}

func (e *MutableEventPayload) Size() int {
	return e.size()
}

func (me *MutableEventPayload) build(h hash.Hash, size int) *EventPayload {
	return &EventPayload{
		SignedEvent: SignedEvent{
			Event: Event{
				baseEvent:    baseEvent{*me.MutableBaseEvent.Build(calcEventID(h))},
				extEventData: me.extEventData,
				_hash:        &h,
			},
			sigData: me.sigData,
		},
		payloadData: me.payloadData,
		_size:       size,
	}
}

func (me *MutableEventPayload) immutable() *EventPayload {
	return me.build(hash.Hash{}, 0)
}

func (e *MutableEventPayload) Build() *EventPayload {
	if e.txs.Len() < 1 {
		e.txHash = EmptyTxHash
	}
	eventSer, _ := e.immutable().Event.MarshalBinary()
	h := eventHash(eventSer)
	payloadSer, _ := e.immutable().MarshalBinary()
	return e.build(h, len(payloadSer))
}
