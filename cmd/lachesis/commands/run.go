package commands

import (
	"fmt"
	"github.com/andrecronje/lachesis/src/lachesis"
	aproxy "github.com/andrecronje/lachesis/src/proxy/app"
	"github.com/andrecronje/lachesis/tester"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/urfave/cli"
)

//AddRunFlags adds flags to the Run command
func AddRunFlags(cmd *cobra.Command) {

	cmd.Flags().String("datadir", config.Lachesis.DataDir, "Top-level directory for configuration and data")
	cmd.Flags().String("log", config.Lachesis.LogLevel, "debug, info, warn, error, fatal, panic")

	// Network
	cmd.Flags().StringP("listen", "l", config.Lachesis.BindAddr, "Listen IP:Port for lachesis node")
	cmd.Flags().DurationP("timeout", "t", config.Lachesis.NodeConfig.TCPTimeout, "TCP Timeout")
	cmd.Flags().Int("max-pool", config.Lachesis.MaxPool, "Connection pool size max")

	// Proxy
	cmd.Flags().Bool("inapp", config.Inapp, "Use an in-app proxy")
	cmd.Flags().StringP("proxy-listen", "p", config.ProxyAddr, "Listen IP:Port for lachesis proxy")
	cmd.Flags().StringP("client-connect", "c", config.ClientAddr, "IP:Port to connect to client")

	// Service
	cmd.Flags().StringP("service-listen", "s", config.Lachesis.ServiceAddr, "Listen IP:Port for HTTP service")

	// Store
	cmd.Flags().Bool("store", config.Lachesis.Store, "Use badgerDB instead of in-mem DB")
	cmd.Flags().Int("cache-size", config.Lachesis.NodeConfig.CacheSize, "Number of items in LRU caches")

	// Node configuration
	cmd.Flags().Duration("heartbeat", config.Lachesis.NodeConfig.HeartbeatTimeout, "Time between gossips")
	cmd.Flags().Int("sync-limit", config.Lachesis.NodeConfig.SyncLimit, "Max number of events for sync")

	// Test
	cmd.Flags().Bool("test", config.Lachesis.Test, "Enable testing (sends transactions to random nodes in the network)")
	cmd.Flags().Uint64("test_n", config.Lachesis.TestN, "Number of transactions to send")
}

//NewRunCmd returns the command that starts a Lachesis node
func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run",
		Short:   "Run node",
		PreRunE: logConfig,
		RunE:    runLachesis,
	}

	AddRunFlags(cmd)

	return cmd
}

func logConfig(cmd *cobra.Command, args []string) error {
	config.Lachesis.Logger.Level = lachesis.LogLevel(config.Lachesis.LogLevel)
	config.Lachesis.NodeConfig.Logger = config.Lachesis.Logger

	config.Lachesis.Logger.WithFields(logrus.Fields{
		"proxy-listen":   config.ProxyAddr,
		"client-connect": config.ClientAddr,
		"inapp":          config.Inapp,

		"lachesis.datadir":        config.Lachesis.DataDir,
		"lachesis.bindaddr":       config.Lachesis.BindAddr,
		"lachesis.service-listen": config.Lachesis.ServiceAddr,
		"lachesis.maxpool":        config.Lachesis.MaxPool,
		"lachesis.store":          config.Lachesis.Store,
		"lachesis.loadpeers":      config.Lachesis.LoadPeers,
		"lachesis.log":            config.Lachesis.LogLevel,

		"lachesis.node.heartbeat":  config.Lachesis.NodeConfig.HeartbeatTimeout,
		"lachesis.node.tcptimeout": config.Lachesis.NodeConfig.TCPTimeout,
		"lachesis.node.cachesize":  config.Lachesis.NodeConfig.CacheSize,
		"lachesis.node.synclimit":  config.Lachesis.NodeConfig.SyncLimit,
	}).Debug("RUN")

	return nil
}

func runLachesis(cmd *cobra.Command, args []string) error {
	if !config.Inapp {
		p, err := aproxy.NewSocketAppProxy(
			config.ClientAddr,
			config.ProxyAddr,
			config.Lachesis.NodeConfig.HeartbeatTimeout,
			config.Lachesis.Logger,
		)

		if err != nil {
			config.Lachesis.Logger.Error("Cannot initialize socket AppProxy:", err)
			return nil
		}

		config.Lachesis.Proxy = p
	}

	engine := lachesis.NewLachesis(&config.Lachesis)

	if err := engine.Init(); err != nil {
		config.Lachesis.Logger.Error("Cannot initialize engine:", err)
		return nil
	}

	if config.Lachesis.Test {
		p, err := engine.Store.Participants()
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("Failed to acquire participants: %s", err),
				1)
		}
		go tester.PingNodesN(p.Sorted, p.ByPubKey, config.Lachesis.TestN, config.Lachesis.ServiceAddr)
	}

	engine.Run()

	return nil
}
