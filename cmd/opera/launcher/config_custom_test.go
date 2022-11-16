package launcher

import (
	"bytes"
	"testing"

	"github.com/Fantom-foundation/lachesis-base/abft"
	"github.com/Fantom-foundation/lachesis-base/utils/cachescale"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/gossip"
	"github.com/Fantom-foundation/go-opera/gossip/emitter"
	"github.com/Fantom-foundation/go-opera/vecmt"
)

func TestConfigFile(t *testing.T) {
	cacheRatio := cachescale.Ratio{
		Base:   uint64(DefaultCacheSize*1 - ConstantCacheSize),
		Target: uint64(DefaultCacheSize*2 - ConstantCacheSize),
	}

	src := config{
		Node:          defaultNodeConfig(),
		Opera:         gossip.DefaultConfig(cacheRatio),
		Emitter:       emitter.DefaultConfig(),
		TxPool:        evmcore.DefaultTxPoolConfig,
		OperaStore:    gossip.DefaultStoreConfig(cacheRatio),
		Lachesis:      abft.DefaultConfig(),
		LachesisStore: abft.DefaultStoreConfig(cacheRatio),
		VectorClock:   vecmt.DefaultConfig(cacheRatio),
	}

	canonical := func(nn []*enode.Node) []*enode.Node {
		if len(nn) == 0 {
			return []*enode.Node{}
		}
		return nn
	}

	for name, val := range map[string][]*enode.Node{
		"Nil":     nil,
		"Empty":   {},
		"Default": asDefault,
		"UserDefined": {enode.MustParse(
			"enr:-J-4QJmPmUmu14Pn7gUtRNfKHaWFQpcX6fgqrNheDSWUN6giKtix8Lh6EKfymTdXCI5HKGmyl0C5eOKvem5xdC70hLEBgmlkgnY0gmlwhMCoAQKFb3BlcmHHxoQHxfIKgIlzZWNwMjU2azGhAjYQROWoAXivxhtYYBXGXzQrBTAHGJT9XPP69oUzDDWwhHNuYXDAg3RjcIITuoN1ZHCCE7o",
		)},
	} {
		t.Run(name+"BootstrapNodes", func(t *testing.T) {
			require := require.New(t)

			src.Node.P2P.BootstrapNodes = val
			src.Node.P2P.BootstrapNodesV5 = val

			stream, err := tomlSettings.Marshal(&src)
			require.NoError(err)

			var got config
			err = tomlSettings.NewDecoder(bytes.NewReader(stream)).Decode(&got)
			require.NoError(err)

			{ // toml workaround
				src.Node.P2P.BootstrapNodes = canonical(src.Node.P2P.BootstrapNodes)
				got.Node.P2P.BootstrapNodes = canonical(got.Node.P2P.BootstrapNodes)
				src.Node.P2P.BootstrapNodesV5 = canonical(src.Node.P2P.BootstrapNodesV5)
				got.Node.P2P.BootstrapNodesV5 = canonical(got.Node.P2P.BootstrapNodesV5)
			}

			require.Equal(src.Node.P2P.BootstrapNodes, got.Node.P2P.BootstrapNodes)
			require.Equal(src.Node.P2P.BootstrapNodesV5, got.Node.P2P.BootstrapNodesV5)
		})
	}
}
