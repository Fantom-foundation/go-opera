package state

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math/big"

	ecommon "github.com/ledgerwatch/erigon/common"
	estate "github.com/ledgerwatch/erigon/core/state"

	"github.com/holiman/uint256"
	"github.com/ledgerwatch/erigon/core/types/accounts"
	//"github.com/ethereum/go-ethereum/log"
)

// WithStateReader sets erigon stateReader on StateDB
func (sdb *StateDB) WithStateReader(stateReader estate.StateReader) {
	sdb.stateReader = stateReader
}

// CommitBlock finalizes the state by removing the self destructed objects
// and clears the journal as well as the refunds.
func (sdb *StateDB) CommitBlock(stateWriter estate.StateWriter) error {
	return sdb.makeWriteSet(stateWriter)
}

func (sdb *StateDB) makeWriteSet(stateWriter estate.StateWriter) error {
	for addr := range sdb.journal.dirties {
		sdb.stateObjectsDirty[addr] = struct{}{}
	}
	for addr, stateObject := range sdb.stateObjects {
		_, isDirty := sdb.stateObjectsDirty[addr]
		if err := updateAccount(stateWriter, addr, stateObject, isDirty); err != nil {
			return err
		}
		//log.Info("CommitBlock", "updateAccount(addr)", addr.Hex())
	}
	// Invalidate journal because reverting across transactions is not allowed.
	sdb.clearJournalAndRefund()
	return nil
}

// SoftFinalize should be called after every transaction.
func (sdb *StateDB) SoftFinalize(stateWriter estate.StateWriter) error {
	for addr := range sdb.journal.dirties {
		if _, exist := sdb.stateObjects[addr]; !exist {
			// ripeMD is 'touched' at block 1714175, in tx 0x1237f737031e40bcde4a8b7e717b2d15e3ecadfe49bb1bbc71ee9deb09c6fcf2
			// That tx goes out of gas, and although the notion of 'touched' does not exist there, the
			// touch-event will still be recorded in the journal. Since ripeMD is a special snowflake,
			// it will persist in the journal even though the journal is reverted. In this special circumstance,
			// it may exist in `sdb.journal.dirties` but not in `sdb.stateObjects`.
			// Thus, we can safely ignore it here
			continue
		}

		sdb.stateObjectsDirty[addr] = struct{}{}
		//log.Info("SoftFinalize", "stateObjectsDirty[addr]", addr.Hex())
	}

	// Invalidate journal because reverting across transactions is not allowed.
	sdb.clearJournalAndRefund()
	return nil
}

// TODO test it properly
func updateAccount(stateWriter estate.StateWriter, addr common.Address, stateObject *stateObject, isDirty bool) error {

	// transform state.Accoount to erigon Account
	account := &stateObject.data
	isContract := !IsEmptyCodeHash(stateObject.CodeHash()) && !account.IsEmptyRoot()

	eAccount := transformStateAccount(account, isContract)
	eAddr := ecommon.Address(addr)

	emptyRemoval := stateObject.empty()

	if stateObject.suicided || (isDirty && emptyRemoval) {
		if err := stateWriter.DeleteAccount(ecommon.Address(addr), &eAccount); err != nil {
			return err
		}
		stateObject.deleted = true
	}

	if isDirty && (stateObject.created || !stateObject.suicided) && !emptyRemoval {
		stateObject.deleted = false

		if stateObject.code != nil && stateObject.dirtyCode {
			if err := stateWriter.UpdateAccountCode(eAddr, eAccount.Incarnation, eAccount.CodeHash, stateObject.code); err != nil {
				return err
			}
		}

		if stateObject.created {
			// works only with change sets
			if err := stateWriter.CreateContract(eAddr); err != nil {
				return err
			}
		}

		if err := stateWriter.UpdateAccountData(eAddr, nil, &eAccount); err != nil {
			return err
		}

		if err := stateObject.updateTrie(stateWriter, eAccount.GetIncarnation()); err != nil {
			return err
		}

	}
	return nil

}

func printAccount(addr common.Address, stateObject *stateObject, isDirty bool) {
	emptyRemoval := stateObject.empty()
	if stateObject.suicided || (isDirty && emptyRemoval) {
		fmt.Printf("delete: %x\n", addr)
	}
	if isDirty && (stateObject.created || !stateObject.suicided) && !emptyRemoval {
		// Write any contract code associated with the state object
		if stateObject.code != nil && stateObject.dirtyCode {
			fmt.Printf("UpdateCode: %x,%x\n", addr, stateObject.CodeHash())
		}
		if stateObject.created {
			fmt.Printf("CreateContract: %x\n", addr)
		}
		stateObject.printTrie()
		if stateObject.data.Balance.IsUint64() {
			fmt.Printf("UpdateAccountData: %x, balance=%d, nonce=%d\n", addr, stateObject.data.Balance.Uint64(), stateObject.data.Nonce)
		} else {
			div := big.NewInt(1_000_000_000)
			fmt.Printf("UpdateAccountData: %x, balance=%d*%d, nonce=%d\n", addr, big.NewInt(0).Div(stateObject.data.Balance, div).Uint64(), div.Uint64(), stateObject.data.Nonce)
		}
	}
}

// updateTrie writes cached storage modifications into the object's storage trie.
// writes storage to kv.Plainstate
// to make sure WriteAccountStorage writes storage correctly
// TODO make some tests
func (so *stateObject) updateTrie(stateWriter estate.StateWriter, incarnation uint64) error {
	for key, value := range so.dirtyStorage {
		original := so.originStorage[key]
		so.originStorage[key] = value
		key := ecommon.Hash(key)
		eOriginal := uint256.NewInt(0).SetBytes(original.Bytes())
		value := uint256.NewInt(0).SetBytes(value.Bytes())
		if err := stateWriter.WriteAccountStorage(ecommon.Address(so.address), incarnation, &key, eOriginal, value); err != nil {
			return err
		}
	}
	return nil
}

func (so *stateObject) printTrie() {
	for key, value := range so.dirtyStorage {
		fmt.Printf("WriteAccountStorage: %x,%x,%s\n", so.address, key, value.Hex())
	}
}

// transformStateAccount transforms state.Account into erigon account representation (https://github.com/ledgerwatch/erigon/blob/devel/docs/programmers_guide/guide.md)
func transformStateAccount(account *Account, isContractAcc bool) accounts.Account {
	eAccount := accounts.NewAccount()
	eAccount.Initialised = true
	bal, overflow := uint256.FromBig(account.Balance)
	if overflow {
		panic("overflow occured while converting from account.Balance")
	}
	eAccount.Nonce = account.Nonce
	eAccount.Balance = *bal
	eAccount.Root = ecommon.Hash(account.Root)
	eAccount.CodeHash = ecommon.Hash(common.BytesToHash(account.CodeHash))
	if isContractAcc {
		eAccount.Incarnation = 1
	}

	return eAccount
}

// this is temporary solution, evaluate which account model to use
func transformErigonAccount(eAccount *accounts.Account) Account {
	if eAccount == nil {
		acc := accounts.NewAccount()
		eAccount = &acc
	}
	var account Account
	account.Nonce = eAccount.Nonce
	account.Balance = eAccount.Balance.ToBig()
	account.Root = common.Hash(eAccount.Root)
	account.CodeHash = eAccount.CodeHash.Bytes()
	return account
}
