package ethclient

import (
	"context"
	"encoding/json"

	"github.com/ethereum/go-ethereum"
)

// GetEvent returns Lachesis event by hash or short ID.
func (ec *Client) GetEvent(ctx context.Context, shortEventID string, inclTx bool) (raw json.RawMessage, err error) {
	err = ec.c.CallContext(ctx, &raw, "dag_getEvent", shortEventID, inclTx)
	if err != nil {
		return nil, err
	} else if len(raw) == 0 {
		return nil, ethereum.NotFound
	}

	return
}
