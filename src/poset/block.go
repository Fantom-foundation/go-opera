package poset

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"reflect"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/golang/protobuf/proto"
)

//StateHash is the hash of the current state of transactions, if you have one
//node talking to an app, and another set of nodes talking to inmem, the
//stateHash will be different
//statehash should be ignored for validator checking

//json encoding of body only
func (bb *BlockBody) ProtoMarshal() ([]byte, error) {
	var bf proto.Buffer
	bf.SetDeterministic(true)
	if err := bf.Marshal(bb); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

func (bb *BlockBody) ProtoUnmarshal(data []byte) error {
	return proto.Unmarshal(data, bb)
}

func (bb *BlockBody) Hash() ([]byte, error) {
	hashBytes, err := bb.ProtoMarshal()
	if err != nil {
		return nil, err
	}
	return crypto.SHA256(hashBytes), nil
}

//------------------------------------------------------------------------------

func (m *BlockSignature) ValidatorHex() string {
	return fmt.Sprintf("0x%X", m.Validator)
}

func (m *BlockSignature) ProtoMarshal() ([]byte, error) {
	var bf proto.Buffer
	bf.SetDeterministic(true)
	if err := bf.Marshal(m); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

func (m *BlockSignature) ProtoUnmarshal(data []byte) error {
	return proto.Unmarshal(data, m)
}

func (m *BlockSignature) ToWire() WireBlockSignature {
	return WireBlockSignature{
		Index:     m.Index,
		Signature: m.Signature,
	}
}

func (m *BlockSignature) Equals(that *BlockSignature) bool {
	return reflect.DeepEqual(m.Validator, that.Validator) &&
		m.Index == that.Index &&
		m.Signature == that.Signature
}

//------------------------------------------------------------------------------

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

func NewBlock(blockIndex, roundReceived int64, frameHash []byte, txs [][]byte) Block {
	body := BlockBody{
		Index:         blockIndex,
		RoundReceived: roundReceived,
		Transactions:  txs,
	}
	return Block{
		Body:       &body,
		FrameHash:  frameHash,
		Signatures: make(map[string]string),
	}
}

func (b *Block) Index() int64 {
	return b.Body.Index
}

func (b *Block) Transactions() [][]byte {
	return b.Body.Transactions
}

func (b *Block) RoundReceived() int64 {
	return b.Body.RoundReceived
}

func (b *Block) BlockHash() ([]byte, error) {
	hashBytes, err := b.ProtoMarshal()
	if err != nil {
		return nil, err
	}
	return crypto.SHA256(hashBytes), nil
}

func (b *Block) BlockHex() string {
	hash, _ := b.BlockHash()
	return fmt.Sprintf("0x%X", hash)
}

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

func (b *Block) AppendTransactions(txs [][]byte) {
	b.Body.Transactions = append(b.Body.Transactions, txs...)
}

func (b *Block) ProtoMarshal() ([]byte, error) {
	var bf proto.Buffer
	bf.SetDeterministic(true)
	if err := bf.Marshal(b); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

func (b *Block) ProtoUnmarshal(data []byte) error {
	return proto.Unmarshal(data, b)
}

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

func (b *Block) SetSignature(bs BlockSignature) error {
	b.Signatures[bs.ValidatorHex()] = bs.Signature
	return nil
}

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

func ListBytesEquals(this [][]byte, that [][]byte) bool {
	if len(this) != len(that) {
		return false
	}
	for i, v := range this {
		if !BytesEquals(v, that[i]) {
			return false
		}
	}
	return true
}

func (this *BlockBody) Equals(that *BlockBody) bool {
	return this.Index == that.Index &&
		this.RoundReceived == that.RoundReceived &&
		ListBytesEquals(this.Transactions, that.Transactions)
}

func (this *WireBlockSignature) Equals(that *WireBlockSignature) bool {
	return this.Index == that.Index && this.Signature == that.Signature
}

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

func (this *Block) Equals(that *Block) bool {
	return this.Body.Equals(that.Body) &&
		MapStringsEquals(this.Signatures, that.Signatures) &&
		BytesEquals(this.Hash, that.Hash) &&
		this.Hex == that.Hex
}
