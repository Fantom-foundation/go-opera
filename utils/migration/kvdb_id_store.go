package migration

import (
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/ethereum/go-ethereum/log"
)

// KvdbIDStore stores id
type KvdbIDStore struct {
	table kvdb.Store
	key   []byte
}

// NewKvdbIDStore constructor
func NewKvdbIDStore(table kvdb.Store) *KvdbIDStore {
	return &KvdbIDStore{
		table: table,
		key:   []byte("id"),
	}
}

// GetID is a getter
func (p *KvdbIDStore) GetID() string {
	id, err := p.table.Get(p.key)
	if err != nil {
		log.Crit("Failed to get key-value", "err", err)
	}

	if id == nil {
		return ""
	}
	return string(id)
}

// SetID is a setter
func (p *KvdbIDStore) SetID(id string) {
	err := p.table.Put(p.key, []byte(id))
	if err != nil {
		log.Crit("Failed to put key-value", "err", err)
	}
}
