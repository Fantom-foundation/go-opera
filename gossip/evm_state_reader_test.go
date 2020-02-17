package gossip

import (
	"math/big"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/app"
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/lachesis"
	"github.com/Fantom-foundation/go-lachesis/lachesis/genesis"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

func TestGetGenesisBlock(t *testing.T) {
	logger.SetTestMode(t)
	assertar := assert.New(t)

	net := lachesis.FakeNetConfig(genesis.FakeValidators(5, big.NewInt(0), pos.StakeToBalance(1)))
	addrWithStorage := net.Genesis.Alloc.Accounts.Addresses()[0]
	accountWithCode := net.Genesis.Alloc.Accounts[addrWithStorage]
	accountWithCode.Code = []byte{1, 2, 3}
	accountWithCode.Storage = make(map[common.Hash]common.Hash)
	accountWithCode.Storage[common.Hash{}] = common.BytesToHash(common.Big1.Bytes())
	net.Genesis.Alloc.Accounts[addrWithStorage] = accountWithCode

	app := app.NewMemStore()
	state, _, err := app.ApplyGenesis(&net, nil)
	if !assertar.NoError(err) {
		return
	}

	store := NewMemStore()
	genesisHash, stateHash, _, err := store.ApplyGenesis(&net, state)
	if !assertar.NoError(err) {
		return
	}

	assertar.NotEqual(common.Hash{}, genesisHash)
	assertar.NotEqual(common.Hash{}, stateHash)

	reader := EvmStateReader{
		store:    store,
		app:      app,
		engineMu: new(sync.RWMutex),
	}
	genesisBlock := reader.GetBlock(common.Hash(genesisHash), 0)

	assertar.Equal(common.Hash(genesisHash), genesisBlock.Hash)
	assertar.Equal(net.Genesis.Time, genesisBlock.Time)
	assertar.Empty(genesisBlock.Transactions)

	statedb, err := reader.StateAt(genesisBlock.Root)
	assertar.NoError(err)
	for addr, account := range net.Genesis.Alloc.Accounts {
		assertar.Equal(account.Balance.String(), statedb.GetBalance(addr).String())
		assertar.Equal(account.Code, statedb.GetCode(addr))
		if len(account.Storage) == 0 {
			assertar.Equal(common.Hash{}, statedb.GetState(addr, common.Hash{}))
		} else {
			for key, val := range account.Storage {
				assertar.Equal(val, statedb.GetState(addr, key))
			}
		}
	}
}

func TestGetBlock(t *testing.T) {
	logger.SetTestMode(t)
	assertar := assert.New(t)

	net := lachesis.FakeNetConfig(genesis.FakeValidators(5, big.NewInt(0), pos.StakeToBalance(1)))

	app := app.NewMemStore()
	state, _, err := app.ApplyGenesis(&net, nil)
	if !assertar.NoError(err) {
		return
	}

	store := NewMemStore()
	genesisHash, _, _, err := store.ApplyGenesis(&net, state)
	if !assertar.NoError(err) {
		return
	}

	txs := types.Transactions{}
	key, err := crypto.GenerateKey()
	assertar.NoError(err)
	for i := 0; i < 6; i++ {
		tx, err := types.SignTx(types.NewTransaction(uint64(i), common.Address{}, big.NewInt(100), 0, big.NewInt(1), nil), types.HomesteadSigner{}, key)
		assertar.NoError(err)
		txs = append(txs, tx)
	}

	event1 := inter.NewEvent()
	event2 := inter.NewEvent()
	event1.Transactions = txs[:1]
	event1.Seq = 1
	event2.Transactions = txs[1:]
	event1.Seq = 2
	block := inter.NewBlock(1, 123, event2.Hash(), genesisHash, hash.Events{event1.Hash(), event2.Hash()})
	block.SkippedTxs = []uint{0, 2, 4}

	store.SetEvent(event1)
	store.SetEvent(event2)
	store.SetBlock(block)

	reader := EvmStateReader{
		store:    store,
		engineMu: new(sync.RWMutex),
	}
	evmBlock := reader.GetDagBlock(block.Atropos, block.Index)

	assertar.Equal(uint64(block.Index), evmBlock.Number.Uint64())
	assertar.Equal(common.Hash(block.Atropos), evmBlock.Hash)
	assertar.Equal(common.Hash(genesisHash), evmBlock.ParentHash)
	assertar.Equal(block.Time, evmBlock.Time)
	assertar.Equal(len(txs)-len(block.SkippedTxs), evmBlock.Transactions.Len())
	assertar.Equal(txs[1].Hash(), evmBlock.Transactions[0].Hash())
	assertar.Equal(txs[3].Hash(), evmBlock.Transactions[1].Hash())
	assertar.Equal(txs[5].Hash(), evmBlock.Transactions[2].Hash())
}
