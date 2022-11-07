package launcher

import (
	"os"
	"path"
	"strings"
	"time"

	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/batched"
	"github.com/Fantom-foundation/lachesis-base/kvdb/multidb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/table"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-opera/integration"
	"github.com/Fantom-foundation/go-opera/utils/compactdb"
)

func dbTransform(ctx *cli.Context) error {
	cfg := makeAllConfigs(ctx)

	tmpPath := path.Join(cfg.Node.DataDir, "tmp")
	integration.MakeDBDirs(tmpPath)
	_ = os.RemoveAll(tmpPath)
	defer os.RemoveAll(tmpPath)

	// get supported DB producers
	dbTypes := makeUncheckedCachedDBsProducers(path.Join(cfg.Node.DataDir, "chaindata"))

	byReq, err := readRoutes(cfg, dbTypes)
	if err != nil {
		log.Crit("Failed to read routes", "err", err)
	}
	byDB := separateIntoDBs(byReq)

	// weed out DBs which don't need transformation
	{
		for _, byReqOfDB := range byDB {
			match := true
			for _, e := range byReqOfDB {
				if e.Old != e.New {
					match = false
					break
				}
			}
			if match {
				for _, e := range byReqOfDB {
					delete(byReq, e.Req)
				}
			}
		}
	}
	if len(byReq) == 0 {
		log.Info("No DB transformation is needed")
		return nil
	}

	// check if new layout is contradictory
	for _, e0 := range byReq {
		for _, e1 := range byReq {
			if e0 == e1 {
				continue
			}
			if dbLocatorOf(e0.New) == dbLocatorOf(e1.New) && strings.HasPrefix(e0.New.Table, e1.New.Table) {
				log.Crit("New DB layout is contradictory", "db_type", e0.New.Type, "db_name", e0.New.Name,
					"req0", e0.Req, "req1", e1.Req, "table0", e0.New.Table, "table1", e1.New.Table)
			}
		}
	}

	// separate entries into inter-linked components
	byComponents := make([]map[string]dbMigrationEntry, 0)
	for componentI := 0; len(byReq) > 0; componentI++ {
		var someEntry dbMigrationEntry
		for _, e := range byReq {
			someEntry = e
			break
		}

		// DFS
		component := make(map[string]dbMigrationEntry)
		stack := make(dbMigrationEntries, 0)
		for pwalk := &someEntry; pwalk != nil; pwalk = stack.Pop() {
			if _, ok := component[pwalk.Req]; ok {
				continue
			}
			component[pwalk.Req] = *pwalk
			delete(byReq, pwalk.Req)
			for _, e := range byDB[dbLocatorOf(pwalk.Old)] {
				stack = append(stack, e)
			}
			for _, e := range byDB[dbLocatorOf(pwalk.New)] {
				stack = append(stack, e)
			}
		}
		byComponents = append(byComponents, component)
	}

	tmpDbTypes := makeUncheckedCachedDBsProducers(path.Join(cfg.Node.DataDir, "tmp"))
	for _, component := range byComponents {
		err := transformComponent(cfg.Node.DataDir, dbTypes, tmpDbTypes, component)
		if err != nil {
			log.Crit("Failed to transform component", "err", err)
		}
	}
	id := bigendian.Uint64ToBytes(uint64(time.Now().UnixNano()))
	for typ, producer := range dbTypes {
		err := clearDirtyFlags(id, producer)
		if err != nil {
			log.Crit("Failed to write clean FlushID", "type", typ, "err", err)
		}
	}

	log.Info("DB transformation is complete")

	return nil
}

type dbMigrationEntry struct {
	Req string
	Old multidb.Route
	New multidb.Route
}

type dbMigrationEntries []dbMigrationEntry

func (ee *dbMigrationEntries) Pop() *dbMigrationEntry {
	l := len(*ee)
	if l == 0 {
		return nil
	}
	res := &(*ee)[l-1]
	*ee = (*ee)[:l-1]
	return res
}

