package inter

import (
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/src/inter/wire"
)

// ToWire converts to proto.Message.
func (e *Event) ToWire() *wire.Event {
	if e == nil {
		return nil
	}

	buf, err := rlp.EncodeToBytes(e)
	if err != nil {
		return nil
	}
	return &wire.Event{
		RlpEncoded: buf,
	}
}

// WireToEvent converts from wire.
func WireToEvent(w *wire.Event) *Event {
	if w == nil {
		return nil
	}

	e := &Event{}
	err := rlp.DecodeBytes(w.RlpEncoded, e)
	if err != nil {
		return nil
	}
	return e
}

// ToWire converts to proto.Message.
func (b *Block) ToWire() *wire.Block {
	if b == nil {
		return nil
	}

	buf, err := rlp.EncodeToBytes(b)
	if err != nil {
		return nil
	}
	return &wire.Block{
		RlpEncoded: buf,
	}
}

// WireToBlock converts from wire.
func WireToBlock(w *wire.Block) *Block {
	if w == nil {
		return nil
	}

	b := &Block{}
	err := rlp.DecodeBytes(w.RlpEncoded, b)
	if err != nil {
		return nil
	}
	return b
}

// ToWire converts to proto.Message.
func (tt *ExtTxns) ToWire() *wire.ExtTxns {
	buf, err := rlp.EncodeToBytes(tt)
	if err != nil {
		return nil
	}
	return &wire.ExtTxns{
		RlpEncoded: buf,
	}
}

// WireToExtTxns converts from wire.
func WireToExtTxns(w *wire.Event) *ExtTxns {
	tt := &ExtTxns{}
	err := rlp.DecodeBytes(w.RlpEncoded, tt)
	if err != nil {
		return nil
	}
	return tt
}

// ToWire converts to wire.
func (tx *InternalTransaction) ToWire() *wire.InternalTransaction {
	buf, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return nil
	}
	return &wire.InternalTransaction{
		RlpEncoded: buf,
	}
}

// WireToInternalTransaction converts from wire.
func WireToInternalTransaction(w *wire.InternalTransaction) *InternalTransaction {
	tx := &InternalTransaction{}
	err := rlp.DecodeBytes(w.RlpEncoded, tx)
	if err != nil {
		return nil
	}
	return tx
}

// InternalTransactionsToWire converts to wire.
func InternalTransactionsToWire(tt []*InternalTransaction) []*wire.InternalTransaction {
	if tt == nil {
		return nil
	}
	res := make([]*wire.InternalTransaction, len(tt))
	for i, t := range tt {
		res[i] = t.ToWire()
	}

	return res
}

// WireToInternalTransactions converts from wire.
func WireToInternalTransactions(tt []*wire.InternalTransaction) []*InternalTransaction {
	if tt == nil {
		return nil
	}
	res := make([]*InternalTransaction, len(tt))
	for i, w := range tt {
		res[i] = WireToInternalTransaction(w)
	}

	return res
}
