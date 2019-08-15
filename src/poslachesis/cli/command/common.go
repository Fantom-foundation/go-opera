package command

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"go.etcd.io/bbolt"

	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/metrics"
	"github.com/Fantom-foundation/go-lachesis/src/poslachesis"
	"github.com/Fantom-foundation/go-lachesis/src/proxy"
)

func initCtrlProxy(cmd *cobra.Command) {
	cmd.Flags().String("addr", "localhost:55557", "node control net addr")
}

func makeCtrlProxy(cmd *cobra.Command) (proxy.NodeProxy, error) {
	addr, err := cmd.Flags().GetString("addr")
	if err != nil {
		return nil, err
	}

	grpcProxy, err := proxy.NewGrpcNodeProxy(addr, nil)
	if err != nil {
		return nil, err
	}

	return grpcProxy, nil
}

func dbProducer(cmd *cobra.Command) lachesis.DbProducer {
	dbdir, err := cmd.Flags().GetString("db")
	if err != nil {
		panic(err)
	}

	if dbdir == "inmemory" {
		return func(name string) kvdb.Database {
			return kvdb.NewMemDatabase()
		}
	}

	return func(name string) kvdb.Database {
		bdb, close, drop, err := openDB(dbdir, name)
		if err != nil {
			panic(err)
		}

		return kvdb.NewBoltDatabase(bdb, close, drop)
	}

}

func openDB(dir, name string) (db *bbolt.DB, close, drop func() error, err error) {
	err = os.MkdirAll(dir, 0600)
	if err != nil {
		return
	}

	f := filepath.Join(dir, name+".bolt")
	db, err = bbolt.Open(f, 0600, nil)
	if err != nil {
		return
	}

	stopWatcher := metrics.StartFileWatcher(name+"_db_file_size", f)

	close = func() error {
		stopWatcher()
		return db.Close()
	}

	drop = func() error {
		return os.Remove(f)
	}

	return
}
