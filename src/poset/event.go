package poset

import (
	"crypto/ecdsa"
	"fmt"
	"reflect"

	"github.com/andrecronje/lachesis/src/crypto"
	"github.com/andrecronje/lachesis/src/peers"
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

func (t *InternalTransaction) ProtoMarshal() ([]byte, error) {
	var bf proto.Buffer
	bf.SetDeterministic(true)
	if err := bf.Marshal(t); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}
func (t *InternalTransaction) ProtoUnmarshal(data []byte) error {
	return proto.Unmarshal(data, t)
}

/*******************************************************************************
EventBody
*******************************************************************************/

func (this *InternalTransaction) Equals(that *InternalTransaction) bool {
	return this.Peer.Equals(that.Peer) &&
		this.Type == that.Type
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

func (this *BlockSignature) Equals(that *BlockSignature) bool {
	return reflect.DeepEqual(this.Validator, that.Validator) &&
		this.Index == that.Index &&
		this.Signature == that.Signature
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

func (this *EventBody) Equals(that *EventBody) bool {
	return reflect.DeepEqual(this.Transactions, that.Transactions) &&
		InternalTransactionListEquals(this.InternalTransactions, that.InternalTransactions) &&
		reflect.DeepEqual(this.Parents, that.Parents) &&
		reflect.DeepEqual(this.Creator, that.Creator) &&
		this.Index == that.Index &&
		BlockSignatureListEquals(this.BlockSignatures, that.BlockSignatures)
}

func (e *EventBody) ProtoMarshal() ([]byte, error) {
	var bf proto.Buffer
	bf.SetDeterministic(true)
	if err := bf.Marshal(e); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

func (e *EventBody) ProtoUnmarshal(data []byte) error {
	return proto.Unmarshal(data, e)
}

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

type Event struct {
	Message EventMessage

	//used for sorting
	round           		*int64
	lamportTimestamp		*int64
	roundReceived			*int64
	selfParentIndex			int64
	otherParentCreatorID	int64
	otherParentIndex		int64
	creatorID				int64
	topologicalIndex		int64
	creator					string
	hash					[]byte
	hex						string
}

func (e EventMessage) ToEvent() Event {
	return Event {
		Message: e,
	}
}

func (this *EventMessage) Equals(that *EventMessage) bool {
	return this.Body.Equals(that.Body) &&
		this.Signature == that.Signature &&
		BytesEquals(this.FlagTable, that.FlagTable) &&
		reflect.DeepEqual(this.WitnessProof, that.WitnessProof)
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

	ft, _ := proto.Marshal(&FlagTableWrapper { Body: flagTable })

	return Event{
		Message: EventMessage {
			Body:      &body,
			FlagTable: ft,
		},
	}
}

// Round returns round of event.
func (e *Event) Round() int64 {
	if e.round == nil || *e.round < 0 {
		return -1
	}
	return *e.round
}

func (e *Event) Creator() string {
	if e.creator == "" {
		e.creator = fmt.Sprintf("0x%X", e.Message.Body.Creator)
	}
	return e.creator
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
	if err := bf.Marshal(&e.Message); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

func (e *Event) ProtoUnmarshal(data []byte) error {
	return proto.Unmarshal(data, &e.Message)
}

//sha256 hash of body
func (e *Event) Hash() ([]byte, error) {
	if len(e.hash) == 0 {
		hash, err := e.Message.Body.Hash()
		if err != nil {
			return nil, err
		}
		e.hash = hash
	}
	return e.hash, nil
}

func (e *Event) Hex() string {
	if e.hex == "" {
		hash, _ := e.Hash()
		e.hex = fmt.Sprintf("0x%X", hash)
	}
	return e.hex
}

func (e *Event) SetRound(r int64) {
	if e.round == nil {
		e.round = new(int64)
	}
	*e.round = r
}

func (e *Event) SetLamportTimestamp(t int64) {
	if e.lamportTimestamp == nil {
		e.lamportTimestamp = new(int64)
	}
	*e.lamportTimestamp = t
}

func (e *Event) SetRoundReceived(rr int64) {
	if e.roundReceived == nil {
		e.roundReceived = new(int64)
	}
	*e.roundReceived = rr
}

func (e *Event) SetWireInfo(selfParentIndex,
	otherParentCreatorID,
	otherParentIndex,
	creatorID int64) {
	e.selfParentIndex = selfParentIndex
	e.otherParentCreatorID = otherParentCreatorID
	e.otherParentIndex = otherParentIndex
	e.creatorID = creatorID
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
			SelfParentIndex:      e.selfParentIndex,
			OtherParentCreatorID: e.otherParentCreatorID,
			OtherParentIndex:     e.otherParentIndex,
			CreatorID:            e.creatorID,
			Index:                e.Message.Body.Index,
			BlockSignatures:      e.WireBlockSignatures(),
		},
		Signature:    e.Message.Signature,
		FlagTable:    e.Message.FlagTable,
		WitnessProof: e.Message.WitnessProof,
	}
}

// ReplaceFlagTable replaces flag table.
func (e *Event) ReplaceFlagTable(flagTable map[string]int64) (err error) {
	e.Message.FlagTable, err = proto.Marshal(&FlagTableWrapper { Body: flagTable })
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
	return e.creatorID
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
	return a[i].topologicalIndex < a[j].topologicalIndex
}

// ByLamportTimestamp implements sort.Interface for []Event based on
// the lamportTimestamp field.
// THIS IS A TOTAL ORDER
type ByLamportTimestamp []Event

func (a ByLamportTimestamp) Len() int      { return len(a) }
func (a ByLamportTimestamp) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByLamportTimestamp) Less(i, j int) bool {
	it, jt := int64(-1), int64(-1)
	if a[i].lamportTimestamp != nil {
		it = *a[i].lamportTimestamp
	}
	if a[j].lamportTimestamp != nil {
		jt = *a[j].lamportTimestamp
	}
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
	Body         WireBody
	Signature    string
	FlagTable    []byte
	WitnessProof []string
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
