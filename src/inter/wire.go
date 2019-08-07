package inter

import (
	"fmt"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/inter/wire"
)

// ToWire converts to proto.Message.
func (e *Event) ToWire() *wire.Event {
	if e == nil {
		return nil
	}

	extTxns := e.ExternalTransactions.ToWire()

	return &wire.Event{
		SfNum:                uint64(e.Epoch),
		Seq:                  uint64(e.Seq),
		Creator:              e.Creator.Hex(),
		Parents:              e.Parents.ToWire(),
		LamportTime:          uint64(e.Lamport),
		InternalTransactions: InternalTransactionsToWire(e.InternalTransactions),
		ExternalTransactions: extTxns,
		Sign:                 e.Sig,
	}
}

// WireToEvent converts from wire.
func WireToEvent(w *wire.Event) *Event {
	if w == nil {
		return nil
	}
	return &Event{
		EventHeader: EventHeader{
			EventHeaderData: EventHeaderData{
				Epoch:   idx.SuperFrame(w.SfNum),
				Seq:     idx.Event(w.Seq),
				Creator: hash.HexToPeer(w.Creator),
				Parents: hash.WireToEvents(w.Parents),
				Lamport: idx.Lamport(w.LamportTime),
			},
			Sig: w.Sign,
		},
		InternalTransactions: WireToInternalTransactions(w.InternalTransactions),
		ExternalTransactions: WireToExtTxns(w),
	}
}

// ToWire converts to proto.Message.
func (tt *ExtTxns) ToWire() *wire.Event_ExtTxnsValue {
	return &wire.Event_ExtTxnsValue{
		ExtTxnsValue: &wire.ExtTxns{
			List: tt.Value,
		},
	}
}

// WireToExtTxns converts from wire.
func WireToExtTxns(w *wire.Event) ExtTxns {
	switch x := w.ExternalTransactions.(type) {
	case *wire.Event_ExtTxnsValue:
		if val := w.GetExtTxnsValue(); val != nil {
			return ExtTxns{
				Value: val.List,
			}
		}
		return ExtTxns{}
	case nil:
		return ExtTxns{}
	default:
		panic(fmt.Errorf("Event.ExternalTransactions has unexpected type %T", x))
	}
}

// ToWire converts to wire.
func (tx *InternalTransaction) ToWire() *wire.InternalTransaction {
	if tx == nil {
		return nil
	}
	return &wire.InternalTransaction{
		Nonce:      uint64(tx.Nonce),
		Amount:     uint64(tx.Amount),
		Receiver:   tx.Receiver.Hex(),
		UntilBlock: uint64(tx.UntilBlock),
	}
}

// WireToInternalTransaction converts from wire.
func WireToInternalTransaction(w *wire.InternalTransaction) *InternalTransaction {
	if w == nil {
		return nil
	}
	return &InternalTransaction{
		Nonce:      idx.Txn(w.Nonce),
		Amount:     Stake(w.Amount),
		Receiver:   hash.HexToPeer(w.Receiver),
		UntilBlock: idx.Block(w.UntilBlock),
	}
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
