package gossip

import (
	"sort"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	notify "github.com/ethereum/go-ethereum/event"

	"github.com/Fantom-foundation/go-opera/evmcore"
)

// dummyTxPool is a fake, helper transaction pool for testing purposes
type dummyTxPool struct {
	lock   sync.RWMutex // Protects the transaction pool
	Signer types.Signer

	index   map[common.Hash]common.Address
	pending map[common.Address]types.Transactions
}

func newDummyTxPool() *dummyTxPool {
	p := new(dummyTxPool)
	p.Clear()
	return p
}

// AddRemotes appends a batch of transactions to the pool, and notifies any
// listeners if the addition channel is non nil
func (p *dummyTxPool) AddRemotes(txs []*types.Transaction) []error {
	p.lock.Lock()
	defer p.lock.Unlock()

	errs := make([]error, 0, len(txs))
	for _, tx := range txs {
		txid := tx.Hash()
		if _, ok := p.index[txid]; ok {
			continue
		}
		from, err := types.Sender(p.Signer, tx)
		if err == nil {
			p.index[txid] = from
			p.pending[from] = append(p.pending[from], tx)
			sort.Sort(types.TxByNonce(p.pending[from]))
		}
		errs = append(errs, err)
	}
	return errs
}

func (p *dummyTxPool) Count() int {
	return len(p.index)
}

func (p *dummyTxPool) Stats() (int, int) {
	return p.Count(), 0
}

func (p *dummyTxPool) Has(txid common.Hash) bool {
	p.lock.RLock()
	defer p.lock.RUnlock()

	_, ok := p.index[txid]
	return ok
}

// Pending returns all the transactions known to the pool
func (p *dummyTxPool) Pending(enforceTips bool) (map[common.Address]types.Transactions, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.clonePending(), nil
}

func (p *dummyTxPool) clonePending() map[common.Address]types.Transactions {
	clone := make(map[common.Address]types.Transactions, len(p.pending))

	for from, txs := range p.pending {
		tt := make(types.Transactions, len(txs))
		for i, tx := range txs {
			tt[i] = tx
		}
		clone[from] = tt
	}

	return clone
}

func (p *dummyTxPool) Clear() {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.index = make(map[common.Hash]common.Address)
	p.pending = make(map[common.Address]types.Transactions)
}

func (p *dummyTxPool) Delete(needle common.Hash) {
	p.lock.Lock()
	defer p.lock.Unlock()

	from, ok := p.index[needle]
	if !ok {
		return
	}

	txs := p.pending[from]
	if len(txs) < 2 {
		delete(p.pending, from)
	} else {
		notErased := make(types.Transactions, 0, len(txs)-1)
		for _, tx := range txs {
			if tx.Hash() != needle {
				notErased = append(notErased, tx)
			}
		}
		p.pending[from] = notErased
	}
	delete(p.index, needle)
}

func (p *dummyTxPool) AddLocals(txs []*types.Transaction) []error {
	panic("is not implemented")
	return nil
}

func (p *dummyTxPool) AddLocal(tx *types.Transaction) error {
	panic("is not implemented")
	return nil
}

func (p *dummyTxPool) Nonce(addr common.Address) uint64 {
	panic("is not implemented")
	return 0
}

func (p *dummyTxPool) Content() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	panic("is not implemented")
	return nil, nil
}

func (p *dummyTxPool) ContentFrom(addr common.Address) (types.Transactions, types.Transactions) {
	panic("is not implemented")
	return nil, nil
}

func (p *dummyTxPool) SubscribeNewTxsNotify(ch chan<- evmcore.NewTxsNotify) notify.Subscription {
	panic("is not implemented")
	return nil
}

func (p *dummyTxPool) Map() map[common.Hash]*types.Transaction {
	panic("is not implemented")
	return nil
}

func (p *dummyTxPool) Get(txid common.Hash) *types.Transaction {
	panic("is not implemented")
	return nil
}

func (p *dummyTxPool) OnlyNotExisting(txids []common.Hash) []common.Hash {
	panic("is not implemented")
	return nil
}

func (p *dummyTxPool) SampleHashes(max int) []common.Hash {
	panic("is not implemented")
	return nil
}
