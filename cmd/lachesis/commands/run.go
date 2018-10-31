package commands

import (
	"fmt"
	"time"

	"github.com/andrecronje/lachesis/src/dummy"
	"github.com/andrecronje/lachesis/src/lachesis"
	"github.com/andrecronje/lachesis/src/log"
	aproxy "github.com/andrecronje/lachesis/src/proxy"
	"github.com/andrecronje/lachesis/tester"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
)

//NewRunCmd returns the command that starts a Lachesis node
func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run node",
		RunE:  runLachesis,
	}
	AddRunFlags(cmd)
	return cmd
}

func runSingleLachesis(config *CLIConfig) error {
	config.Lachesis.Logger.Level = lachesis.LogLevel(config.Lachesis.LogLevel)
	config.Lachesis.NodeConfig.Logger = config.Lachesis.Logger

	lachesis_log.NewLocal(config.Lachesis.Logger, config.Lachesis.LogLevel)

	config.Lachesis.Logger.WithFields(logrus.Fields{
		"proxy-listen":   config.ProxyAddr,
		"client-connect": config.ClientAddr,
		"standalone":     config.Standalone,
		"service-only":   config.Lachesis.ServiceOnly,

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

	if !config.Standalone {
		p, err := aproxy.NewGrpcAppProxy(
			config.ProxyAddr,
			config.Lachesis.NodeConfig.HeartbeatTimeout,
			config.Lachesis.Logger,
		)

		if err != nil {
			config.Lachesis.Logger.Error("Cannot initialize socket AppProxy:", err)
			return nil
		}
		config.Lachesis.Proxy = p
	} else {
		p := dummy.NewInmemDummyApp(config.Lachesis.Logger)
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
		go func() {
			for {
				time.Sleep(10 * time.Second)
				ct := engine.Node.GetConsensusTransactionsCount()
				// 3 - number of notes in test; 10 - number of transactions sended at once
				if ct >= 3*10*config.Lachesis.TestN {
					time.Sleep(10 * time.Second)
					engine.Node.Shutdown()
					break
				}
			}
		}()
		go tester.PingNodesN(p.Sorted, p.ByPubKey, config.Lachesis.TestN, config.Lachesis.Logger)
	}

	engine.Run()

	return nil
}

//AddRunFlags adds flags to the Run command
func AddRunFlags(cmd *cobra.Command) {

	// local config here is used to set default values for the flags below
	config := NewDefaultCLIConfig()

	cmd.Flags().String("datadir", config.Lachesis.DataDir, "Top-level directory for configuration and data")
	cmd.Flags().String("log", config.Lachesis.LogLevel, "debug, info, warn, error, fatal, panic")

	// Network
	cmd.Flags().StringP("listen", "l", config.Lachesis.BindAddr, "Listen IP:Port for lachesis node")
	cmd.Flags().DurationP("timeout", "t", config.Lachesis.NodeConfig.TCPTimeout, "TCP Timeout")
	cmd.Flags().Int("max-pool", config.Lachesis.MaxPool, "Connection pool size max")

	// Proxy
	cmd.Flags().Bool("standalone", config.Standalone, "Do not create a proxy")
	cmd.Flags().Bool("service-only", config.Lachesis.ServiceOnly, "Only host the http service")
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

//Bind all flags and read the config into viper
func bindFlagsLoadViper(cmd *cobra.Command, config *CLIConfig) error {
	// cmd.Flags() includes flags from this command and all persistent flags from the parent
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}
	viper.SetConfigName("lachesis")              // name of config file (without extension)
	viper.AddConfigPath(config.Lachesis.DataDir) // search root directory
	// viper.AddConfigPath(filepath.Join(config.Lachesis.DataDir, "lachesis")) // search root directory /config
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		config.Lachesis.Logger.Debugf("Using config file: %s", viper.ConfigFileUsed())
	} else if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		config.Lachesis.Logger.Debugf("No config file found in: %s", config.Lachesis.DataDir)
	} else {
		return err
	}
	return nil
}

func logLevel(l string) logrus.Level {
	switch l {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	case "panic":
		return logrus.PanicLevel
	default:
		return logrus.DebugLevel
	}
}
