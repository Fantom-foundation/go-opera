package ethapi

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

// PublicDAGChainAPI provides an API to access the directed acyclic graph chain.
// It offers only methods that operate on public data that is freely available to anyone.
type PublicDAGChainAPI struct {
	b Backend
}

// NewPublicDAGChainAPI creates a new DAG chain API.
func NewPublicDAGChainAPI(b Backend) *PublicDAGChainAPI {
	return &PublicDAGChainAPI{b}
}

// GetEvent returns the Lachesis event header by hash or short ID.
func (s *PublicDAGChainAPI) GetEvent(ctx context.Context, shortEventID string) (map[string]interface{}, error) {
	header, err := s.b.GetEvent(ctx, shortEventID)
	if err != nil {
		return nil, err
	}
	if header == nil {
		return nil, fmt.Errorf("event %s not found", shortEventID)
	}
	return RPCMarshalEventHeader(header), nil
}

// GetEventPayload returns Lachesis event by hash or short ID.
func (s *PublicDAGChainAPI) GetEventPayload(ctx context.Context, shortEventID string, inclTx bool) (map[string]interface{}, error) {
	event, err := s.b.GetEventPayload(ctx, shortEventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, fmt.Errorf("event %s not found", shortEventID)
	}
	return RPCMarshalEvent(event, inclTx, false)
}

// GetHeads returns IDs of all the epoch events with no descendants.
// * When epoch is -2 the heads for latest epoch are returned.
// * When epoch is -1 the heads for latest sealed epoch are returned.
func (s *PublicDAGChainAPI) GetHeads(ctx context.Context, epoch rpc.BlockNumber) ([]hexutil.Bytes, error) {
	res, err := s.b.GetHeads(ctx, epoch)

	if err != nil {
		return nil, err
	}

	return eventIDsToHex(res), nil
}

// GetEpochStats returns epoch statistics.
// * When epoch is -2 the statistics for latest epoch is returned.
// * When epoch is -1 the statistics for latest sealed epoch is returned.
func (s *PublicBlockChainAPI) GetEpochStats(ctx context.Context, requestedEpoch rpc.BlockNumber) (map[string]interface{}, error) {
	log.Warn("GetEpochStats API call is deprecated. Consider retrieving data from SFC v3 contract.")
	if requestedEpoch != rpc.LatestBlockNumber && requestedEpoch != rpc.BlockNumber(s.b.CurrentEpoch(ctx))-1 {
		return nil, errors.New("getEpochStats API call doesn't support retrieving previous sealed epochs")
	}
	start, end := s.b.SealedEpochTiming(ctx)
	return map[string]interface{}{
		"epoch":                 hexutil.Uint64(s.b.CurrentEpoch(ctx) - 1),
		"start":                 hexutil.Uint64(start),
		"end":                   hexutil.Uint64(end),
		"totalFee":              (*hexutil.Big)(new(big.Int)),
		"totalBaseRewardWeight": (*hexutil.Big)(new(big.Int)),
		"totalTxRewardWeight":   (*hexutil.Big)(new(big.Int)),
	}, nil
}
