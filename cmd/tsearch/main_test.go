package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/Fantom-foundation/go-opera/gossip"
	"github.com/Fantom-foundation/go-opera/integration"
	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/kvdb/flushable"
	"github.com/Fantom-foundation/lachesis-base/kvdb/leveldb"
)

var (
	datadir   = "/home/fabel/Work/fantom/blockchains/mainnet/chaindata"
	blockFrom = idx.Block(0x256e898) // 39250072 - 39350071 = -99999
	blockTo   = idx.Block(0x2586f37)
	pattern   = [][]common.Hash{
		[]common.Hash{},
		[]common.Hash{common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")},
		[]common.Hash(nil),
		[]common.Hash{common.HexToHash("0x00000000000000000000000089716ad7edc3be3b35695789c475f3e7a3deb12a")},
	}
)

func TestDirect(t *testing.T) {
	require := require.New(t)

	db, err := leveldb.New(datadir+"/gossip", cache64mb(""), 0, nil, nil)
	require.NoError(err)
	defer db.Close()

	const pos = 1

	it := db.NewIterator(
		append([]byte("Lt"), append(pattern[pos][0].Bytes(), pos)...),
		uintToBytes(uint64(blockFrom)))
	defer it.Release()

	rows := 0
	defer func(r *int) {
		fmt.Printf("rows = %d\n", *r)
	}(&rows)
	for it.Next() {
		rows++
		n := idx.Block(bytesToUint(it.Key()[(2 + 32 + 1) : (2+32+1)+8]))
		if n > blockTo {
			break
		}
		if rows%10000 == 0 {
			fmt.Printf("rows\t~ %d\t%d\n", rows, n)
		}
	}
}

func xTestMain(t *testing.T) {
	require := require.New(t)

	rawProducer := leveldb.NewProducer(datadir, cache64mb)
	dbs := flushable.NewSyncedPool(rawProducer, integration.FlushIDKey)
	dbs.Initialize(rawProducer.Names())

	cfg := gossip.LiteStoreConfig()
	gdb := gossip.NewStore(dbs, cfg)
	defer gdb.Close()

	edb := gdb.EvmStore()
	logs, err := edb.EvmLogs.FindInBlocks(context.TODO(), blockFrom, blockTo, pattern)
	t.Logf("Got: %d", len(logs))
	require.NoError(err)
}

func cache64mb(string) int {
	return 64 * opt.MiB
}
func uintToBytes(n uint64) []byte {
	return bigendian.Uint64ToBytes(n)
}

func bytesToUint(b []byte) uint64 {
	return bigendian.BytesToUint64(b)
}
