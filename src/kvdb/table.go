package kvdb

import (
	"reflect"
)

const tableTag = "table"

type table struct {
	db     Database
	prefix string
}

// NewTable returns a Database object that prefixes all keys with a given
// string.
func NewTable(db Database, prefix string) Database {
	return &table{
		db:     db,
		prefix: prefix,
	}
}

func MigrateTables(store interface{}, db Database, populate bool) {
	s, ok := store.(interface {
		Close()
		Open()
	})

	if !ok {
		return
	}

	value := reflect.ValueOf(s).Elem()
	for i := 0; i < value.NumField(); i++ {
		if tag := value.Type().Field(i).Tag.Get(tableTag); tag != "" && tag != "-" {
			v := reflect.ValueOf(nil)
			if populate {
				v = reflect.ValueOf(NewTable(db, tag))
			}
			value.Field(i).Set(v)
		}
	}
}

func (dt *table) Put(key []byte, value []byte) error {
	return dt.db.Put(append([]byte(dt.prefix), key...), value)
}

func (dt *table) Has(key []byte) (bool, error) {
	return dt.db.Has(append([]byte(dt.prefix), key...))
}

func (dt *table) Get(key []byte) ([]byte, error) {
	return dt.db.Get(append([]byte(dt.prefix), key...))
}

func (dt *table) Delete(key []byte) error {
	return dt.db.Delete(append([]byte(dt.prefix), key...))
}

func (dt *table) Close() {
	// Do nothing; don't close the underlying DB.
}
