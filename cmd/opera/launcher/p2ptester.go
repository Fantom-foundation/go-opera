package launcher

import (
	"flag"
	"os"

	"github.com/Fantom-foundation/go-opera/gossip"
	"github.com/Fantom-foundation/go-opera/inter/validatorpk"
	"github.com/Fantom-foundation/go-opera/opera/genesis"
	"github.com/Fantom-foundation/go-opera/valkeystore"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"gopkg.in/urfave/cli.v1"
)

type P2PTestingNode struct {
	Node      *node.Node
	Service   *gossip.Service
	P2PServer *p2p.Server
	NodeClose func()
	Signer    valkeystore.SignerI
	Store     *gossip.Store
	Genesis   *genesis.Genesis
	PubKey    validatorpk.PubKey
}

// NewP2PTestingNode has to manually create an urfave.cli context,
// setting flags manually, and then starting the node
func NewP2PTestingNode() *P2PTestingNode {
	fs := flag.NewFlagSet("", flag.ExitOnError)
	app := cli.NewApp()

	// define flags manually
	fs.String(CacheFlag.Name, "8000", CacheFlag.Name)
	fs.String(DataDirFlag.Name, "/tmp/d", DataDirFlag.Name)
	fs.String(FakeNetFlag.Name, "4/4", FakeNetFlag.Name)

	dir, err := os.MkdirTemp("", "p2p-testing")
	if err != nil {
		panic(err)
	}
	// set flags manually
	fs.Set(CacheFlag.Name, "8000")
	fs.Set(DataDirFlag.Name, dir)
	fs.Set(FakeNetFlag.Name, "4/4")
	ctx := cli.NewContext(app, fs, nil)
	cfg := makeAllConfigs(ctx)
	genesisStore := mayGetGenesisStore(ctx)

	return makeP2PTestNode(ctx, cfg, genesisStore)
}
