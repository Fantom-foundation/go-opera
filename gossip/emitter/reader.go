package emitter

import (
	"errors"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/sfctype"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	notify "github.com/ethereum/go-ethereum/event"
)

var (
	NotEnoughGasPower = errors.New("not enough gas power")
)

// Reader is a callback for getting events from an external storage.
type Reader interface {
	GetValidatorProfiles() []sfctype.SfcStakerAndID
	GetEpoch() idx.Epoch
	GetValidators() *pos.Validators
	GetEpochValidators() (*pos.Validators, idx.Epoch)
	GetEvent(hash.Event) *inter.Event
	GetLastEvent(epoch idx.Epoch, from idx.ValidatorID) *hash.Event
	GetHeads(idx.Epoch) hash.Events
}

type txPool interface {
	// AddRemotes should add the given transactions to the pool.
	AddRemotes([]*types.Transaction) []error

	// Pending should return pending transactions.
	// The slice should be modifiable by the caller.
	Pending() (map[common.Address]types.Transactions, error)

	// SubscribeNewTxsNotify should return an event subscription of
	// NewTxsNotify and send events to the given channel.
	SubscribeNewTxsNotify(chan<- evmcore.NewTxsNotify) notify.Subscription
}
