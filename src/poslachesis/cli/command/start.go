package command

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/dgraph-io/badger"
	"github.com/spf13/cobra"

	"github.com/Fantom-foundation/go-lachesis/src/logger"
	lachesis "github.com/Fantom-foundation/go-lachesis/src/poslachesis"
)

// Start starts lachesis node.
var Start = &cobra.Command{
	Use:   "start",
	Short: "Starts lachesis node",
	RunE: func(cmd *cobra.Command, args []string) error {
		logLevel, err := cmd.Flags().GetString("log")
		if err != nil {
			return err
		}
		logger.GetLevel(logLevel)

		fakegen, err := cmd.Flags().GetString("fakegen")
		if err != nil {
			return err
		}

		num, total, err := parseFakeGen(fakegen)
		if err != nil {
			return err
		}

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

		net, keys := lachesis.FakeNet(total)
		conf := lachesis.DefaultConfig()
		conf.Net = net

		l := lachesis.New(db, "", keys[num], conf)
		l.Start()
		defer l.Stop()

		hosts, err := cmd.Flags().GetStringSlice("peer")
		if err != nil {
			return err
		}
		l.AddPeers(trim(hosts)...)

		dsn, err := cmd.Flags().GetString("dsn")
		if err != nil {
			return err
		}
		logger.SetDSN(dsn)

		wait()

		return nil
	},
}

func init() {
	Start.Flags().String("fakegen", "1/1", "use N/T format to use N-th key from T genesis keys")
	Start.Flags().String("db", "inmemory", "badger database dir")
	Start.Flags().StringSlice("peer", nil, "hosts of peers")
	Start.Flags().String("log", "info", "log level")
	Start.Flags().String("dsn", "", "Sentry client DSN")
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
	num = int(num64)

	total64, err := strconv.ParseUint(parts[1], 10, 64)
	total = int(total64)

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