var dbLocatorOf = multidb.DBLocatorOf

func readRoutes(cfg *config, dbTypes map[multidb.TypeName]kvdb.FullDBProducer) (map[string]dbMigrationEntry, error) {
	router, err := multidb.NewProducer(dbTypes, cfg.DBs.Routing.Table, integration.TablesKey)
	if err != nil {
		return nil, err
	}
	byReq := make(map[string]dbMigrationEntry)

	for typ, producer := range dbTypes {
		for _, dbName := range producer.Names() {
			db, err := producer.OpenDB(dbName)
			if err != nil {
				log.Crit("DB opening error", "name", dbName, "err", err)
			}
			defer db.Close()
			tables, err := multidb.ReadTablesList(db, integration.TablesKey)
			if err != nil {
				log.Crit("Failed to read tables list", "name", dbName, "err", err)
			}
			for _, t := range tables {
				oldRoute := multidb.Route{
					Type:  typ,
					Name:  dbName,
					Table: t.Table,
				}
				newRoute := router.RouteOf(t.Req)
				newRoute.NoDrop = false
				byReq[t.Req] = dbMigrationEntry{
					Req: t.Req,
					New: newRoute,
					Old: oldRoute,
				}
			}
		}
	}
	return byReq, nil
}

func writeCleanTableRecords(dbTypes map[multidb.TypeName]kvdb.FullDBProducer, byReq map[string]dbMigrationEntry) error {
	records := make(map[multidb.DBLocator][]multidb.TableRecord, 0)
	for _, e := range byReq {
		records[dbLocatorOf(e.New)] = append(records[dbLocatorOf(e.New)], multidb.TableRecord{
			Req:   e.Req,
			Table: e.New.Table,
		})
	}
	written := make(map[multidb.DBLocator]bool)
	for _, e := range byReq {
		if written[dbLocatorOf(e.New)] {
			continue
		}
		written[dbLocatorOf(e.New)] = true

		db, err := dbTypes[e.New.Type].OpenDB(e.New.Name)
		if err != nil {
			return err
		}
		defer db.Close()
		err = multidb.WriteTablesList(db, integration.TablesKey, records[dbLocatorOf(e.New)])
		if err != nil {
			return err
		}
	}
	return nil
}

func interchangeableType(a_, b_ multidb.TypeName, types map[multidb.TypeName]kvdb.FullDBProducer) bool {
	for t_ := range types {
		a, b, t := string(a_), string(b_), string(t_)
		t = strings.TrimSuffix(t, "fsh")
		t = strings.TrimSuffix(t, "flg")
		t = strings.TrimSuffix(t, "drc")
		if strings.HasPrefix(a, t) && strings.HasPrefix(b, t) {
			return true
		}
	}
	return false
}

