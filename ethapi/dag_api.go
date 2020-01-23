package ethapi

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/beorn7/perks/histogram"
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
func NewPublicDAGChainAPI(b Backend) *PublicDAGChainAPI {
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

// ValidatorVersions returns published node version of each validator.
// If validator didn't have an event in beginning of epoch, then it will not be listed.
func (s *PublicDebugAPI) ValidatorVersions(ctx context.Context) (map[hexutil.Uint64]string, error) {
	epoch := rpc.LatestBlockNumber
	maxDepth := idx.Event(20)
	prefix := []byte("v-")
	versions := map[hexutil.Uint64]string{}

	err := s.b.ForEachEvent(ctx, epoch, func(event *inter.Event) bool {
		creator := hexutil.Uint64(event.Creator)
		if bytes.HasPrefix(event.Extra, prefix) {
			version := string(event.Extra[len(prefix):])
			versions[creator] = version
		} else if _, ok := versions[creator]; !ok {
			versions[creator] = "not found"
		}

		return event.Seq <= maxDepth // iterate until first met event with high seq
	})
	return versions, err
}

// TtfReport for a range of blocks
// Verbosity. Number. If 0, then include only avg, min, max.
// Verbosity. Number. If >= 1, then include histogram with 6 bins.
// Verbosity. Number. If >= 2, then include histogram with 16 bins.
// Verbosity. Number. If >= 3, then include raw data.
func (s *PublicDebugAPI) TtfReport(ctx context.Context, untilBlock rpc.BlockNumber, maxBlocks hexutil.Uint64, verbosity hexutil.Uint64) (map[string]interface{}, error) {
	ttfs, err := s.b.TtfReport(ctx, untilBlock, idx.Block(maxBlocks))
	if err != nil {
		return nil, err
	}
	ttfsRpc := map[string]interface{}{}

	hist := histogram.New(6)
	if verbosity >= 2 {
		hist = histogram.New(16)
	}

	rawTtfs := map[string]interface{}{}
	totalTtf := time.Duration(0)
	minTtf := time.Duration(0)
	maxTtf := time.Duration(0)
	for id, ttf := range ttfs {
		rawTtfs[id.FullID()] = hexutil.Uint64(ttf)

		// stats
		totalTtf += ttf
		if minTtf == 0 || minTtf > ttf {
			minTtf = ttf
		}
		if maxTtf == 0 || maxTtf < ttf {
			maxTtf = ttf
		}

		hist.Insert(float64(ttf))
	}

	avgTtf := totalTtf
	if len(ttfs) > 0 {
		avgTtf /= time.Duration(len(ttfs))
	}

	if verbosity >= 0 {
		statsTtfs := map[string]interface{}{}
		statsTtfs["avg"] = hexutil.Uint64(avgTtf)
		statsTtfs["min"] = hexutil.Uint64(minTtf)
		statsTtfs["max"] = hexutil.Uint64(maxTtf)
		statsTtfs["samples"] = hexutil.Uint64(len(ttfs))
		ttfsRpc["stats"] = statsTtfs
	}
	if verbosity >= 1 {
		histRpc := make([]map[string]interface{}, len(hist.Bins()))
		for i, bin := range hist.Bins() {
			histRpc[i] = map[string]interface{}{}
			histRpc[i]["count"] = hexutil.Uint64(bin.Count)
			histRpc[i]["mean"] = hexutil.Uint64(bin.Mean())
		}
		ttfsRpc["histogram"] = histRpc
	}
	if verbosity >= 3 {
		ttfsRpc["raw"] = rawTtfs
	}

	return ttfsRpc, nil
}
