package launcher

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/Fantom-foundation/lachesis-base/abft"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/params"
	"github.com/naoina/toml"
	"gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/gossip"
	"github.com/Fantom-foundation/go-opera/gossip/gasprice"
	"github.com/Fantom-foundation/go-opera/integration"
	"github.com/Fantom-foundation/go-opera/integration/makegenesis"
	"github.com/Fantom-foundation/go-opera/opera/genesisstore"
	futils "github.com/Fantom-foundation/go-opera/utils"
	"github.com/Fantom-foundation/go-opera/vecmt"
)

var (
	dumpConfigCommand = cli.Command{
		Action:      utils.MigrateFlags(dumpConfig),
		Name:        "dumpconfig",
		Usage:       "Show configuration values",
		ArgsUsage:   "",
		Flags:       append(nodeFlags, testFlags...),
		Category:    "MISCELLANEOUS COMMANDS",
		Description: `The dumpconfig command shows configuration values.`,
	}
	checkConfigCommand = cli.Command{
		Action:      utils.MigrateFlags(checkConfig),
		Name:        "checkconfig",
		Usage:       "Checks configuration file",
		ArgsUsage:   "",
		Flags:       append(nodeFlags, testFlags...),
		Category:    "MISCELLANEOUS COMMANDS",
		Description: `The checkconfig checks configuration file.`,
	}

	configFileFlag = cli.StringFlag{
		Name:  "config",
		Usage: "TOML configuration file",
	}

	// DataDirFlag defines directory to store Lachesis state and user's wallets
	DataDirFlag = utils.DirectoryFlag{
		Name:  "datadir",
		Usage: "Data directory for the databases and keystore",
		Value: utils.DirectoryString(DefaultDataDir()),
	}

	// FakeNetFlag enables special testnet, where validators are automatically created
	GenesisFlag = cli.StringFlag{
		Name:  "genesis",
		Usage: "'path to genesis file' - sets the network genesis configuration.",
	}
)

// These settings ensure that TOML keys use the same names as Go struct fields.
var tomlSettings = toml.Config{
	NormFieldName: func(rt reflect.Type, key string) string {
		return key
	},
	FieldToKey: func(rt reflect.Type, field string) string {
		return field
	},
	MissingField: func(rt reflect.Type, field string) error {
		return fmt.Errorf("field '%s' is not defined in %s", field, rt.String())
	},
}

type config struct {
	Node          node.Config
	Opera         gossip.Config
	OperaStore    gossip.StoreConfig
	Lachesis      abft.Config
	LachesisStore abft.StoreConfig
	VectorClock   vecmt.IndexConfig
}

func (c *config) AppConfigs() integration.Configs {
	return integration.Configs{
		Opera:         c.Opera,
		OperaStore:    c.OperaStore,
		Lachesis:      c.Lachesis,
		LachesisStore: c.LachesisStore,
		VectorClock:   c.VectorClock,
	}
}

func loadAllConfigs(file string, cfg *config) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	err = tomlSettings.NewDecoder(bufio.NewReader(f)).Decode(cfg)
	// Add file name to errors that have a line number.
	if _, ok := err.(*toml.LineError); ok {
		err = errors.New(file + ", " + err.Error())
	}
	if err != nil {
		return errors.New(fmt.Sprintf("TOML config file error: %v.\n"+
			"Use 'dumpconfig' command to get an example config file.\n"+
			"If node was recently upgraded and a previous network config file is used, then check updates for the config file.", err))
	}
	return err
}

