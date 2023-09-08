package ftmclient

import (
	"context"
	"math/big"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// GetEpochBlock returns block height in a beginning of an epoch.
func (ec *Client) GetEpochBlock(ctx context.Context, epoch idx.Epoch) (idx.Block, error) {
	var raw hexutil.Uint64
	err := ec.c.CallContext(ctx, &raw, "ftm_getEpochBlock", toBlockNumArg(big.NewInt(int64(epoch))))

	return idx.Block(raw), err
}
