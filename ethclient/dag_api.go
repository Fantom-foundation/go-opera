package ethclient

import (
	"context"

	"github.com/ethereum/go-ethereum"

	"github.com/Fantom-foundation/go-opera/inter"
)

// GetEvent returns Lachesis event by hash or short ID.
func (ec *Client) GetEvent(ctx context.Context, shortEventID string, inclTx bool) (inter.EventI, error) {
	var raw map[string]interface{}
	err := ec.c.CallContext(ctx, &raw, "dag_getEvent", shortEventID, inclTx)
	if err != nil {
		return nil, err
	} else if len(raw) == 0 {
		return nil, ethereum.NotFound
	}

	return inter.RPCUnmarshalEvent(raw), nil
}