func transformComponent(datadir string, dbTypes, tmpDbTypes map[multidb.TypeName]kvdb.FullDBProducer, byReq map[string]dbMigrationEntry) error {
	byDB := separateIntoDBs(byReq)
	// if it can be transformed just by DB renaming
	if len(byDB) == 2 {
		oldDB := multidb.DBLocator{}
		newDB := multidb.DBLocator{}
		ok := true
		for _, e := range byReq {
			if len(oldDB.Type) == 0 {
				oldDB = dbLocatorOf(e.Old)
				newDB = dbLocatorOf(e.New)
			}
			if !interchangeableType(oldDB.Type, newDB.Type, dbTypes) || e.Old.Table != e.New.Table || e.New.Name != newDB.Name ||
				e.Old.Name != oldDB.Name || e.Old.Type != oldDB.Type || e.New.Type != newDB.Type {
				ok = false
				break
			}
		}
		if ok {
			oldPath := path.Join(datadir, "chaindata", string(oldDB.Type), oldDB.Name)
			newPath := path.Join(datadir, "chaindata", string(newDB.Type), newDB.Name)
			log.Info("Renaming DB", "old", oldPath, "new", newPath)
			return os.Rename(oldPath, newPath)
		}
	}

	toMove := make(map[multidb.DBLocator]bool)
	{
		const batchKeys = 100000
		keys := make([][]byte, 0, batchKeys)
		values := make([][]byte, 0, batchKeys)
		for _, e := range byReq {
			err := func() error {
				oldDB, err := dbTypes[e.Old.Type].OpenDB(e.Old.Name)
				if err != nil {
					return err
				}
				oldDB = batched.Wrap(oldDB)
				defer oldDB.Close()
				oldHumanName := path.Join(string(e.Old.Type), e.Old.Name)
				newDB, err := tmpDbTypes[e.New.Type].OpenDB(e.New.Name)
				if err != nil {
					return err
				}
				toMove[dbLocatorOf(e.New)] = true
				newDB = batched.Wrap(newDB)
				defer newDB.Close()
				newHumanName := path.Join("tmp", string(e.New.Type), e.New.Name)
				log.Info("Copying DB table", "req", e.Req, "old_db", oldHumanName, "old_table", e.Old.Table,
					"new_db", newHumanName, "new_table", e.New.Table)
				oldTable := table.New(oldDB, []byte(e.Old.Table))
				newTable := table.New(newDB, []byte(e.New.Table))
				it := oldTable.NewIterator(nil, nil)
				defer it.Release()

				for next := true; next; {
					for len(keys) < batchKeys {
						next = it.Next()
						if !next {
							break
						}
						keys = append(keys, common.CopyBytes(it.Key()))
						values = append(values, common.CopyBytes(it.Value()))
					}
					for i := 0; i < len(keys); i++ {
						err = newTable.Put(keys[i], values[i])
						if err != nil {
							return err
						}
					}
					keys = keys[:0]
					values = values[:0]
				}
				err = compactdb.Compact(newTable, newHumanName)
				if err != nil {
					log.Error("Database compaction failed", "err", err)
					return err
				}
				return nil
			}()
			if err != nil {
				return err
			}
		}
	}

	// finalize tmp DBs
	err := writeCleanTableRecords(tmpDbTypes, byReq)
	if err != nil {
		return err
	}

	// drop unused DBs
	dropped := make(map[multidb.DBLocator]bool)
	for _, e := range byReq {
		if dropped[dbLocatorOf(e.Old)] {
			continue
		}
		dropped[dbLocatorOf(e.Old)] = true
		log.Info("Dropping old DB", "db_type", e.Old.Type, "db_name", e.Old.Name)
		deletePath := path.Join(datadir, "chaindata", string(e.Old.Type), e.Old.Name)
		err := os.RemoveAll(deletePath)
		if err != nil {
			return err
		}
	}
	// move tmp DBs
	for e := range toMove {
		oldPath := path.Join(datadir, "tmp", string(e.Type), e.Name)
		newPath := path.Join(datadir, "chaindata", string(e.Type), e.Name)
		log.Info("Moving tmp DB to clean dir", "old", oldPath, "new", newPath)
		err := os.Rename(oldPath, newPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func separateIntoDBs(byReq map[string]dbMigrationEntry) map[multidb.DBLocator]map[string]dbMigrationEntry {
	byDB := make(map[multidb.DBLocator]map[string]dbMigrationEntry)
	for _, e := range byReq {
		if byDB[dbLocatorOf(e.Old)] == nil {
			byDB[dbLocatorOf(e.Old)] = make(map[string]dbMigrationEntry)
		}
		byDB[dbLocatorOf(e.Old)][e.Req] = e
		if byDB[dbLocatorOf(e.New)] == nil {
			byDB[dbLocatorOf(e.New)] = make(map[string]dbMigrationEntry)
		}
		byDB[dbLocatorOf(e.New)][e.Req] = e
	}
	return byDB
}
