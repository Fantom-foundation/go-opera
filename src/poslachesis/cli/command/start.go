package command

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/dgraph-io/badger"
	"github.com/spf13/cobra"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	lachesis "github.com/Fantom-foundation/go-lachesis/src/poslachesis"
)

// Start starts lachesis node.
var Start = &cobra.Command{
	Use:   "start",
	Short: "Starts lachesis node",
	RunE: func(cmd *cobra.Command, args []string) error {
		// log
		logLevel, err := cmd.Flags().GetString("log")
		if err != nil {
			return err
		}
		logger.GetLevel(logLevel)

		dsn, err := cmd.Flags().GetString("dsn")
		if err != nil {
			return err
		}
		logger.SetDSN(dsn)

		// db
		var db *badger.DB
		dbdir, err := cmd.Flags().GetString("db")
		if err != nil {
			return err
		}
		if dbdir != "inmemory" {
			db, err = ondiskDB(dbdir)
			if err != nil {
				return err
			}
		}

		// network
		var (
			conf = lachesis.DefaultConfig()
			key  *common.PrivateKey
		)
		netName, err := cmd.Flags().GetString("network")
		if err != nil {
			return err
		}
		keyPath, err := cmd.Flags().GetString("key")
		if err != nil {
			return err
		}
		switch {
		case strings.HasPrefix(netName, "fake:"):
			num, total, err := parseFakeGen(strings.Split(netName, ":")[1])
			if err != nil {
				return err
			}
			net, keys := lachesis.FakeNet(total)
			conf.Net = net
			key = keys[num]
		case netName == "test":
			conf.Net = lachesis.TestNet()
			rKey, err := readKey(keyPath)
			if err != nil {
				return nil
			}
			key = rKey

		case netName == "main":
			conf.Net = lachesis.MainNet()
			rKey, err := readKey(keyPath)
			if err != nil {
				return nil
			}
			key = rKey
		default:
			return fmt.Errorf("unknown network name: %s", netName)
		}

		// start
		l := lachesis.New(db, "", key, conf)
		l.Start()
		defer l.Stop()

		hosts, err := cmd.Flags().GetStringSlice("peer")
		if err != nil {
			return err
		}
		l.AddPeers(trim(hosts)...)

		wait()

		return nil
	},
}

func init() {
	Start.Flags().String("network", "fake:1/1", `one of: 
	- "fake:N/T" to use fakenet (N-th key from T genesis keys);
	- "test" to use testnet;
	- "main" to use mainnet;`)
	Start.Flags().String("db", "inmemory", "badger database dir")
	Start.Flags().StringSlice("peer", nil, "hosts of peers")
	Start.Flags().String("log", "info", "log level")
	Start.Flags().String("key", "", "private pem key path")
	Start.Flags().String("dsn", "", "Sentry client DSN")
}

func readKey(path string) (*common.PrivateKey, error) {
	keyFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer keyFile.Close()

	key, err := crypto.ReadPrivateKey(keyFile)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func parseFakeGen(s string) (num, total int, err error) {
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		err = fmt.Errorf("use %%d/%%d format")
		return
	}

	num64, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return
	}
	num = int(num64) - 1

	total64, err := strconv.ParseUint(parts[1], 10, 64)
	total = int(total64)

	if num64 < 1 || num64 > total64 {
		err = fmt.Errorf("key-num should be in range from 1 to total : <key-num>/<total>")
	}

	return
}

func ondiskDB(dir string) (*badger.DB, error) {
	opts := badger.DefaultOptions
	opts.Dir = dir
	opts.ValueDir = dir
	opts.SyncWrites = false

	return badger.Open(opts)
}

func wait() {
	done := make(chan os.Signal)
	signal.Notify(done, os.Interrupt, os.Kill)
	<-done
}

func trim(ss []string) []string {
	for i, s := range ss {
		ss[i] = strings.TrimSpace(s)
	}

	return ss
}
