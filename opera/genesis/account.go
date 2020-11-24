package genesis

import (
	"math/big"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-opera/inter"
)

// Accounts specifies the initial state that is part of the genesis block.
type (
	Accounts interface {
		ForEach(fn func(common.Address, Account))
	}
	Storage interface {
		ForEach(fn func(common.Address, common.Hash, common.Hash))
	}
	Delegations interface {
		ForEach(fn func(common.Address, idx.ValidatorID, Delegation))
	}
	Blocks interface {
		ForEach(fn func(idx.Block, Block))
	}

	Delegation struct {
		Stake   *big.Int
		Rewards *big.Int
	}
	// Account is an account in the state of the genesis block.
	Account struct {
		Code    []byte
		Balance *big.Int
		Nonce   uint64
	}

	Block struct {
		Time        inter.Timestamp
		Atropos     hash.Event
		Txs         types.Transactions
		InternalTxs types.Transactions
		Root        hash.Hash
		Receipts    []*types.ReceiptForStorage
	}
)