func getOperaGenesis(ctx *cli.Context) integration.InputGenesis {

	var genesis integration.InputGenesis
	switch {
	case ctx.GlobalIsSet(FakeNetFlag.Name):
		_, num, err := parseFakeGen(ctx.GlobalString(FakeNetFlag.Name))
		if err != nil {
			log.Crit("Invalid flag", "flag", FakeNetFlag.Name, "err", err)
		}
		fakeGenesisStore := makegenesis.FakeGenesisStore(num, futils.ToFtm(1000000000), futils.ToFtm(5000000))
		genesis = integration.InputGenesis{
			Hash: fakeGenesisStore.Hash(),
			Read: func(store *genesisstore.Store) error {
				buf := bytes.NewBuffer(nil)
				err = fakeGenesisStore.Export(buf)
				if err != nil {
					return err
				}
				return store.Import(buf)
			},
			Close: func() error {
				return nil
			},
		}
	case ctx.GlobalIsSet(GenesisFlag.Name):
		genesisPath := ctx.GlobalString(GenesisFlag.Name)

		genesisFile, err := os.Open(genesisPath)
		if err != nil {
			utils.Fatalf("Failed to open genesis file: %v", err)
		}
		inputGenesisHash, readGenesisStore, err := genesisstore.OpenGenesisStore(genesisFile)
		if err != nil {
			utils.Fatalf("Failed to read genesis file: %v", err)
		}

		genesis = integration.InputGenesis{
			Hash:  inputGenesisHash,
			Read:  readGenesisStore,
			Close: genesisFile.Close,
		}
	default:
		log.Crit("Network genesis is not specified")
	}
	return genesis
}

func setBootnodes(ctx *cli.Context, urls []string, cfg *node.Config) {
	for _, url := range urls {
		if url != "" {
			node, err := enode.ParseV4(url)
			if err != nil {
				log.Error("Bootstrap URL invalid", "enode", url, "err", err)
				continue
			}
			cfg.P2P.BootstrapNodes = append(cfg.P2P.BootstrapNodes, node)
		}
	}
}

func setDataDir(ctx *cli.Context, cfg *node.Config) {
	defaultDataDir := DefaultDataDir()

	switch {
	case ctx.GlobalIsSet(utils.DataDirFlag.Name):
		cfg.DataDir = ctx.GlobalString(utils.DataDirFlag.Name)
	case ctx.GlobalIsSet(FakeNetFlag.Name):
		_, num, err := parseFakeGen(ctx.GlobalString(FakeNetFlag.Name))
		if err != nil {
			log.Crit("Invalid flag", "flag", FakeNetFlag.Name, "err", err)
		}
		cfg.DataDir = filepath.Join(defaultDataDir, fmt.Sprintf("fakenet-%d", num))
	case ctx.GlobalBool(utils.LegacyTestnetFlag.Name):
		cfg.DataDir = filepath.Join(defaultDataDir, "testnet")
	default:
		cfg.DataDir = defaultDataDir
	}
}

func setGPO(ctx *cli.Context, cfg *gasprice.Config) {
	if ctx.GlobalIsSet(utils.LegacyGpoBlocksFlag.Name) {
		cfg.Blocks = ctx.GlobalInt(utils.LegacyGpoBlocksFlag.Name)
	}
	if ctx.GlobalIsSet(utils.GpoBlocksFlag.Name) {
		cfg.Blocks = ctx.GlobalInt(utils.GpoBlocksFlag.Name)
	}
	if ctx.GlobalIsSet(utils.LegacyGpoPercentileFlag.Name) {
		cfg.Percentile = ctx.GlobalInt(utils.LegacyGpoPercentileFlag.Name)
	}
	if ctx.GlobalIsSet(utils.GpoPercentileFlag.Name) {
		cfg.Percentile = ctx.GlobalInt(utils.GpoPercentileFlag.Name)
	}
}

