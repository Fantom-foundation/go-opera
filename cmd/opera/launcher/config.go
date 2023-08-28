package launcher

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/Fantom-foundation/lachesis-base/abft"
	"github.com/Fantom-foundation/lachesis-base/utils/cachescale"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/params"
	"github.com/naoina/toml"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/urfave/cli/v2"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/flags"
	"github.com/Fantom-foundation/go-opera/gossip"
	"github.com/Fantom-foundation/go-opera/gossip/emitter"
	"github.com/Fantom-foundation/go-opera/gossip/gasprice"
	"github.com/Fantom-foundation/go-opera/integration"
	"github.com/Fantom-foundation/go-opera/integration/makefakegenesis"
	"github.com/Fantom-foundation/go-opera/opera/genesis"
	"github.com/Fantom-foundation/go-opera/opera/genesisstore"
	futils "github.com/Fantom-foundation/go-opera/utils"
	"github.com/Fantom-foundation/go-opera/utils/memory"
	"github.com/Fantom-foundation/go-opera/vecmt"
)

var (
	dumpConfigCommand = &cli.Command{
		Action:      dumpConfig,
		Name:        "dumpconfig",
		Usage:       "Show configuration values",
		ArgsUsage:   "",
		Flags:       append(nodeFlags, testFlags...),
		Category:    "MISCELLANEOUS COMMANDS",
		Description: `The dumpconfig command shows configuration values.`,
	}
	checkConfigCommand = &cli.Command{
		Action:      checkConfig,
		Name:        "checkconfig",
		Usage:       "Checks configuration file",
		ArgsUsage:   "",
		Flags:       append(nodeFlags, testFlags...),
		Category:    "MISCELLANEOUS COMMANDS",
		Description: `The checkconfig checks configuration file.`,
	}

	configFileFlag = &cli.StringFlag{
		Name:  "config",
		Usage: "TOML configuration file",
	}

	// DataDirFlag defines directory to store Lachesis state and user's wallets
	DataDirFlag = &flags.DirectoryFlag{
		Name:  "datadir",
		Usage: "Data directory for the databases and keystore",
		Value: flags.DirectoryString(DefaultDataDir()),
	}

	CacheFlag = &cli.IntFlag{
		Name:  "cache",
		Usage: "Megabytes of memory allocated to internal caching",
		Value: DefaultCacheSize,
	}
	// GenesisFlag specifies network genesis configuration
	GenesisFlag = &cli.StringFlag{
		Name:  "genesis",
		Usage: "'path to genesis file' - sets the network genesis configuration.",
	}
	ExperimentalGenesisFlag = &cli.BoolFlag{
		Name:  "genesis.allowExperimental",
		Usage: "Allow to use experimental genesis file.",
	}

	RPCGlobalGasCapFlag = &cli.Uint64Flag{
		Name:  "rpc.gascap",
		Usage: "Sets a cap on gas that can be used in ftm_call/estimateGas (0=infinite)",
		Value: gossip.DefaultConfig(cachescale.Identity).RPCGasCap,
	}
	RPCGlobalEVMTimeoutFlag = &cli.DurationFlag{
		Name:  "rpc.evmtimeout",
		Usage: "Sets a timeout used for eth_call (0=infinite)",
		Value: gossip.DefaultConfig(cachescale.Identity).RPCEVMTimeout,
	}
	RPCGlobalTxFeeCapFlag = &cli.Float64Flag{
		Name:  "rpc.txfeecap",
		Usage: "Sets a cap on transaction fee (in FTM) that can be sent via the RPC APIs (0 = no cap)",
		Value: gossip.DefaultConfig(cachescale.Identity).RPCTxFeeCap,
	}
	RPCGlobalTimeoutFlag = &cli.DurationFlag{
		Name:  "rpc.timeout",
		Usage: "Time limit for RPC calls execution",
		Value: gossip.DefaultConfig(cachescale.Identity).RPCTimeout,
	}

	SyncModeFlag = &cli.StringFlag{
		Name:  "syncmode",
		Usage: `Blockchain sync mode ("full" or "snap")`,
		Value: "full",
	}

	GCModeFlag = &cli.StringFlag{
		Name:  "gcmode",
		Usage: `Blockchain garbage collection mode ("light", "full", "archive")`,
		Value: "archive",
	}
	StateSchemeFlag = &cli.StringFlag{
		Name:  "state.scheme",
		Usage: "Scheme to use for storing ethereum state ('hash' or 'path')",
		Value: rawdb.HashScheme,
	}
	StateHistoryFlag = &cli.Uint64Flag{
		Name:  "history.state",
		Usage: "Number of recent blocks to retain state history for (default = 90,000 blocks, 0 = entire chain)",
		Value: ethconfig.Defaults.StateHistory,
	}
	ExitWhenAgeFlag = &cli.DurationFlag{
		Name:  "exitwhensynced.age",
		Usage: "Exits after synchronisation reaches the required age",
	}
	ExitWhenEpochFlag = &cli.Uint64Flag{
		Name:  "exitwhensynced.epoch",
		Usage: "Exits after synchronisation reaches the required epoch",
	}

	DBMigrationModeFlag = &cli.StringFlag{
		Name:  "db.migration.mode",
		Usage: "MultiDB migration mode ('reformat' or 'rebuild')",
	}
	DBPresetFlag = &cli.StringFlag{
		Name:  "db.preset",
		Usage: "DBs layout preset ('pbl-1' or 'ldb-1' or 'legacy-ldb' or 'legacy-pbl')",
	}
)

