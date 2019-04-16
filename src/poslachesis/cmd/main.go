package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/dgraph-io/badger"
	"github.com/spf13/cobra"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/poslachesis"
)

func main() {
	app := &cobra.Command{
		Use:   os.Args[0],
		Short: "It runs lachesis node",
	}

	fakegen := app.Flags().String(
		"fakegen", "1/1", "use N/T format to use N-th key from T genesis keys")
	dbdir := app.Flags().String(
		"db", "inmemory", "badger database dir")
	hosts := app.Flags().StringSlice(
		"peer", nil, "hosts of peers")

	app.RunE = func(cmd *cobra.Command, args []string) error {

		num, total, err := parseFakeGen(*fakegen)
		if err != nil {
			return err
		}

		var db *badger.DB
		if *dbdir != "inmemory" {
			db, err = ondiskDB(*dbdir)
			if err != nil {
				return err
			}
		}

		net := lachesis.FakeNet(total)

		l := lachesis.New(db, "", crypto.GenerateFakeKey(num), nil)
		l.Start(net.Genesis)
		defer l.Stop()

		l.AddPeers(trim(hosts)...)

		wait()

		return nil
	}

	app.Execute()
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

func trim(ss *[]string) []string {
	if ss == nil {
		return nil
	}

	res := *ss
	for i, s := range res {
		res[i] = strings.TrimSpace(s)
	}

	return res
}
