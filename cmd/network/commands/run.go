package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/lachesis"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//NewRunCmd returns the command that starts a Lachesis node
func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run",
		Short:   "Run node",
		PreRunE: loadConfig,
		RunE:    runLachesis,
	}

	AddRunFlags(cmd)

	return cmd
}

/*******************************************************************************
* RUN
*******************************************************************************/

func buildConfig() error {
	lachesisPort := 1337

	peersJSON := `[`

	for i := 0; i < config.NbNodes; i++ {
		nb := strconv.Itoa(i)

		lachesisPortStr := strconv.Itoa(lachesisPort + (i * 10))

		lachesisNode := exec.Command("lachesis", "keygen", "--pem=/tmp/lachesis_configs/.lachesis"+nb+"/priv_key.pem", "--pub=/tmp/lachesis_configs/.lachesis"+nb+"/key.pub")

		res, err := lachesisNode.CombinedOutput()
		if err != nil {
			log.Fatal(err, res)
		}

		pubKey, err := ioutil.ReadFile("/tmp/lachesis_configs/.lachesis" + nb + "/key.pub")
		if err != nil {
			log.Fatal(err, res)
		}

		peersJSON += `	{
		"NetAddr":"127.0.0.1:` + lachesisPortStr + `",
		"PubKeyHex":"` + string(pubKey) + `"
	},
`
	}

	peersJSON = peersJSON[:len(peersJSON)-2]
	peersJSON += `
]
`

	for i := 0; i < config.NbNodes; i++ {
		nb := strconv.Itoa(i)

		err := ioutil.WriteFile("/tmp/lachesis_configs/.lachesis"+nb+"/peers.json", []byte(peersJSON), 0644)
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

func sendTxs(lachesisNode *exec.Cmd, i int) {
	ticker := time.NewTicker(1 * time.Second)
	nb := strconv.Itoa(i)

	txNb := 0

	for range ticker.C {
		if txNb == config.SendTxs {
			ticker.Stop()

			break
		}

		network := exec.Command("network", "proxy", "--node="+nb, "--submit="+nb+"_"+strconv.Itoa(txNb))

		err := network.Run()
		if err != nil {
			continue
		}

		txNb++
	}
}

func runLachesis(cmd *cobra.Command, args []string) error {
	if err := os.RemoveAll("/tmp/lachesis_configs"); err != nil {
		log.Fatal(err)
	}

	if err := buildConfig(); err != nil {
		log.Fatal(err)
	}

	lachesisPort := 1337
	servicePort := 8080

	wg := sync.WaitGroup{}

	var processes = make([]*os.Process, config.NbNodes)

	for i := 0; i < config.NbNodes; i++ {
		wg.Add(1)

		go func(i int) {
			nb := strconv.Itoa(i)
			lachesisPortStr := strconv.Itoa(lachesisPort + (i * 10))
			proxyServPortStr := strconv.Itoa(lachesisPort + (i * 10) + 1)
			proxyCliPortStr := strconv.Itoa(lachesisPort + (i * 10) + 2)

			servicePort := strconv.Itoa(servicePort + i)

			defer wg.Done()

			lachesisNode := exec.Command("lachesis", "run", "-l=127.0.0.1:"+lachesisPortStr, "--datadir=/tmp/lachesis_configs/.lachesis"+nb, "--proxy-listen=127.0.0.1:"+proxyServPortStr, "--client-connect=127.0.0.1:"+proxyCliPortStr, "-s=127.0.0.1:"+servicePort, "--heartbeat="+config.Lachesis.NodeConfig.HeartbeatTimeout.String())
			err := lachesisNode.Start()

			if err != nil {
				log.Fatal(err)
			}

			fmt.Println("Running", i)

			if config.SendTxs > 0 {
				go sendTxs(lachesisNode, i)
			}

			processes[i] = lachesisNode.Process

			if err := lachesisNode.Wait(); err != nil {
				log.Fatal(err)
			}

			fmt.Println("Terminated", i)

		}(i)
	}

	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		for range c {
			for _, proc := range processes {
				if err := proc.Kill(); err != nil {
					panic(err)
				}
			}
		}
	}()

	wg.Wait()

	return nil
}

/*******************************************************************************
* CONFIG
*******************************************************************************/

//AddRunFlags adds flags to the Run command
func AddRunFlags(cmd *cobra.Command) {
	cmd.Flags().Int("nodes", config.NbNodes, "Amount of nodes to spawn")
	cmd.Flags().String("datadir", config.Lachesis.DataDir, "Top-level directory for configuration and data")
	cmd.Flags().String("log", config.Lachesis.LogLevel, "debug, info, warn, error, fatal, panic")
	cmd.Flags().Duration("heartbeat", config.Lachesis.NodeConfig.HeartbeatTimeout, "Time between gossips")

	cmd.Flags().Int64("sync-limit", config.Lachesis.NodeConfig.SyncLimit, "Max number of events for sync")
	cmd.Flags().Int("send-txs", config.SendTxs, "Send some random transactions")
}

func loadConfig(cmd *cobra.Command, args []string) error {

	err := bindFlagsLoadViper(cmd)
	if err != nil {
		return err
	}

	config, err = parseConfig()
	if err != nil {
		return err
	}

	config.Lachesis.Logger.Level = lachesis.LogLevel(config.Lachesis.LogLevel)
	config.Lachesis.NodeConfig.Logger = config.Lachesis.Logger

	return nil
}

//Bind all flags and read the config into viper
func bindFlagsLoadViper(cmd *cobra.Command) error {
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

//Retrieve the default environment configuration.
func parseConfig() (*CLIConfig, error) {
	conf := NewDefaultCLIConfig()
	err := viper.Unmarshal(conf)
	if err != nil {
		return nil, err
	}
	return conf, err
}
