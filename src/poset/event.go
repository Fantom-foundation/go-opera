package poset

import (
	"crypto/ecdsa"
	"fmt"
	"reflect"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/peers"
	"github.com/golang/protobuf/proto"
)

/*******************************************************************************
InternalTransactions
*******************************************************************************/

// NewInternalTransaction constructor
func NewInternalTransaction(tType TransactionType, peer peers.Peer) InternalTransaction {
	return InternalTransaction{
		Type: tType,
		Peer: &peer,
	}
}

// ProtoMarshal marshal internal transaction to protobuff
func (t *InternalTransaction) ProtoMarshal() ([]byte, error) {
	var bf proto.Buffer
	bf.SetDeterministic(true)
	if err := bf.Marshal(t); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

// ProtoUnmarshal unmarshal protobuff to internal transaction
func (t *InternalTransaction) ProtoUnmarshal(data []byte) error {
	return proto.Unmarshal(data, t)
}

/*******************************************************************************
EventBody
*******************************************************************************/

// Equals equality check for internal transaction
func (t *InternalTransaction) Equals(that *InternalTransaction) bool {
	return t.Peer.Equals(that.Peer) &&
		t.Type == that.Type
}

// BytesEquals compares bytes for equality
func BytesEquals(this []byte, that []byte) bool {
	if len(this) != len(that) {
		return false
	}
	for i, v := range this {
		if v != that[i] {
			return false
		}
	}
	return true
}

// InternalTransactionListEquals list equality check
func InternalTransactionListEquals(this []*InternalTransaction, that []*InternalTransaction) bool {
	if len(this) != len(that) {
		return false
	}
	for i, v := range this {
		if !v.Equals(that[i]) {
			return false
		}
	}
	return true
}

// BlockSignatureListEquals block signature list equality check
func BlockSignatureListEquals(this []*BlockSignature, that []*BlockSignature) bool {
	if len(this) != len(that) {
		return false
	}
	for i, v := range this {
		if !v.Equals(that[i]) {
			return false
		}
	}
	return true
}

// Equals event body equality check
func (e *EventBody) Equals(that *EventBody) bool {
	return reflect.DeepEqual(e.Transactions, that.Transactions) &&
		InternalTransactionListEquals(e.InternalTransactions, that.InternalTransactions) &&
		reflect.DeepEqual(e.Parents, that.Parents) &&
		reflect.DeepEqual(e.Creator, that.Creator) &&
		e.Index == that.Index &&
		BlockSignatureListEquals(e.BlockSignatures, that.BlockSignatures)
}

// ProtoMarshal marshal event body to protobuff
func (e *EventBody) ProtoMarshal() ([]byte, error) {
	var bf proto.Buffer
	bf.SetDeterministic(true)
	if err := bf.Marshal(e); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

// ProtoUnmarshal unmarshal protobuff to event body
func (e *EventBody) ProtoUnmarshal(data []byte) error {
	return proto.Unmarshal(data, e)
}

// Hash returns hash of event body
func (e *EventBody) Hash() ([]byte, error) {
	hashBytes, err := e.ProtoMarshal()
	if err != nil {
		return nil, err
	}
	return crypto.SHA256(hashBytes), nil
}

/*******************************************************************************
Event
*******************************************************************************/

// LamportTimestampNIL nil value for lamport
const LamportTimestampNIL int64 = -1
// RoundNIL nil value for round
const RoundNIL int64 = -1

// ToEvent converts message to event
func (m *EventMessage) ToEvent() Event {
	return Event{
		Message:          m,
		lamportTimestamp: LamportTimestampNIL,
		round:            RoundNIL,
		roundReceived:    RoundNIL,
	}
}

// Equals compares equality of two event messages
func (m *EventMessage) Equals(that *EventMessage) bool {
	return m.Body.Equals(that.Body) &&
		m.Signature == that.Signature &&
		BytesEquals(m.FlagTable, that.FlagTable) &&
		reflect.DeepEqual(m.ClothoProof, that.ClothoProof)
}

// NewEvent creates new block event.
func NewEvent(transactions [][]byte,
	internalTransactions []InternalTransaction,
	blockSignatures []BlockSignature,
	parents []string, creator []byte, index int64,
	flagTable map[string]int64) Event {

	internalTransactionPointers := make([]*InternalTransaction, len(internalTransactions))
	for i, v := range internalTransactions {
		internalTransactionPointers[i] = new(InternalTransaction)
		*internalTransactionPointers[i] = v
	}
	blockSignaturePointers := make([]*BlockSignature, len(blockSignatures))
	for i, v := range blockSignatures {
		blockSignaturePointers[i] = new(BlockSignature)
		*blockSignaturePointers[i] = v
	}
	body := EventBody{
		Transactions:         transactions,
		InternalTransactions: internalTransactionPointers,
		BlockSignatures:      blockSignaturePointers,
		Parents:              parents,
		Creator:              creator,
		Index:                index,
	}

	ft, _ := proto.Marshal(&FlagTableWrapper{Body: flagTable})

	return Event{
		Message: &EventMessage{
			Body:      &body,
			FlagTable: ft,
		},
		lamportTimestamp: LamportTimestampNIL,
		round:            RoundNIL,
		roundReceived:    RoundNIL,
	}
}

// Event struct
type Event struct {
	Message          *EventMessage
	lamportTimestamp int64
	round            int64
	roundReceived    int64
}

// GetRound Round returns round of event.
func (e *Event) GetRound() int64 {
	if e.round < 0 {
		return RoundNIL
	}
	return e.round
}

// GetRoundReceived Round returns round in which the event is received.
func (e *Event) GetRoundReceived() int64 {
	if e.roundReceived < 0 {
		return RoundNIL
	}
	return e.round
}

// GetLamportTimestamp returns the lamport timestamp
func (e *Event) GetLamportTimestamp() int64 {
	if e.lamportTimestamp < 0 {
		return LamportTimestampNIL
	}
	return e.lamportTimestamp
}

// GetCreator returns the creator for the event
func (e *Event) GetCreator() string {
	return fmt.Sprintf("0x%X", e.Message.Body.Creator)
}

// SelfParent returns the previous event block in this creator DAG
func (e *Event) SelfParent() string {
	return e.Message.Body.Parents[0]
}

// OtherParent returns the other (not this node) parent(s)
func (e *Event) OtherParent() string {
	return e.Message.Body.Parents[1]
}

// Transactions returns all transactions in the event
func (e *Event) Transactions() [][]byte {
	return e.Message.Body.Transactions
}

// Index returns the index (heigh) of this event
func (e *Event) Index() int64 {
	return e.Message.Body.Index
}

// BlockSignatures returns all block signatures for this event
func (e *Event) BlockSignatures() []*BlockSignature {
	return e.Message.Body.BlockSignatures
}

// IsLoaded True if Event contains a payload or is the initial Event of its creator
func (e *Event) IsLoaded() bool {
	if e.Message.Body.Index == 0 {
		return true
	}

	hasTransactions := e.Message.Body.Transactions != nil &&
		(len(e.Message.Body.Transactions) > 0 || len(e.Message.Body.InternalTransactions) > 0)

	return hasTransactions
}

// Sign ecdsa sig
func (e *Event) Sign(privKey *ecdsa.PrivateKey) error {
	signBytes, err := e.Message.Body.Hash()
	if err != nil {
		return err
	}
	R, S, err := crypto.Sign(privKey, signBytes)
	if err != nil {
		return err
	}
	e.Message.Signature = crypto.EncodeSignature(R, S)
	return err
}

// Verify ecdsa sig
func (e *Event) Verify() (bool, error) {
	pubBytes := e.Message.Body.Creator
	pubKey := crypto.ToECDSAPub(pubBytes)

	signBytes, err := e.Message.Body.Hash()
	if err != nil {
		return false, err
	}

	r, s, err := crypto.DecodeSignature(e.Message.Signature)
	if err != nil {
		return false, err
	}

	return crypto.Verify(pubKey, signBytes, r, s), nil
}

// ProtoMarshal event to protobuff
func (e *Event) ProtoMarshal() ([]byte, error) {
	var bf proto.Buffer
	bf.SetDeterministic(true)
	if err := bf.Marshal(e.Message); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

// ProtoUnmarshal profotbuff to event
func (e *Event) ProtoUnmarshal(data []byte) error {
	e.Message = &EventMessage{}
	return proto.Unmarshal(data, e.Message)
}

// Hash sha256 hash of body
func (e *Event) Hash() ([]byte, error) {
	if len(e.Message.Hash) == 0 {
		hash, err := e.Message.Body.Hash()
		if err != nil {
			return nil, err
		}
		e.Message.Hash = hash
	}
	return e.Message.Hash, nil
}

// Hex of hash
func (e *Event) Hex() string {
	hash, _ := e.Hash()
	return fmt.Sprintf("0x%X", hash)
}

// SetRound for event
func (e *Event) SetRound(r int64) {
	e.round = r
}

// SetLamportTimestamp for event
func (e *Event) SetLamportTimestamp(t int64) {
	e.lamportTimestamp = t
}

// SetRoundReceived for event
func (e *Event) SetRoundReceived(rr int64) {
	e.roundReceived = rr
}

// SetWireInfo for event
func (e *Event) SetWireInfo(selfParentIndex,
	otherParentCreatorID,
	otherParentIndex,
	creatorID int64) {
	e.Message.SelfParentIndex = selfParentIndex
	e.Message.OtherParentCreatorID = otherParentCreatorID
	e.Message.OtherParentIndex = otherParentIndex
	e.Message.CreatorID = creatorID
}

// WireBlockSignatures returns the wire block signatures for the event
func (e *Event) WireBlockSignatures() []WireBlockSignature {
	if e.Message.Body.BlockSignatures != nil {
		wireSignatures := make([]WireBlockSignature, len(e.Message.Body.BlockSignatures))
		for i, bs := range e.Message.Body.BlockSignatures {
			wireSignatures[i] = bs.ToWire()
		}

		return wireSignatures
	}
	return nil
}

// ToWire converts event to wire event
func (e *Event) ToWire() WireEvent {

	transactions := make([]InternalTransaction, len(e.Message.Body.InternalTransactions))
	for i, v := range e.Message.Body.InternalTransactions {
		transactions[i] = *v
	}
	return WireEvent{
		Body: WireBody{
			Transactions:         e.Message.Body.Transactions,
			InternalTransactions: transactions,
			SelfParentIndex:      e.Message.SelfParentIndex,
			OtherParentCreatorID: e.Message.OtherParentCreatorID,
			OtherParentIndex:     e.Message.OtherParentIndex,
			CreatorID:            e.Message.CreatorID,
			Index:                e.Message.Body.Index,
			BlockSignatures:      e.WireBlockSignatures(),
		},
		Signature:   e.Message.Signature,
		FlagTable:   e.Message.FlagTable,
		ClothoProof: e.Message.ClothoProof,
	}
}

// ReplaceFlagTable replaces flag table.
func (e *Event) ReplaceFlagTable(flagTable map[string]int64) (err error) {
	e.Message.FlagTable, err = proto.Marshal(&FlagTableWrapper{Body: flagTable})
	return err
}

// GetFlagTable returns the flag table.
func (e *Event) GetFlagTable() (result map[string]int64, err error) {
	flagTable := new(FlagTableWrapper)
	err = proto.Unmarshal(e.Message.FlagTable, flagTable)
	return flagTable.Body, err
}

// MergeFlagTable returns merged flag table object.
func (e *Event) MergeFlagTable(
	dst map[string]int64) (result map[string]int64, err error) {
	src := new(FlagTableWrapper)
	if err := proto.Unmarshal(e.Message.FlagTable, src); err != nil {
		return nil, err
	}

	for id, flag := range dst {
		if src.Body[id] == 0 && flag == 1 {
			src.Body[id] = 1
		}
	}
	return src.Body, err
}

// CreatorID returns the creator ID for an event
func (e *Event) CreatorID() int64 {
	return e.Message.CreatorID
}

// OtherParentCreatorID ID of other parent(s)
func (e *Event) OtherParentCreatorID() int64 {
	return e.Message.OtherParentCreatorID
}

func rootSelfParent(participantID int64) string {
	return fmt.Sprintf("Root%d", participantID)
}

/*******************************************************************************
sorting
*******************************************************************************/

// ByTopologicalOrder implements sort.Interface for []Event based on
// the topologicalIndex field.
// THIS IS A PARTIAL ORDER
type ByTopologicalOrder []Event

func (a ByTopologicalOrder) Len() int      { return len(a) }
func (a ByTopologicalOrder) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByTopologicalOrder) Less(i, j int) bool {
	return a[i].Message.TopologicalIndex < a[j].Message.TopologicalIndex
}

// ByLamportTimestamp implements sort.Interface for []Event based on
// the lamportTimestamp field.
// THIS IS A TOTAL ORDER
type ByLamportTimestamp []Event

func (a ByLamportTimestamp) Len() int      { return len(a) }
func (a ByLamportTimestamp) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByLamportTimestamp) Less(i, j int) bool {
	it := a[i].lamportTimestamp
	jt := a[j].lamportTimestamp
	if it != jt {
		return it < jt
	}

	wsi, _, _ := crypto.DecodeSignature(a[i].Message.Signature)
	wsj, _, _ := crypto.DecodeSignature(a[j].Message.Signature)
	return wsi.Cmp(wsj) < 0
}

/*******************************************************************************
 WireEvent
*******************************************************************************/

// WireBody struct
type WireBody struct {
	Transactions         [][]byte
	InternalTransactions []InternalTransaction
	BlockSignatures      []WireBlockSignature

	SelfParentIndex      int64
	OtherParentCreatorID int64
	OtherParentIndex     int64
	CreatorID            int64

	Index int64
}

// WireEvent struct
type WireEvent struct {
	Body        WireBody
	Signature   string
	FlagTable   []byte
	ClothoProof []string
}

// BlockSignatures TODO
func (we *WireEvent) BlockSignatures(validator []byte) []BlockSignature {
	if we.Body.BlockSignatures != nil {
		blockSignatures := make([]BlockSignature, len(we.Body.BlockSignatures))
		for k, bs := range we.Body.BlockSignatures {
			blockSignatures[k] = BlockSignature{
				Validator: validator,
				Index:     bs.Index,
				Signature: bs.Signature,
			}
		}
		return blockSignatures
	}
	return nil
}
