package emitter

import (
	"errors"
	"sync"
	"time"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	notify "github.com/ethereum/go-ethereum/event"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/valkeystore"
	"github.com/Fantom-foundation/go-opera/vecmt"
)

var (
	ErrNotEnoughGasPower = errors.New("not enough gas power")
)

type (
	// External world
	External interface {
		sync.Locker
		Reader

		Check(e *inter.EventPayload, parents inter.Events) error
		Process(*inter.EventPayload) error
		Build(*inter.MutableEventPayload, func()) error
		DagIndex() *vecmt.Index

		IsBusy() bool
		IsSynced() bool
		PeersNum() int
	}

	// World is an emitter's environment
	World struct {
		Clock
		External
		TxPool   TxPool
		Signer   valkeystore.SignerI
		TxSigner types.Signer
	}

	// aliases for mock generator
	Signer   valkeystore.SignerI
	TxSigner types.Signer
)

// Reader is a callback for getting events from an external storage.
type Reader interface {
	GetLatestBlockIndex() idx.Block
	GetEpochValidators() (*pos.Validators, idx.Epoch)
	GetEvent(hash.Event) *inter.Event
	GetEventPayload(hash.Event) *inter.EventPayload
	GetLastEvent(epoch idx.Epoch, from idx.ValidatorID) *hash.Event
	GetHeads(idx.Epoch) hash.Events
	GetGenesisTime() inter.Timestamp
	GetRules() opera.Rules
}

type TxPool interface {
	// Has returns an indicator whether txpool has a transaction cached with the
	// given hash.
	Has(hash common.Hash) bool
	// Pending should return pending transactions.
	// The slice should be modifiable by the caller.
	Pending() (map[common.Address]types.Transactions, error)

	// SubscribeNewTxsNotify should return an event subscription of
	// NewTxsNotify and send events to the given channel.
	SubscribeNewTxsNotify(chan<- evmcore.NewTxsNotify) notify.Subscription
}

type (
	Clock interface {
		Now() time.Time
	}

	realClock struct {
	}
)

func (c realClock) Now() time.Time {
	return time.Now()
}