func setTxPool(ctx *cli.Context, cfg *evmcore.TxPoolConfig) {
	if ctx.GlobalIsSet(utils.TxPoolLocalsFlag.Name) {
		locals := strings.Split(ctx.GlobalString(utils.TxPoolLocalsFlag.Name), ",")
		for _, account := range locals {
			if trimmed := strings.TrimSpace(account); !common.IsHexAddress(trimmed) {
				utils.Fatalf("Invalid account in --txpool.locals: %s", trimmed)
			} else {
				cfg.Locals = append(cfg.Locals, common.HexToAddress(account))
			}
		}
	}
	if ctx.GlobalIsSet(utils.TxPoolNoLocalsFlag.Name) {
		cfg.NoLocals = ctx.GlobalBool(utils.TxPoolNoLocalsFlag.Name)
	}
	if ctx.GlobalIsSet(utils.TxPoolJournalFlag.Name) {
		cfg.Journal = ctx.GlobalString(utils.TxPoolJournalFlag.Name)
	}
	if ctx.GlobalIsSet(utils.TxPoolRejournalFlag.Name) {
		cfg.Rejournal = ctx.GlobalDuration(utils.TxPoolRejournalFlag.Name)
	}
	if ctx.GlobalIsSet(utils.TxPoolPriceLimitFlag.Name) {
		cfg.PriceLimit = ctx.GlobalUint64(utils.TxPoolPriceLimitFlag.Name)
	}
	if ctx.GlobalIsSet(utils.TxPoolPriceBumpFlag.Name) {
		cfg.PriceBump = ctx.GlobalUint64(utils.TxPoolPriceBumpFlag.Name)
	}
	if ctx.GlobalIsSet(utils.TxPoolAccountSlotsFlag.Name) {
		cfg.AccountSlots = ctx.GlobalUint64(utils.TxPoolAccountSlotsFlag.Name)
	}
	if ctx.GlobalIsSet(utils.TxPoolGlobalSlotsFlag.Name) {
		cfg.GlobalSlots = ctx.GlobalUint64(utils.TxPoolGlobalSlotsFlag.Name)
	}
	if ctx.GlobalIsSet(utils.TxPoolAccountQueueFlag.Name) {
		cfg.AccountQueue = ctx.GlobalUint64(utils.TxPoolAccountQueueFlag.Name)
	}
	if ctx.GlobalIsSet(utils.TxPoolGlobalQueueFlag.Name) {
		cfg.GlobalQueue = ctx.GlobalUint64(utils.TxPoolGlobalQueueFlag.Name)
	}
	if ctx.GlobalIsSet(utils.TxPoolLifetimeFlag.Name) {
		cfg.Lifetime = ctx.GlobalDuration(utils.TxPoolLifetimeFlag.Name)
	}
}

