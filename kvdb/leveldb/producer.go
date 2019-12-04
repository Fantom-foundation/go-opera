package leveldb

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Fantom-foundation/go-lachesis/kvdb"
)

type producer struct {
	datadir string
}

// NewProducer of level db.
func NewProducer(datadir string) kvdb.DbProducer {
	return &producer{
		datadir: datadir,
	}
}

// Names of existing databases.
func (p *producer) Names() []string {
	var names []string

	files, err := ioutil.ReadDir(p.datadir)
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		dirname := f.Name()
		if strings.HasSuffix(dirname, "-ldb") {
			name := strings.TrimSuffix(dirname, "-ldb")
			names = append(names, name)
		}
	}
	return names
}

// OpenDb or create db with name.
func (p *producer) OpenDb(name string) kvdb.KeyValueStore {
	dir := name + "-ldb"
	path := filepath.Join(p.datadir, dir)

	err := os.MkdirAll(path, 0700)
	if err != nil {
		panic(err)
	}

	var stopWatcher func()
	onClose := func() error {
		if stopWatcher != nil {
			stopWatcher()
		}
		return nil
	}
	onDrop := func() {
		err := os.RemoveAll(path)
		if err != nil {
			panic(err)
		}

	}

	db, err := New(path, 64, 0, "", onClose, onDrop)
	if err != nil {
		panic(err)
	}

	// TODO: dir watcher instead of file watcher needed.
	//stopWatcher = metrics.StartFileWatcher(name+"_db_file_size", f)

	return db
}
