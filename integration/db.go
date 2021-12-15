package integration

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/Fantom-foundation/go-opera/gossip"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/leveldb"
	"github.com/Fantom-foundation/lachesis-base/utils/cachescale"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

const (
	// metricsGatheringInterval specifies the interval to retrieve leveldb database
	// compaction, io and pause stats to report to the user.
	metricsGatheringInterval = 3 * time.Second
)

type DBProducerWithMetrics struct {
	kvdb.FlushableDBProducer
}

type DropableStoreWithMetrics struct {
	kvdb.DropableStore

	diskReadMeter  metrics.Meter // Meter for measuring the effective amount of data read
	diskWriteMeter metrics.Meter // Meter for measuring the effective amount of data written

	quitLock sync.Mutex      // Mutex protecting the quit channel access
	quitChan chan chan error // Quit channel to stop the metrics collection before closing the database

	log log.Logger // Contextual logger tracking the database path
}

func WrapDatabaseWithMetrics(db kvdb.FlushableDBProducer) kvdb.FlushableDBProducer {
	wrapper := &DBProducerWithMetrics{db}
	return wrapper
}

func WrapStoreWithMetrics(ds kvdb.DropableStore) *DropableStoreWithMetrics {
	wrapper := &DropableStoreWithMetrics{
		DropableStore: ds,
		quitChan:      make(chan chan error),
	}
	return wrapper
}

func (ds *DropableStoreWithMetrics) Close() error {
	ds.quitLock.Lock()
	defer ds.quitLock.Unlock()

	if ds.quitChan != nil {
		errc := make(chan error)
		ds.quitChan <- errc
		if err := <-errc; err != nil {
			ds.log.Error("Metrics collection failed", "err", err)
		}
		ds.quitChan = nil
	}
	return ds.DropableStore.Close()
}

func (ds *DropableStoreWithMetrics) meter(refresh time.Duration) {
	// Create storage for iostats.
	var iostats [2]float64

	var (
		errc chan error
		merr error
	)

	timer := time.NewTimer(refresh)
	defer timer.Stop()
	// Iterate ad infinitum and collect the stats
	for i := 1; errc == nil && merr == nil; i++ {
		// Retrieve the database iostats.
		ioStats, err := ds.Stat("leveldb.iostats")
		if err != nil {
			ds.log.Error("Failed to read database iostats", "err", err)
			merr = err
			continue
		}
		var nRead, nWrite float64
		parts := strings.Split(ioStats, " ")
		if len(parts) < 2 {
			ds.log.Error("Bad syntax of ioStats", "ioStats", ioStats)
			merr = fmt.Errorf("bad syntax of ioStats %s", ioStats)
			continue
		}
		if n, err := fmt.Sscanf(parts[0], "Read(MB):%f", &nRead); n != 1 || err != nil {
			ds.log.Error("Bad syntax of read entry", "entry", parts[0])
			merr = err
			continue
		}
		if n, err := fmt.Sscanf(parts[1], "Write(MB):%f", &nWrite); n != 1 || err != nil {
			log.Error("Bad syntax of write entry", "entry", parts[1])
			merr = err
			continue
		}
		if ds.diskReadMeter != nil {
			ds.diskReadMeter.Mark(int64((nRead - iostats[0]) * 1024 * 1024))
		}
		if ds.diskWriteMeter != nil {
			ds.diskWriteMeter.Mark(int64((nWrite - iostats[1]) * 1024 * 1024))
		}
		iostats[0], iostats[1] = nRead, nWrite

		// Sleep a bit, then repeat the stats collection
		select {
		case errc = <-ds.quitChan:
			// Quit requesting, stop hammering the database
		case <-timer.C:
			timer.Reset(refresh)
			// Timeout, gather a new set of stats
		}
	}
	if errc == nil {
		errc = <-ds.quitChan
	}
	errc <- merr
}

func (db *DBProducerWithMetrics) OpenDB(name string) (kvdb.DropableStore, error) {
	ds, err := db.FlushableDBProducer.OpenDB(name)
	if err != nil {
		return nil, err
	}
	dm := WrapStoreWithMetrics(ds)
	if strings.HasPrefix(name, "gossip-") || strings.HasPrefix(name, "lachesis-") {
		name = "epochs"
	}
	logger := log.New("database", name)
	dm.log = logger
	dm.diskReadMeter = metrics.GetOrRegisterMeter("opera/chaindata/"+name+"/disk/read", nil)
	dm.diskWriteMeter = metrics.GetOrRegisterMeter("opera/chaindata/"+name+"/disk/write", nil)

	// Start up the metrics gathering and return
	go dm.meter(metricsGatheringInterval)
	return dm, nil
}

func DBProducer(chaindataDir string, scale cachescale.Func) kvdb.IterableDBProducer {
	if chaindataDir == "inmemory" || chaindataDir == "" {
		chaindataDir, _ = ioutil.TempDir("", "opera-integration")
	}

	return leveldb.NewProducer(chaindataDir, func(name string) int {
		return dbCacheSize(name, scale.I)
	})
}

func CheckDBList(names []string) error {
	if len(names) == 0 {
		return nil
	}
	namesMap := make(map[string]bool)
	for _, name := range names {
		namesMap[name] = true
	}
	if !namesMap["gossip"] {
		return errors.New("gossip DB is not found")
	}
	if !namesMap["lachesis"] {
		return errors.New("lachesis DB is not found")
	}
	return nil
}

func dbCacheSize(name string, scale func(int) int) int {
	if name == "gossip" {
		return scale(128 * opt.MiB)
	}
	if name == "lachesis" {
		return scale(4 * opt.MiB)
	}
	if strings.HasPrefix(name, "lachesis-") {
		return scale(8 * opt.MiB)
	}
	if strings.HasPrefix(name, "gossip-") {
		return scale(8 * opt.MiB)
	}
	return scale(2 * opt.MiB)
}

func dropAllDBs(producer kvdb.IterableDBProducer) {
	names := producer.Names()
	for _, name := range names {
		db, err := producer.OpenDB(name)
		if err != nil {
			continue
		}
		_ = db.Close()
		db.Drop()
	}
}

func dropAllDBsIfInterrupted(rawProducer kvdb.IterableDBProducer) {
	names := rawProducer.Names()
	if len(names) == 0 {
		return
	}
	// if flushID is not written, then previous genesis processing attempt was interrupted
	for _, name := range names {
		db, err := rawProducer.OpenDB(name)
		if err != nil {
			return
		}
		flushID, err := db.Get(FlushIDKey)
		_ = db.Close()
		if flushID != nil || err != nil {
			return
		}
	}
	dropAllDBs(rawProducer)
}

type GossipStoreAdapter struct {
	*gossip.Store
}

func (g *GossipStoreAdapter) GetEvent(id hash.Event) dag.Event {
	e := g.Store.GetEvent(id)
	if e == nil {
		return nil
	}
	return e
}

type DummyFlushableProducer struct {
	kvdb.DBProducer
}

func (p *DummyFlushableProducer) NotFlushedSizeEst() int {
	return 0
}

func (p *DummyFlushableProducer) Flush(_ []byte) error {
	return nil
}