func gossipConfigWithFlags(ctx *cli.Context, src gossip.Config) (gossip.Config, error) {
	cfg := src

	// Avoid conflicting network flags
	utils.CheckExclusive(ctx, FakeNetFlag, utils.DeveloperFlag, utils.LegacyTestnetFlag)
	utils.CheckExclusive(ctx, FakeNetFlag, utils.DeveloperFlag, utils.ExternalSignerFlag) // Can't use both ephemeral unlocked and external signer

	setGPO(ctx, &cfg.GPO)
	setTxPool(ctx, &cfg.TxPool)

	// TODO cache config
	//if ctx.GlobalIsSet(utils.CacheFlag.Name) || ctx.GlobalIsSet(utils.CacheDatabaseFlag.Name) {
	//	cfg.DatabaseCache = ctx.GlobalInt(utils.CacheFlag.Name) * ctx.GlobalInt(utils.CacheDatabaseFlag.Name) / 100
	//}
	//if ctx.GlobalIsSet(utils.CacheFlag.Name) || ctx.GlobalIsSet(CacheTrieFlag.Name) {
	//	cfg.TrieCleanCache = ctx.GlobalInt(utils.CacheFlag.Name) * ctx.GlobalInt(CacheTrieFlag.Name) / 100
	//}
	//if ctx.GlobalIsSet(utils.CacheFlag.Name) || ctx.GlobalIsSet(CacheGCFlag.Name) {
	//	cfg.TrieDirtyCache = ctx.GlobalInt(utils.CacheFlag.Name) * ctx.GlobalInt(CacheGCFlag.Name) / 100
	//}

	if ctx.GlobalIsSet(utils.VMEnableDebugFlag.Name) {
		cfg.EnablePreimageRecording = ctx.GlobalBool(utils.VMEnableDebugFlag.Name)
	}

	if ctx.GlobalIsSet(utils.EWASMInterpreterFlag.Name) {
		cfg.EWASMInterpreter = ctx.GlobalString(utils.EWASMInterpreterFlag.Name)
	}

	if ctx.GlobalIsSet(utils.EVMInterpreterFlag.Name) {
		cfg.EVMInterpreter = ctx.GlobalString(utils.EVMInterpreterFlag.Name)
	}
	if ctx.GlobalIsSet(utils.RPCGlobalGasCap.Name) {
		cfg.RPCGasCap = ctx.GlobalUint64(utils.RPCGlobalGasCap.Name)
	}
	if ctx.GlobalIsSet(utils.RPCGlobalTxFeeCap.Name) {
		cfg.RPCTxFeeCap = ctx.GlobalFloat64(utils.RPCGlobalTxFeeCap.Name)
	}

	err := setValidator(ctx, &cfg.Emitter)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func nodeConfigWithFlags(ctx *cli.Context, cfg node.Config) node.Config {
	utils.SetNodeConfig(ctx, &cfg)

	if !ctx.GlobalIsSet(FakeNetFlag.Name) {
		setBootnodes(ctx, Bootnodes, &cfg)
	}
	setDataDir(ctx, &cfg)
	return cfg
}

func mayMakeAllConfigs(ctx *cli.Context) (*config, error) {
	// Defaults (low priority)
	cfg := config{
		Node:          defaultNodeConfig(),
		Opera:         gossip.DefaultConfig(),
		OperaStore:    gossip.DefaultStoreConfig(),
		Lachesis:      abft.DefaultConfig(),
		LachesisStore: abft.DefaultStoreConfig(),
		VectorClock:   vecmt.DefaultConfig(),
	}
	if ctx.GlobalIsSet(FakeNetFlag.Name) {
		_, num, _ := parseFakeGen(ctx.GlobalString(FakeNetFlag.Name))
		cfg.Opera = gossip.FakeConfig(num)
	}

	// Load config file (medium priority)
	if file := ctx.GlobalString(configFileFlag.Name); file != "" {
		if err := loadAllConfigs(file, &cfg); err != nil {
			return &cfg, err
		}
	}

	// Apply flags (high priority)
	var err error
	cfg.Opera, err = gossipConfigWithFlags(ctx, cfg.Opera)
	if err != nil {
		return nil, err
	}
	cfg.Node = nodeConfigWithFlags(ctx, cfg.Node)

	if err := cfg.Opera.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func makeAllConfigs(ctx *cli.Context) *config {
	cfg, err := mayMakeAllConfigs(ctx)
	if err != nil {
		utils.Fatalf("%v", err)
	}
	return cfg
}

func defaultNodeConfig() node.Config {
	cfg := NodeDefaultConfig
	cfg.Name = clientIdentifier
	cfg.Version = params.VersionWithCommit(gitCommit, gitDate)
	cfg.HTTPModules = append(cfg.HTTPModules, "eth", "ftm", "dag", "web3")
	cfg.WSModules = append(cfg.WSModules, "eth", "ftm", "dag", "web3")
	cfg.IPCPath = "opera.ipc"
	cfg.DataDir = DefaultDataDir()
	return cfg
}

// dumpConfig is the dumpconfig command.
func dumpConfig(ctx *cli.Context) error {
	cfg := makeAllConfigs(ctx)
	comment := ""

	out, err := tomlSettings.Marshal(&cfg)
	if err != nil {
		return err
	}

	dump := os.Stdout
	if ctx.NArg() > 0 {
		dump, err = os.OpenFile(ctx.Args().Get(0), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer dump.Close()
	}
	dump.WriteString(comment)
	dump.Write(out)

	return nil
}

func checkConfig(ctx *cli.Context) error {
	_, err := mayMakeAllConfigs(ctx)
	return err
}
