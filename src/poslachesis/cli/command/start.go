package command

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/dgraph-io/badger"
	"github.com/spf13/cobra"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	lachesis "github.com/Fantom-foundation/go-lachesis/src/poslachesis"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/api"
	"github.com/Fantom-foundation/go-lachesis/src/posposet"
)

// Start starts lachesis node.
var Start = &cobra.Command{
	Use:   "start",
	Short: "Starts lachesis node",
	RunE: func(cmd *cobra.Command, args []string) error {
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

		net := lachesis.FakeNet(total)

		l := lachesis.New(db, "", crypto.GenerateFakeKey(num), nil)
		l.Start(net.Genesis)
		defer l.Stop()

		api.SetGenesisHash(posposet.GenesisHash(net.Genesis))

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
	Start.Flags().String("fakegen", "1/1", "use N/T format to use N-th key from T genesis keys")
	Start.Flags().String("db", "inmemory", "badger database dir")
	Start.Flags().StringSlice("peer", nil, "hosts of peers")
}

func parseFakeGen(s string) (num, total uint64, err error) {
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		err = fmt.Errorf("use %%d/%%d format")
		return
	}

	num, err = strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return
	}

	total, err = strconv.ParseUint(parts[1], 10, 64)
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
