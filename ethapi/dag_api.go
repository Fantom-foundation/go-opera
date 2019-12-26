package ethapi

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/Fantom-foundation/go-lachesis/inter"
)

// PublicDAGChainAPI provides an API to access the directed acyclic graph chain.
// It offers only methods that operate on public data that is freely available to anyone.
type PublicDAGChainAPI struct {
	b Backend
}

// NewPublicDAGChainAPI creates a new DAG chain API.
func NewPublicDagAPI(b Backend) *PublicDAGChainAPI {
	return &PublicDAGChainAPI{b}
}

// GetEventHeader returns the Lachesis event header by hash or short ID.
func (s *PublicDAGChainAPI) GetEventHeader(ctx context.Context, shortEventID string) (map[string]interface{}, error) {
	header, err := s.b.GetEventHeader(ctx, shortEventID)
	if err != nil {
		return nil, err
	}
	if header == nil {
		return nil, fmt.Errorf("event %s not found", shortEventID)
	}
	return RPCMarshalEventHeader(header), nil
}

// GetEvent returns Lachesis event by hash or short ID.
func (s *PublicDAGChainAPI) GetEvent(ctx context.Context, shortEventID string, inclTx bool) (map[string]interface{}, error) {
	event, err := s.b.GetEvent(ctx, shortEventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, fmt.Errorf("event %s not found", shortEventID)
	}
	return RPCMarshalEvent(event, inclTx, false)
}

// GetConsensusTime returns event's consensus time, if event is confirmed.
func (s *PublicDAGChainAPI) GetConsensusTime(ctx context.Context, shortEventID string) (inter.Timestamp, error) {
	return s.b.GetConsensusTime(ctx, shortEventID)
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

// CurrentEpoch returns current epoch number.
func (s *PublicDAGChainAPI) CurrentEpoch(ctx context.Context) hexutil.Uint64 {
	return hexutil.Uint64(s.b.CurrentEpoch(ctx))
}

// GetEpochStats returns epoch statistics.
// * When epoch is -2 the statistics for latest epoch is returned.
// * When epoch is -1 the statistics for latest sealed epoch is returned.
func (s *PublicDAGChainAPI) GetEpochStats(ctx context.Context, requestedEpoch rpc.BlockNumber) (map[string]interface{}, error) {
	stats, err := s.b.GetEpochStats(ctx, requestedEpoch)
	if err != nil {
		return nil, err
	}
	if stats == nil {
		return nil, nil
	}
	return map[string]interface{}{
		"epoch":                 hexutil.Uint64(stats.Epoch),
		"start":                 hexutil.Uint64(stats.Start),
		"end":                   hexutil.Uint64(stats.End),
		"totalFee":              (*hexutil.Big)(stats.TotalFee),
		"totalBaseRewardWeight": (*hexutil.Big)(stats.TotalBaseRewardWeight),
		"totalTxRewardWeight":   (*hexutil.Big)(stats.TotalTxRewardWeight),
	}, nil
}
