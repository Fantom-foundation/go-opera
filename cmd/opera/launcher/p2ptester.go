package launcher

import (
	"flag"
	"os"

	"gopkg.in/urfave/cli.v1"
)

// NewP2PTestingNode has to manually create an urfave.cli context,
// setting flags manually, and then starting the node
func NewP2PTestingNode() *OperaNodeStaff {
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
	cfg := MakeAllConfigs(ctx)
	genesisStore := MakeGenesisStore(ctx)

	return MakeOperaNodeStaff(ctx, cfg, genesisStore)
}
