package poset

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"reflect"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/golang/protobuf/proto"
)

// StateHash is the hash of the current state of transactions, if you have one
// node talking to an app, and another set of nodes talking to inmem, the
// stateHash will be different
// statehash should be ignored for validator checking

// ProtoMarshal json encoding of body only
func (bb *BlockBody) ProtoMarshal() ([]byte, error) {
	var bf proto.Buffer
	bf.SetDeterministic(true)
	if err := bf.Marshal(bb); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

// ProtoUnmarshal unmarshal the protobuff for the block body
func (bb *BlockBody) ProtoUnmarshal(data []byte) error {
	return proto.Unmarshal(data, bb)
}

// Hash returns the block body hash
func (bb *BlockBody) Hash() ([]byte, error) {
	hashBytes, err := bb.ProtoMarshal()
	if err != nil {
		return nil, err
	}
	return crypto.Keccak256(hashBytes), nil
}

// ------------------------------------------------------------------------------

// ValidatorHex returns the Hex ID of a validator for this block
func (bs *BlockSignature) ValidatorHex() string {
	return fmt.Sprintf("0x%X", bs.Validator)
}

// ProtoMarshal marshal the block signatures to protobuff
func (bs *BlockSignature) ProtoMarshal() ([]byte, error) {
	var bf proto.Buffer
	bf.SetDeterministic(true)
	if err := bf.Marshal(bs); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

// ProtoUnmarshal unmarshals the blocksignature from protobuff
func (bs *BlockSignature) ProtoUnmarshal(data []byte) error {
	return proto.Unmarshal(data, bs)
}

// ToWire converts block signatures to wire (transport)
func (bs *BlockSignature) ToWire() WireBlockSignature {
	return WireBlockSignature{
		Index:     bs.Index,
		Signature: bs.Signature,
	}
}

// Equals checks blocksignature equality
func (bs *BlockSignature) Equals(that *BlockSignature) bool {
	return reflect.DeepEqual(bs.Validator, that.Validator) &&
		bs.Index == that.Index &&
		bs.Signature == that.Signature
}

// ------------------------------------------------------------------------------

// NewBlockFromFrame creates a new block from the given frame
func NewBlockFromFrame(blockIndex int64, frame Frame) (Block, error) {
	frameHash, err := frame.Hash()
	if err != nil {
		return Block{}, err
	}
	var transactions [][]byte
	for _, e := range frame.Events {
		transactions = append(transactions, e.Body.Transactions...)
	}
	return NewBlock(blockIndex, frame.Round, frameHash, transactions), nil
}

// NewBlock creates a new empty block with current time
func NewBlock(blockIndex, roundReceived int64, frameHash []byte, txs [][]byte) Block {
	body := BlockBody{
		Index:         blockIndex,
		RoundReceived: roundReceived,
		Transactions:  txs,
	}
	return Block{
		Body:        &body,
		FrameHash:   frameHash,
		Signatures:  make(map[string]string),
		CreatedTime: time.Now().Unix(),
	}
}

// Index returns the index (height) of the block
func (b *Block) Index() int64 {
	return b.Body.Index
}

// Transactions returns the transactions in a block
func (b *Block) Transactions() [][]byte {
	return b.Body.Transactions
}

// RoundReceived returns the round in which the block was received
func (b *Block) RoundReceived() int64 {
	return b.Body.RoundReceived
}

// BlockHash returns the Hash of the block (used for API)
func (b *Block) BlockHash() ([]byte, error) {
	hashBytes, err := b.ProtoMarshal()
	if err != nil {
		return nil, err
	}
	return crypto.Keccak256(hashBytes), nil
}

// BlockHex returns the Hex of the block (used for API)
func (b *Block) BlockHex() string {
	hash, _ := b.BlockHash()
	return fmt.Sprintf("0x%X", hash)
}

// GetBlockSignatures returns the block signatures for the block
func (b *Block) GetBlockSignatures() []BlockSignature {
	res := make([]BlockSignature, len(b.Signatures))
	i := 0
	for val, sig := range b.Signatures {
		validatorBytes, _ := hex.DecodeString(val[2:])
		res[i] = BlockSignature{
			Validator: validatorBytes,
			Index:     b.Index(),
			Signature: sig,
		}
		i++
	}
	return res
}

// GetSignature returns all validator signatures for the block
func (b *Block) GetSignature(validator string) (res BlockSignature, err error) {
	sig, ok := b.Signatures[validator]
	if !ok {
		return res, fmt.Errorf("signature not found")
	}

	validatorBytes, _ := hex.DecodeString(validator[2:])
	return BlockSignature{
		Validator: validatorBytes,
		Index:     b.Index(),
		Signature: sig,
	}, nil
}

// AppendTransactions appends the transactions to the block body
func (b *Block) AppendTransactions(txs [][]byte) {
	b.Body.Transactions = append(b.Body.Transactions, txs...)
}

// ProtoMarshal marshals the block into protobuff
func (b *Block) ProtoMarshal() ([]byte, error) {
	var bf proto.Buffer
	bf.SetDeterministic(true)
	if err := bf.Marshal(b); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

// ProtoUnmarshal unamrshals protobuff into a block
func (b *Block) ProtoUnmarshal(data []byte) error {
	return proto.Unmarshal(data, b)
}

// Sign the block for this node
func (b *Block) Sign(privKey *ecdsa.PrivateKey) (bs BlockSignature, err error) {
	signBytes, err := b.Body.Hash()
	if err != nil {
		return bs, err
	}
	R, S, err := crypto.Sign(privKey, signBytes)
	if err != nil {
		return bs, err
	}
	signature := BlockSignature{
		Validator: crypto.FromECDSAPub(&privKey.PublicKey),
		Index:     b.Index(),
		Signature: crypto.EncodeSignature(R, S),
	}

	return signature, nil
}

// SetSignature sets the known blocksignatures for the block
func (b *Block) SetSignature(bs BlockSignature) error {
	b.Signatures[bs.ValidatorHex()] = bs.Signature
	return nil
}

// Verify verifies a blocksignature is from the node that signed
func (b *Block) Verify(sig BlockSignature) (bool, error) {
	signBytes, err := b.Body.Hash()
	if err != nil {
		return false, err
	}

	pubKey := crypto.ToECDSAPub(sig.Validator)

	r, s, err := crypto.DecodeSignature(sig.Signature)
	if err != nil {
		return false, err
	}

	return crypto.Verify(pubKey, signBytes, r, s), nil
}

// ListBytesEquals compares the equality of two lists
func ListBytesEquals(this [][]byte, that [][]byte) bool {
	if len(this) != len(that) {
		return false
	}
	for i, v := range this {
		if !bytes.Equal(v, that[i]) {
			return false
		}
	}
	return true
}

// Equals compares the equality of a block body
func (bb *BlockBody) Equals(that *BlockBody) bool {
	return bb.Index == that.Index &&
		bb.RoundReceived == that.RoundReceived &&
		ListBytesEquals(bb.Transactions, that.Transactions)
}

// Equals compares the equality of wire block signatures
func (wbs *WireBlockSignature) Equals(that *WireBlockSignature) bool {
	return wbs.Index == that.Index && wbs.Signature == that.Signature
}

// MapStringsEquals compares the equality of two string maps
func MapStringsEquals(this map[string]string, that map[string]string) bool {
	if len(this) != len(that) {
		return false
	}
	for k, v := range this {
		v1, ok := that[k]
		if !ok || v != v1 {
			return false
		}
	}
	return true
}

// Equals compares the equality of two blocks
func (b *Block) Equals(that *Block) bool {
	return b.Body.Equals(that.Body) &&
		MapStringsEquals(b.Signatures, that.Signatures) &&
		b.Hex == that.Hex &&
		bytes.Equal(b.Hash, that.Hash) &&
		bytes.Equal(b.FrameHash, that.FrameHash) &&
		bytes.Equal(b.StateHash, that.StateHash)
}
