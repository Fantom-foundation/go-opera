package ethapi

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
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

// CurrentEpoch returns current epoch number.
func (s *PublicDAGChainAPI) CurrentEpoch(ctx context.Context) hexutil.Uint64 {
	return hexutil.Uint64(s.b.CurrentEpoch(ctx))
}