type GenesisTemplate struct {
	Name   string
	Header genesis.Header
	Hashes genesis.Hashes
}

const (
	// DefaultCacheSize is calculated as memory consumption in a worst case scenario with default configuration
	// Average memory consumption might be 3-5 times lower than the maximum
	DefaultCacheSize  = 3600
	ConstantCacheSize = 400
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
	Emitter       emitter.Config
	TxPool        evmcore.TxPoolConfig
	OperaStore    gossip.StoreConfig
	Lachesis      abft.Config
	LachesisStore abft.StoreConfig
	VectorClock   vecmt.IndexConfig
	DBs           integration.DBsConfig
}

func (c *config) AppConfigs() integration.Configs {
	return integration.Configs{
		Opera:         c.Opera,
		OperaStore:    c.OperaStore,
		Lachesis:      c.Lachesis,
		LachesisStore: c.LachesisStore,
		VectorClock:   c.VectorClock,
		DBs:           c.DBs,
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

func mayGetGenesisStore(ctx *cli.Context) *genesisstore.Store {
	switch {
	case ctx.IsSet(FakeNetFlag.Name):
		_, num, err := parseFakeGen(ctx.String(FakeNetFlag.Name))
		if err != nil {
			log.Crit("Invalid flag", "flag", FakeNetFlag.Name, "err", err)
		}
		return makefakegenesis.FakeGenesisStore(num, futils.ToFtm(1000000000), futils.ToFtm(5000000))
	case ctx.IsSet(GenesisFlag.Name):
		genesisPath := ctx.String(GenesisFlag.Name)

		f, err := os.Open(genesisPath)
		if err != nil {
			utils.Fatalf("Failed to open genesis file: %v", err)
		}
		genesisStore, genesisHashes, err := genesisstore.OpenGenesisStore(f)
		if err != nil {
			utils.Fatalf("Failed to read genesis file: %v", err)
		}

		// check if it's a trusted preset
		{
			g := genesisStore.Genesis()
			gHeader := genesis.Header{
				GenesisID:   g.GenesisID,
				NetworkID:   g.NetworkID,
				NetworkName: g.NetworkName,
			}
			for _, allowed := range AllowedOperaGenesis {
				if allowed.Hashes.Equal(genesisHashes) && allowed.Header.Equal(gHeader) {
					log.Info("Genesis file is a known preset", "name", allowed.Name)
					goto notExperimental
				}
			}
			if ctx.Bool(ExperimentalGenesisFlag.Name) {
				log.Warn("Genesis file doesn't refer to any trusted preset")
			} else {
				utils.Fatalf("Genesis file doesn't refer to any trusted preset. Enable experimental genesis with --genesis.allowExperimental")
			}
		notExperimental:
		}
		return genesisStore
	}
	return nil
}

func setBootnodes(ctx *cli.Context, urls []string, cfg *node.Config) {
	cfg.P2P.BootstrapNodesV5 = []*enode.Node{}
	for _, url := range urls {
		if url != "" {
			node, err := enode.Parse(enode.ValidSchemes, url)
			if err != nil {
				log.Error("Bootstrap URL invalid", "enode", url, "err", err)
				continue
			}
			cfg.P2P.BootstrapNodesV5 = append(cfg.P2P.BootstrapNodesV5, node)
		}
	}
	cfg.P2P.BootstrapNodes = cfg.P2P.BootstrapNodesV5
}

func setDataDir(ctx *cli.Context, cfg *node.Config) {
	defaultDataDir := DefaultDataDir()

	switch {
	case ctx.IsSet(DataDirFlag.Name):
		cfg.DataDir = ctx.String(DataDirFlag.Name)
	case ctx.IsSet(FakeNetFlag.Name):
		_, num, err := parseFakeGen(ctx.String(FakeNetFlag.Name))
		if err != nil {
			log.Crit("Invalid flag", "flag", FakeNetFlag.Name, "err", err)
		}
		cfg.DataDir = filepath.Join(defaultDataDir, fmt.Sprintf("fakenet-%d", num))
	}
}

func setGPO(ctx *cli.Context, cfg *gasprice.Config) {}

func setTxPool(ctx *cli.Context, cfg *evmcore.TxPoolConfig) {
	if ctx.IsSet(utils.TxPoolLocalsFlag.Name) {
		locals := strings.Split(ctx.String(utils.TxPoolLocalsFlag.Name), ",")
		for _, account := range locals {
			if trimmed := strings.TrimSpace(account); !common.IsHexAddress(trimmed) {
				utils.Fatalf("Invalid account in --txpool.locals: %s", trimmed)
			} else {
				cfg.Locals = append(cfg.Locals, common.HexToAddress(account))
			}
		}
	}
	if ctx.IsSet(utils.TxPoolNoLocalsFlag.Name) {
		cfg.NoLocals = ctx.Bool(utils.TxPoolNoLocalsFlag.Name)
	}
	if ctx.IsSet(utils.TxPoolJournalFlag.Name) {
		cfg.Journal = ctx.String(utils.TxPoolJournalFlag.Name)
	}
	if ctx.IsSet(utils.TxPoolRejournalFlag.Name) {
		cfg.Rejournal = ctx.Duration(utils.TxPoolRejournalFlag.Name)
	}
	if ctx.IsSet(utils.TxPoolPriceLimitFlag.Name) {
		cfg.PriceLimit = ctx.Uint64(utils.TxPoolPriceLimitFlag.Name)
	}
	if ctx.IsSet(utils.TxPoolPriceBumpFlag.Name) {
		cfg.PriceBump = ctx.Uint64(utils.TxPoolPriceBumpFlag.Name)
	}
	if ctx.IsSet(utils.TxPoolAccountSlotsFlag.Name) {
		cfg.AccountSlots = ctx.Uint64(utils.TxPoolAccountSlotsFlag.Name)
	}
	if ctx.IsSet(utils.TxPoolGlobalSlotsFlag.Name) {
		cfg.GlobalSlots = ctx.Uint64(utils.TxPoolGlobalSlotsFlag.Name)
	}
	if ctx.IsSet(utils.TxPoolAccountQueueFlag.Name) {
		cfg.AccountQueue = ctx.Uint64(utils.TxPoolAccountQueueFlag.Name)
	}
	if ctx.IsSet(utils.TxPoolGlobalQueueFlag.Name) {
		cfg.GlobalQueue = ctx.Uint64(utils.TxPoolGlobalQueueFlag.Name)
	}
	if ctx.IsSet(utils.TxPoolLifetimeFlag.Name) {
		cfg.Lifetime = ctx.Duration(utils.TxPoolLifetimeFlag.Name)
	}
}

func gossipConfigWithFlags(ctx *cli.Context, src gossip.Config) (gossip.Config, error) {
	cfg := src

	setGPO(ctx, &cfg.GPO)

	if ctx.IsSet(RPCGlobalGasCapFlag.Name) {
		cfg.RPCGasCap = ctx.Uint64(RPCGlobalGasCapFlag.Name)
	}
	if ctx.IsSet(RPCGlobalEVMTimeoutFlag.Name) {
		cfg.RPCEVMTimeout = ctx.Duration(RPCGlobalEVMTimeoutFlag.Name)
	}
	if ctx.IsSet(RPCGlobalTxFeeCapFlag.Name) {
		cfg.RPCTxFeeCap = ctx.Float64(RPCGlobalTxFeeCapFlag.Name)
	}
	if ctx.IsSet(RPCGlobalTimeoutFlag.Name) {
		cfg.RPCTimeout = ctx.Duration(RPCGlobalTimeoutFlag.Name)
	}
	if ctx.IsSet(SyncModeFlag.Name) {
		if syncmode := ctx.String(SyncModeFlag.Name); syncmode != "full" && syncmode != "snap" {
			utils.Fatalf("--%s must be either 'full' or 'snap'", SyncModeFlag.Name)
		}
		cfg.AllowSnapsync = ctx.String(SyncModeFlag.Name) == "snap"
	}

	return cfg, nil
}

func gossipStoreConfigWithFlags(ctx *cli.Context, src gossip.StoreConfig) (gossip.StoreConfig, error) {
	cfg := src
	if ctx.IsSet(utils.GCModeFlag.Name) {
		if gcmode := ctx.String(utils.GCModeFlag.Name); gcmode != "light" && gcmode != "full" && gcmode != "archive" {
			utils.Fatalf("--%s must be 'light', 'full' or 'archive'", GCModeFlag.Name)
		}
		cfg.EVM.Cache.TrieDirtyDisabled = ctx.String(utils.GCModeFlag.Name) == "archive"
		cfg.EVM.Cache.GreedyGC = ctx.String(utils.GCModeFlag.Name) == "full"
	}
	return cfg, nil
}

func setStateSchemaConfig(ctx *cli.Context, src config) config {
	cfg := src
	schema, err := parseStateScheme(ctx, cfg)
	if err != nil {
		utils.Fatalf("%v", err)
	}
	cfg.OperaStore.EVM.Cache.StateScheme = schema
	if ctx.IsSet(StateHistoryFlag.Name) {
		cfg.OperaStore.EVM.Cache.StateHistory = ctx.Uint64(StateHistoryFlag.Name)
	}
	return cfg
}

func parseStateScheme(ctx *cli.Context, cfg config) (string, error) {
	stored := string(futils.FileGet(path.Join(cfg.Node.DataDir, "chaindata", "schema")))
	if !ctx.IsSet(StateSchemeFlag.Name) {
		if stored == "" {
			// use default scheme for empty database, flip it when
			// path mode is chosen as default
			log.Info("State schema set to default", "scheme", "hash")
			return rawdb.HashScheme, nil
		}
		log.Info("State scheme set to already existing", "scheme", stored)
		return stored, nil // reuse scheme of persistent scheme
	}
	scheme := ctx.String(StateSchemeFlag.Name)
	if stored == "" || scheme == stored {
		log.Info("State scheme set by user", "scheme", scheme)
		return scheme, nil
	}
	return "", fmt.Errorf("incompatible state scheme, stored: %s, provided: %s", string(stored), scheme)
}

func setDBConfig(ctx *cli.Context, cfg integration.DBsConfig, cacheRatio cachescale.Func) integration.DBsConfig {
	if ctx.IsSet(DBPresetFlag.Name) {
		preset := ctx.String(DBPresetFlag.Name)
		cfg = setDBConfigStr(cfg, cacheRatio, preset)
	}
	if ctx.IsSet(DBMigrationModeFlag.Name) {
		cfg.MigrationMode = ctx.String(DBMigrationModeFlag.Name)
	}
	return cfg
}

func setDBConfigStr(cfg integration.DBsConfig, cacheRatio cachescale.Func, preset string) integration.DBsConfig {
	switch preset {
	case "pbl-1":
		cfg = integration.Pbl1DBsConfig(cacheRatio.U64, uint64(utils.MakeDatabaseHandles(0)))
	case "ldb-1":
		cfg = integration.Ldb1DBsConfig(cacheRatio.U64, uint64(utils.MakeDatabaseHandles(0)))
	case "legacy-ldb":
		cfg = integration.LdbLegacyDBsConfig(cacheRatio.U64, uint64(utils.MakeDatabaseHandles(0)))
	case "legacy-pbl":
		cfg = integration.PblLegacyDBsConfig(cacheRatio.U64, uint64(utils.MakeDatabaseHandles(0)))
	default:
		utils.Fatalf("--%s must be 'pbl-1', 'ldb-1', 'legacy-pbl' or 'legacy-ldb'", DBPresetFlag.Name)
	}
	// sanity check
	if preset != reversePresetName(cfg.Routing) {
		log.Error("Preset name cannot be reversed")
	}
	return cfg
}

func reversePresetName(cfg integration.RoutingConfig) string {
	pbl1 := integration.Pbl1RoutingConfig()
	ldb1 := integration.Ldb1RoutingConfig()
	ldbLegacy := integration.LdbLegacyRoutingConfig()
	pblLegacy := integration.PblLegacyRoutingConfig()
	if cfg.Equal(pbl1) {
		return "pbl-1"
	}
	if cfg.Equal(ldb1) {
		return "ldb-1"
	}
	if cfg.Equal(ldbLegacy) {
		return "legacy-ldb"
	}
	if cfg.Equal(pblLegacy) {
		return "legacy-pbl"
	}
	return ""
}

func memorizeDBPreset(cfg *config) {
	preset := reversePresetName(cfg.DBs.Routing)
	pPath := path.Join(cfg.Node.DataDir, "chaindata", "preset")
	if len(preset) != 0 {
		futils.FilePut(pPath, []byte(preset), true)
	} else {
		_ = os.Remove(pPath)
	}
}

func memorizeStateScheme(cfg *config) {
	schema := cfg.OperaStore.EVM.Cache.StateScheme
	pPath := path.Join(cfg.Node.DataDir, "chaindata", "schema")
	if len(schema) != 0 {
		futils.FilePut(pPath, []byte(schema), true)
	} else {
		_ = os.Remove(pPath)
	}
}

func setDBConfigDefault(cfg config, cacheRatio cachescale.Func) config {
	if len(cfg.DBs.Routing.Table) == 0 && len(cfg.DBs.GenesisCache.Table) == 0 && len(cfg.DBs.RuntimeCache.Table) == 0 {
		// Substitute memorized db preset from datadir, unless already set
		datadirPreset := futils.FileGet(path.Join(cfg.Node.DataDir, "chaindata", "preset"))
		if len(datadirPreset) != 0 {
			cfg.DBs = setDBConfigStr(cfg.DBs, cacheRatio, string(datadirPreset))
		}
	}
	// apply default for DB config if it wasn't touched by config file or flags, and there's no datadir's default value
	dbDefault := integration.DefaultDBsConfig(cacheRatio.U64, uint64(utils.MakeDatabaseHandles(0)))
	if len(cfg.DBs.Routing.Table) == 0 {
		cfg.DBs.Routing = dbDefault.Routing
	}
	if len(cfg.DBs.GenesisCache.Table) == 0 {
		cfg.DBs.GenesisCache = dbDefault.GenesisCache
	}
	if len(cfg.DBs.RuntimeCache.Table) == 0 {
		cfg.DBs.RuntimeCache = dbDefault.RuntimeCache
	}
	return cfg
}

func nodeConfigWithFlags(ctx *cli.Context, cfg node.Config) node.Config {
	utils.SetNodeConfig(ctx, &cfg)

	setDataDir(ctx, &cfg)
	return cfg
}

func cacheScaler(ctx *cli.Context) cachescale.Func {
	targetCache := ctx.Int(CacheFlag.Name)
	baseSize := DefaultCacheSize
	totalMemory := int(memory.TotalMemory() / opt.MiB)
	maxCache := totalMemory * 3 / 5
	if maxCache < baseSize {
		maxCache = baseSize
	}
	if !ctx.IsSet(CacheFlag.Name) {
		recommendedCache := totalMemory / 2
		if recommendedCache > baseSize {
			log.Warn(fmt.Sprintf("Please add '--%s %d' flag to allocate more cache for Opera. Total memory is %d MB.", CacheFlag.Name, recommendedCache, totalMemory))
		}
		return cachescale.Identity
	}
	if targetCache < baseSize {
		log.Crit("Invalid flag", "flag", CacheFlag.Name, "err", fmt.Sprintf("minimum cache size is %d MB", baseSize))
	}
	if totalMemory != 0 && targetCache > maxCache {
		log.Warn(fmt.Sprintf("Requested cache size exceeds 60%% of available memory. Reducing cache size to %d MB.", maxCache))
		targetCache = maxCache
	}

	return cachescale.Ratio{
		Base:   uint64(baseSize - ConstantCacheSize),
		Target: uint64(targetCache - ConstantCacheSize),
	}
}

func mayMakeAllConfigs(ctx *cli.Context) (*config, error) {
	// Defaults (low priority)
	cacheRatio := cacheScaler(ctx)
	cfg := config{
		Node:          defaultNodeConfig(),
		Opera:         gossip.DefaultConfig(cacheRatio),
		Emitter:       emitter.DefaultConfig(),
		TxPool:        evmcore.DefaultTxPoolConfig,
		OperaStore:    gossip.DefaultStoreConfig(cacheRatio),
		Lachesis:      abft.DefaultConfig(),
		LachesisStore: abft.DefaultStoreConfig(cacheRatio),
		VectorClock:   vecmt.DefaultConfig(cacheRatio),
	}

	if ctx.IsSet(FakeNetFlag.Name) {
		_, num, err := parseFakeGen(ctx.String(FakeNetFlag.Name))
		if err != nil {
			return nil, fmt.Errorf("invalid fakenet flag")
		}
		cfg.Emitter = emitter.FakeConfig(num)
		setBootnodes(ctx, []string{}, &cfg.Node)
	} else {
		// "asDefault" means set network defaults
		cfg.Node.P2P.BootstrapNodes = asDefault
		cfg.Node.P2P.BootstrapNodesV5 = asDefault
	}

	// Load config file (medium priority)
	if file := ctx.String(configFileFlag.Name); file != "" {
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
	cfg.OperaStore, err = gossipStoreConfigWithFlags(ctx, cfg.OperaStore)
	if err != nil {
		return nil, err
	}
	cfg.Node = nodeConfigWithFlags(ctx, cfg.Node)
	cfg.DBs = setDBConfig(ctx, cfg.DBs, cacheRatio)
	cfg = setStateSchemaConfig(ctx, cfg)

	err = setValidator(ctx, &cfg.Emitter)
	if err != nil {
		return nil, err
	}
	if cfg.Emitter.Validator.ID != 0 && len(cfg.Emitter.PrevEmittedEventFile.Path) == 0 {
		cfg.Emitter.PrevEmittedEventFile.Path = cfg.Node.ResolvePath(path.Join("emitter", fmt.Sprintf("last-%d", cfg.Emitter.Validator.ID)))
	}
	setTxPool(ctx, &cfg.TxPool)

	// Process DBs defaults in the end because they are applied only in absence of config or flags
	cfg = setDBConfigDefault(cfg, cacheRatio)

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
	cfg.HTTPModules = append(cfg.HTTPModules, "eth", "ftm", "dag", "abft", "web3")
	cfg.WSModules = append(cfg.WSModules, "eth", "ftm", "dag", "abft", "web3")
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
