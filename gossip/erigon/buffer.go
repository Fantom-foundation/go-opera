package erigon

import (
	"bytes"
	"context"
	"errors"
	"github.com/c2h5oh/datasize"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon/common"
)

type sortableBufferEntry struct {
	key   []byte
	value []byte
}

func newAppendBuffer(bufferOptimalSize datasize.ByteSize) *appendSortableBuffer {
	return &appendSortableBuffer{
		entries:     make(map[string][]byte),
		optimalSize: int(bufferOptimalSize.Bytes()),
	}
}

type appendSortableBuffer struct {
	entries     map[string][]byte
	size        int
	optimalSize int
	sortedBuf   []sortableBufferEntry
	comparator  kv.CmpFunc
}

func (b *appendSortableBuffer) Put(k, v []byte) error {
	if _, ok := b.entries[string(k)]; ok {
		return errors.New("dup entry found")
	}
	b.size += len(k)
	b.size += len(v)
	stored := make([]byte, len(v))
	stored = append(stored, v...)
	b.entries[string(k)] = stored
	return nil
}

func (b *appendSortableBuffer) SetComparator(cmp kv.CmpFunc) {
	b.comparator = cmp
}

func (b *appendSortableBuffer) Size() int {
	return b.size
}

func (b *appendSortableBuffer) Len() int {
	return len(b.entries)
}
func (b *appendSortableBuffer) Sort() {
	for i := range b.entries {
		b.sortedBuf = append(b.sortedBuf, sortableBufferEntry{key: []byte(i), value: b.entries[i]})
	}
	sort.Stable(b)
}

func (b *appendSortableBuffer) Less(i, j int) bool {
	if b.comparator != nil {
		return b.comparator(b.sortedBuf[i].key, b.sortedBuf[j].key, b.sortedBuf[i].value, b.sortedBuf[j].value) < 0
	}
	return bytes.Compare(b.sortedBuf[i].key, b.sortedBuf[j].key) < 0
}

func (b *appendSortableBuffer) Swap(i, j int) {
	b.sortedBuf[i], b.sortedBuf[j] = b.sortedBuf[j], b.sortedBuf[i]
}

func (b *appendSortableBuffer) Get(i int, keyBuf, valBuf []byte) ([]byte, []byte) {
	keyBuf = append(keyBuf, b.sortedBuf[i].key...)
	valBuf = append(valBuf, b.sortedBuf[i].value...)
	return keyBuf, valBuf
}
func (b *appendSortableBuffer) Reset() {
	b.sortedBuf = nil
	b.entries = make(map[string][]byte)
	b.size = 0
}

// writeIntoTable writes data into Erigon table from sorted buffer
func (b *appendSortableBuffer) writeIntoTable(db kv.RwDB, table string) error {

	// start erigon write tx
	tx, err := db.BeginRw(context.Background())
	if err != nil {
		return err
	}

	defer tx.Rollback()

	c, err := tx.RwCursor(table)
	if err != nil {
		return err
	}

	defer c.Close()

	log.Info("Iterate over sorted non duplicated buf key-value pairs", "and write them into", table)
	start := time.Now()
	records := 0
	for _, entry := range b.sortedBuf {
		if err := c.Append(entry.key, entry.value); err != nil {
			return err
		}
		records += 1
	}

	elapsed := common.PrettyDuration(time.Since(start))
	log.Info("Writing data into table from sorted buffer completed", "Number of records written", records, "elapsed", elapsed)

	return tx.Commit()
}
