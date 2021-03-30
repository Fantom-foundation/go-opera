package ethapi

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/status-im/keycard-go/hexutils"
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

// PrivateDAGChainAPI provides an API to access the directed acyclic graph chain.
// It offers methods that operate on public data that is freely available to anyone,
// but can be hard to execute.
type PrivateDAGChainAPI struct {
	b Backend
}

// NewPrivateDAGChainAPI creates a new DAG chain API.
func NewPrivateDAGChainAPI(b Backend) *PrivateDAGChainAPI {
	return &PrivateDAGChainAPI{b}
}

var ( // consts
	eventsFileHeader  = hexutils.HexToBytes("7e995678")
	eventsFileVersion = hexutils.HexToBytes("00010001")
)

// statsReportLimit is the time limit during import and export after which we
// always print out progress. This avoids the user wondering what's going on.
const statsReportLimit = 8 * time.Second

// ExportEvents writes RLP-encoded events into file.
func (s *PrivateDAGChainAPI) ExportEvents(ctx context.Context, epochFrom, epochTo rpc.BlockNumber, fileName string) error {
	start, reported := time.Now(), time.Time{}
	from := idx.Epoch(epochFrom)
	to := idx.Epoch(epochTo)
	log.Info("Exporting events to file", "file", fileName, "from", from, "to", to)

	// Open the file handle and potentially wrap with a gzip stream
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	var w io.Writer = f
	if strings.HasSuffix(fileName, ".gz") {
		w = gzip.NewWriter(w)
		defer w.(*gzip.Writer).Close()
	}

	// Write header and version
	_, err = w.Write(append(eventsFileHeader, eventsFileVersion...))
	if err != nil {
		return err
	}

	var (
		counter int
		last    hash.Event
	)
	err = s.b.ForEachEventRLP(ctx, epochFrom, func(id hash.Event, event rlp.RawValue) bool {
		if to >= from && id.Epoch() > to {
			return false
		}
		counter++
		_, err = w.Write(event)
		if err != nil {
			return false
		}
		last = id
		if counter%100 == 1 && time.Since(reported) >= statsReportLimit {
			log.Info("Exporting events", "last", last.String(), "exported", counter, "elapsed", common.PrettyDuration(time.Since(start)))
			reported = time.Now()
		}
		return true
	})

	if err != nil {
		log.Error("Exported events", "last", last.String(), "exported", counter, "elapsed", common.PrettyDuration(time.Since(start)), "err", err)
		return err
	}

	log.Info("Exported events", "last", last.String(), "exported", counter, "elapsed", common.PrettyDuration(time.Since(start)))
	return nil
}
