package vecmt

import (
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/table"
	"github.com/Fantom-foundation/lachesis-base/utils/cachescale"
	"github.com/Fantom-foundation/lachesis-base/utils/wlru"
	"github.com/Fantom-foundation/lachesis-base/vecengine"
	"github.com/Fantom-foundation/lachesis-base/vecfc"
)

// IndexCacheConfig - config for cache sizes of Engine
type IndexCacheConfig struct {
	HighestBeforeTimeSize uint
}

// IndexConfig - Engine config (cache sizes)
type IndexConfig struct {
	Fc     vecfc.IndexConfig
	Caches IndexCacheConfig
}

// Index is a data to detect forkless-cause condition, calculate median timestamp, detect forks.
type Index struct {
	*vecfc.Index
	Base          *vecfc.Index
	baseCallbacks vecengine.Callbacks

	crit          func(error)
	validators    *pos.Validators
	validatorIdxs map[idx.ValidatorID]idx.Validator

	getEvent func(hash.Event) dag.Event

	vecDb kvdb.Store
	table struct {
		HighestBeforeTime kvdb.Store `table:"T"`
	}

	cache struct {
		HighestBeforeTime *wlru.Cache
	}

	cfg IndexConfig
}

// DefaultConfig returns default index config
func DefaultConfig(scale cachescale.Func) IndexConfig {
	return IndexConfig{
		Fc: vecfc.DefaultConfig(scale),
		Caches: IndexCacheConfig{
			HighestBeforeTimeSize: scale.U(160 * 1024),
		},
	}
}

// LiteConfig returns default index config for tests
func LiteConfig() IndexConfig {
	return IndexConfig{
		Fc: vecfc.LiteConfig(),
		Caches: IndexCacheConfig{
			HighestBeforeTimeSize: 4 * 1024,
		},
	}
}

// NewIndex creates Index instance.
func NewIndex(crit func(error), config IndexConfig) *Index {
	vi := &Index{
		cfg:  config,
		crit: crit,
	}
	engine := vecengine.NewIndex(crit, vi.GetEngineCallbacks())

	vi.Base = vecfc.NewIndexWithEngine(crit, config.Fc, engine)
	vi.Index = vi.Base
	vi.baseCallbacks = vi.Base.GetEngineCallbacks()
	vi.initCaches()

	return vi
}

func NewIndexWithBase(crit func(error), config IndexConfig, base *vecfc.Index) *Index {
	vi := &Index{
		Index:         base,
		Base:          base,
		baseCallbacks: base.GetEngineCallbacks(),
		cfg:           config,
		crit:          crit,
	}
	vi.initCaches()

	return vi
}

func (vi *Index) initCaches() {
	vi.cache.HighestBeforeTime, _ = wlru.New(vi.cfg.Caches.HighestBeforeTimeSize, int(vi.cfg.Caches.HighestBeforeTimeSize))
}

// Reset resets buffers.
func (vi *Index) Reset(validators *pos.Validators, db kvdb.Store, getEvent func(hash.Event) dag.Event) {
	vi.Base.Reset(validators, db, getEvent)
	vi.getEvent = getEvent
	vi.validators = validators
	vi.validatorIdxs = validators.Idxs()
	vi.onDropNotFlushed()

	table.MigrateTables(&vi.table, vi.vecDb)
}

func (vi *Index) GetEngineCallbacks() vecengine.Callbacks {
	return vecengine.Callbacks{
		GetHighestBefore: func(event hash.Event) vecengine.HighestBeforeI {
			return vi.GetHighestBefore(event)
		},
		GetLowestAfter: func(event hash.Event) vecengine.LowestAfterI {
			return vi.baseCallbacks.GetLowestAfter(event)
		},
		SetHighestBefore: func(event hash.Event, b vecengine.HighestBeforeI) {
			vi.SetHighestBefore(event, b.(*HighestBefore))
		},
		SetLowestAfter: func(event hash.Event, i vecengine.LowestAfterI) {
			vi.baseCallbacks.SetLowestAfter(event, i)
		},
		NewHighestBefore: func(size idx.Validator) vecengine.HighestBeforeI {
			return NewHighestBefore(size)
		},
		NewLowestAfter: func(size idx.Validator) vecengine.LowestAfterI {
			return vi.baseCallbacks.NewLowestAfter(size)
		},
		OnDbReset: func(db kvdb.Store) {
			vi.baseCallbacks.OnDbReset(db)
			vi.onDbReset(db)
		},
		OnDropNotFlushed: func() {
			vi.baseCallbacks.OnDropNotFlushed()
			vi.onDropNotFlushed()
		},
	}
}

func (vi *Index) onDbReset(db kvdb.Store) {
	vi.vecDb = db
}

func (vi *Index) onDropNotFlushed() {
	vi.cache.HighestBeforeTime.Purge()
}

// GetMergedHighestBefore returns HighestBefore vector clock without branches, where branches are merged into one
func (vi *Index) GetMergedHighestBefore(id hash.Event) *HighestBefore {
	return vi.Engine.GetMergedHighestBefore(id).(*HighestBefore)
}
