package ethapi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/beorn7/perks/histogram"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
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

func durationToRPC(t time.Duration) string {
	/*if t < 0 {
		t = -t
		return "-" + hexutil.Uint64(t).String()
	}
	return hexutil.Uint64(t).String()*/
	return t.String()
}

func rpcEncodeEventsTimeStats(data map[hash.Event]time.Duration, verbosity hexutil.Uint64) (map[string]interface{}, error) {
	resRPC := map[string]interface{}{}

	hist := histogram.New(6)
	if verbosity >= 2 {
		hist = histogram.New(16)
	}

	raw := map[string]interface{}{}
	totalTtf := time.Duration(0)
	minTtf := time.Duration(0)
	maxTtf := time.Duration(0)
	for id, ttf := range data {
		raw[id.FullID()] = hexutil.Uint64(ttf)

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
	if len(data) > 0 {
		avgTtf /= time.Duration(len(data))
	}

	{
		stats := map[string]interface{}{}
		stats["avg"] = durationToRPC(avgTtf)
		stats["min"] = durationToRPC(minTtf)
		stats["max"] = durationToRPC(maxTtf)
		stats["samples"] = hexutil.Uint64(len(data))
		resRPC["stats"] = stats
	}
	if verbosity >= 1 {
		histRPC := make([]map[string]interface{}, len(hist.Bins()))
		for i, bin := range hist.Bins() {
			histRPC[i] = map[string]interface{}{}
			histRPC[i]["count"] = hexutil.Uint64(bin.Count)
			histRPC[i]["mean"] = durationToRPC(time.Duration(bin.Mean()))
		}
		resRPC["histogram"] = histRPC
	}
	if verbosity >= 3 {
		resRPC["raw"] = raw
	}

	return resRPC, nil
}

// TtfReport for a range of blocks
// maxEvents. Number.  maximum number of events to process
// Verbosity. Number. If 0, then include only avg, min, max.
// Verbosity. Number. If >= 1, then include histogram with 6 bins.
// Verbosity. Number. If >= 2, then include histogram with 16 bins.
// Verbosity. Number. If >= 3, then include raw data.
// Mode. String. One of {"arrival_time", "claimed_time"}
func (s *PublicDebugAPI) TtfReport(ctx context.Context, untilBlock rpc.BlockNumber, maxBlocks hexutil.Uint64, mode string, verbosity hexutil.Uint64) (map[string]interface{}, error) {
	if mode != "arrival_time" && mode != "claimed_time" {
		return nil, errors.New("mode must be one of {arrival_time, claimed_time}")
	}

	ttfs, err := s.b.TtfReport(ctx, untilBlock, idx.Block(maxBlocks), mode)
	if err != nil {
		return nil, err
	}
	return rpcEncodeEventsTimeStats(ttfs, verbosity)
}

// ValidatorVersions returns published node version of each validator.
// If validator didn't have an event in beginning of epoch, then it will not be listed.
func (s *PublicDebugAPI) ValidatorVersions(ctx context.Context, epoch rpc.BlockNumber, maxEvents hexutil.Uint64) (map[hexutil.Uint64]string, error) {
	processed := 0
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
		processed++

		return processed < int(maxEvents) // iterate until first met event with high seq
	})
	return versions, err
}

// ValidatorTimeDrifts for an epoch
// maxEvents. Number.  maximum number of events to process
// Verbosity. Number. If 0, then include only avg, min, max.
// Verbosity. Number. If >= 1, then include histogram with 6 bins.
// Verbosity. Number. If >= 2, then include histogram with 16 bins.
// Verbosity. Number. If >= 3, then include raw data.
func (s *PublicDebugAPI) ValidatorTimeDrifts(ctx context.Context, epoch rpc.BlockNumber, maxEvents hexutil.Uint64, verbosity hexutil.Uint64) (map[hexutil.Uint64]map[string]interface{}, error) {
	validatorsDrifts, err := s.b.ValidatorTimeDrifts(ctx, epoch, idx.Event(maxEvents))

	resRPC := map[hexutil.Uint64]map[string]interface{}{}
	for v, drifts := range validatorsDrifts {
		vRPC, err := rpcEncodeEventsTimeStats(drifts, verbosity)
		if err != nil {
			return nil, err
		}
		resRPC[hexutil.Uint64(v)] = vRPC
	}

	return resRPC, err
}
