package inter

import (
	"fmt"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/wire"
)

// ExtTxns is a slice or hash of external transactions.
type ExtTxns struct {
	Value [][]byte
	Hash  *hash.Hash
}

// ToWire converts to proto.Message.
func (tt *ExtTxns) ToWire() (*wire.Event_ExtTxnsValue, *wire.Event_ExtTxnsHash) {
	var h []byte
	if tt.Hash == nil {
		h = hash.Of(tt.Value...).Bytes()
	} else {
		h = tt.Hash.Bytes()
	}

	return &wire.Event_ExtTxnsValue{
			ExtTxnsValue: &wire.ExtTxns{
				List: tt.Value,
			},
		},
		&wire.Event_ExtTxnsHash{
			ExtTxnsHash: h,
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
	case *wire.Event_ExtTxnsHash:
		h := hash.FromBytes(w.GetExtTxnsHash())
		return ExtTxns{
			Hash: &h,
		}
	case nil:
		return ExtTxns{}
	default:
		panic(fmt.Errorf("Event.ExternalTransactions has unexpected type %T", x))
	}
}
