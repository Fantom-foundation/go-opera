package ftmclient

import (
	"context"
	"math/big"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/Fantom-foundation/go-opera/inter"
)

// GetEvent returns Lachesis event by hash or short ID.
func (ec *Client) GetEvent(ctx context.Context, h hash.Event) (e inter.EventI, err error) {
	var raw map[string]interface{}
	err = ec.c.CallContext(ctx, &raw, "dag_getEvent", h.Hex())
	if err != nil {
		return
	} else if len(raw) == 0 {
		err = ethereum.NotFound
		return
	}

	e = inter.RPCUnmarshalEvent(raw)
	return
}

// GetEvent returns Lachesis event by hash or short ID.
func (ec *Client) GetEventPayload(ctx context.Context, h hash.Event, inclTx bool) (e inter.EventI, txs []common.Hash, err error) {
	var raw map[string]interface{}
	err = ec.c.CallContext(ctx, &raw, "dag_getEventPayload", h.Hex(), inclTx)
	if err != nil {
		return
	} else if len(raw) == 0 {
		err = ethereum.NotFound
		return
	}

	e = inter.RPCUnmarshalEvent(raw)

	if inclTx {
		vv := raw["transactions"].([]interface{})
		txs = make([]common.Hash, len(vv))
		for i, v := range vv {
			txs[i] = common.HexToHash(v.(string))
		}
	}

	return
}

// GetHeads returns IDs of all the epoch events with no descendants.
// * When epoch is -2 the heads for latest epoch are returned.
// * When epoch is -1 the heads for latest sealed epoch are returned.
func (ec *Client) GetHeads(ctx context.Context, epoch *big.Int) (hash.Events, error) {
	var raw []interface{}
	err := ec.c.CallContext(ctx, &raw, "dag_getHeads", toBlockNumArg(epoch))
	if err != nil {
		return nil, err
	}

	return inter.HexToEventIDs(raw), nil
}

// GetEpochStats returns epoch statistics.
// * When epoch is -2 the statistics for latest epoch is returned.
// * When epoch is -1 the statistics for latest sealed epoch is returned.
func (ec *Client) GetEpochStats(ctx context.Context, epoch *big.Int) (map[string]interface{}, error) {
	var raw map[string]interface{}
	err := ec.c.CallContext(ctx, &raw, "dag_getEpochStats", toBlockNumArg(epoch))
	if err != nil {
		return nil, err
	} else if len(raw) == 0 {
		return nil, ethereum.NotFound
	}

	return raw, nil
}

func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	pending := big.NewInt(-1)
	if number.Cmp(pending) == 0 {
		return "pending"
	}
	return hexutil.EncodeBig(number)
}
