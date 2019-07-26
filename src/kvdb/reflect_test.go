package kvdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStore struct {
	physicalDB Database

	table struct {
		T0 Database `table:"-"`
		T1 Database `table:"pre1_"`
		T2 Database `table:"pre2_"`
	}

	cache struct {
		C0 Database
		C1 Database `cache:"-"`
		C2 Database `cache:"-"`
	}
}

func TestMigrateTables(t *testing.T) {
	assertar := assert.New(t)

	s := &testStore{
		physicalDB: NewMemDatabase(),
	}

	if !assertar.NotNil(s.physicalDB) ||
		!assertar.Nil(s.table.T0) ||
		!assertar.Nil(s.table.T1) ||
		!assertar.Nil(s.table.T2) {
		return
	}

	MigrateTables(&s.table, s.physicalDB)

	if !assertar.NotNil(s.physicalDB) ||
		!assertar.Nil(s.table.T0) ||
		!assertar.NotNil(s.table.T1) ||
		!assertar.NotNil(s.table.T2) {
		return
	}

	MigrateTables(&s.table, nil)

	if !assertar.NotNil(s.physicalDB) ||
		!assertar.Nil(s.table.T0) ||
		!assertar.Nil(s.table.T1) ||
		!assertar.Nil(s.table.T2) {
		return
	}
}

func TestMigrateCaches(t *testing.T) {
	assertar := assert.New(t)

	s := &testStore{}

	if !assertar.Nil(s.cache.C0) ||
		!assertar.Nil(s.cache.C1) ||
		!assertar.Nil(s.cache.C2) {
		return
	}

	MigrateCaches(&s.cache, func() interface{} {
		return NewMemDatabase()
	})

	if !assertar.Nil(s.cache.C0) ||
		!assertar.NotNil(s.cache.C1) ||
		!assertar.NotNil(s.cache.C2) {
		return
	}

	MigrateCaches(&s.cache, nil)

	if !assertar.Nil(s.cache.C0) ||
		!assertar.Nil(s.cache.C1) ||
		!assertar.Nil(s.cache.C2) {
		return
	}
}
