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
func NewInternalTransaction(tType TransactionType, peer peers.Peer) InternalTransaction {
	return InternalTransaction{
		Type: tType,
		Peer: &peer,
	}
}

func (m *InternalTransaction) ProtoMarshal() ([]byte, error) {
	var bf proto.Buffer
	bf.SetDeterministic(true)
	if err := bf.Marshal(m); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}
func (m *InternalTransaction) ProtoUnmarshal(data []byte) error {
	return proto.Unmarshal(data, m)
}

/*******************************************************************************
EventBody
*******************************************************************************/

func (m *InternalTransaction) Equals(that *InternalTransaction) bool {
	return m.Peer.Equals(that.Peer) &&
		m.Type == that.Type
}

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

func (m *EventBody) Equals(that *EventBody) bool {
	return reflect.DeepEqual(m.Transactions, that.Transactions) &&
		InternalTransactionListEquals(m.InternalTransactions, that.InternalTransactions) &&
		reflect.DeepEqual(m.Parents, that.Parents) &&
		reflect.DeepEqual(m.Creator, that.Creator) &&
		m.Index == that.Index &&
		BlockSignatureListEquals(m.BlockSignatures, that.BlockSignatures)
}

func (m *EventBody) ProtoMarshal() ([]byte, error) {
	var bf proto.Buffer
	bf.SetDeterministic(true)
	if err := bf.Marshal(m); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

func (m *EventBody) ProtoUnmarshal(data []byte) error {
	return proto.Unmarshal(data, m)
}

func (m *EventBody) Hash() ([]byte, error) {
	hashBytes, err := m.ProtoMarshal()
	if err != nil {
		return nil, err
	}
	return crypto.SHA256(hashBytes), nil
}

/*******************************************************************************
Event
*******************************************************************************/

const LamportTimestampNIL int64 = -1
const RoundNIL int64 = -1

func (m *EventMessage) ToEvent() Event {
	return Event{
		Message:          m,
		lamportTimestamp: LamportTimestampNIL,
		round:            RoundNIL,
		roundReceived:    RoundNIL,
	}
}

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

type Event struct {
	Message          *EventMessage
	lamportTimestamp int64
	round            int64
	roundReceived    int64
}

// Round returns round of event.
func (m *Event) GetRound() int64 {
	if m.round < 0 {
		return RoundNIL
	}
	return m.round
}

// Round returns round in which the event is received.
func (m *Event) GetRoundReceived() int64 {
	if m.roundReceived < 0 {
		return RoundNIL
	}
	return m.round
}

func (e *Event) GetLamportTimestamp() int64 {
	if e.lamportTimestamp < 0 {
		return LamportTimestampNIL
	}
	return e.lamportTimestamp
}

func (e *Event) GetCreator() string {
	return fmt.Sprintf("0x%X", e.Message.Body.Creator)
}

func (e *Event) SelfParent() string {
	return e.Message.Body.Parents[0]
}

func (e *Event) OtherParent() string {
	return e.Message.Body.Parents[1]
}

func (e *Event) Transactions() [][]byte {
	return e.Message.Body.Transactions
}

func (e *Event) Index() int64 {
	return e.Message.Body.Index
}

func (e *Event) BlockSignatures() []*BlockSignature {
	return e.Message.Body.BlockSignatures
}

//True if Event contains a payload or is the initial Event of its creator
func (e *Event) IsLoaded() bool {
	if e.Message.Body.Index == 0 {
		return true
	}

	hasTransactions := e.Message.Body.Transactions != nil &&
		(len(e.Message.Body.Transactions) > 0 || len(e.Message.Body.InternalTransactions) > 0)

	return hasTransactions
}

//ecdsa sig
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

func (e *Event) ProtoMarshal() ([]byte, error) {
	var bf proto.Buffer
	bf.SetDeterministic(true)
	if err := bf.Marshal(e.Message); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

func (e *Event) ProtoUnmarshal(data []byte) error {
	e.Message = &EventMessage{}
	return proto.Unmarshal(data, e.Message)
}

//sha256 hash of body
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

func (e *Event) Hex() string {
	hash, _ := e.Hash()
	return fmt.Sprintf("0x%X", hash)
}

func (e *Event) SetRound(r int64) {
	e.round = r
}

func (e *Event) SetLamportTimestamp(t int64) {
	e.lamportTimestamp = t
}

func (e *Event) SetRoundReceived(rr int64) {
	e.roundReceived = rr
}

func (e *Event) SetWireInfo(selfParentIndex,
	otherParentCreatorID,
	otherParentIndex,
	creatorID int64) {
	e.Message.SelfParentIndex = selfParentIndex
	e.Message.OtherParentCreatorID = otherParentCreatorID
	e.Message.OtherParentIndex = otherParentIndex
	e.Message.CreatorID = creatorID
}

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

func (e *Event) CreatorID() int64 {
	return e.Message.CreatorID
}

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

type WireEvent struct {
	Body        WireBody
	Signature   string
	FlagTable   []byte
	ClothoProof []string
}

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
