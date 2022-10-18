package state

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"

	ecommon "github.com/ledgerwatch/erigon/common"
	estate "github.com/ledgerwatch/erigon/core/state"

	"github.com/holiman/uint256"
	"github.com/ledgerwatch/erigon/core/types/accounts"
)

// WithStateReader sets erigon stateReader on StateDB
func (sdb *StateDB) WithStateReader(stateReader estate.StateReader) {
	sdb.stateReader = stateReader
}

// FinalizeTx should be called after every transaction.
func (sdb *StateDB) FinalizeTx(chainRules *params.Rules, stateWriter estate.StateWriter) error {
	/*
		for addr, bi := range sdb.balanceInc {
			if !bi.transferred {
				sdb.getStateObject(addr)
			}
		}
	*/
	for addr := range sdb.journal.dirties {
		so, exist := sdb.stateObjects[addr]
		if !exist {
			// ripeMD is 'touched' at block 1714175, in tx 0x1237f737031e40bcde4a8b7e717b2d15e3ecadfe49bb1bbc71ee9deb09c6fcf2
			// That tx goes out of gas, and although the notion of 'touched' does not exist there, the
			// touch-event will still be recorded in the journal. Since ripeMD is a special snowflake,
			// it will persist in the journal even though the journal is reverted. In this special circumstance,
			// it may exist in `sdb.journal.dirties` but not in `sdb.stateObjects`.
			// Thus, we can safely ignore it here
			continue
		}
		// TODO fto deal with first argument of updateAccount
		if err := updateAccount(true, stateWriter, addr, so, true); err != nil {
			return err
		}

		sdb.stateObjectsDirty[addr] = struct{}{}
	}
	// Invalidate journal because reverting across transactions is not allowed.
	sdb.clearJournalAndRefund()
	return nil
}

// TODO test it properly
func updateAccount(EIPEnabled bool, stateWriter estate.StateWriter, addr common.Address, stateObject *stateObject, isDirty bool) error {
	emptyRemoval := EIPEnabled && stateObject.empty()
	// TODO handle account removal

	if isDirty && !stateObject.suicided && !emptyRemoval {
		stateObject.deleted = false

		account := &stateObject.data
		isContract := !IsEmptyCodeHash(stateObject.CodeHash()) && !account.IsEmptyRoot()

		eAccount := transformStateAccount(account, isContract)

		if err := stateWriter.UpdateAccountData(ecommon.Address(addr), &eAccount, &eAccount); err != nil {
			return err
		}

		if err := stateObject.updateAccountStorage(stateWriter, eAccount.GetIncarnation()); err != nil {
			return err
		}

	}
	return nil
}

// updateAccountStorage writes cached storage modifications into the object's storage trie.
// writes storage to kv.Plainstate
// to make sure WriteAccountStorage writes storage correctly
// TODO make some tests
func (so *stateObject) updateAccountStorage(stateWriter estate.StateWriter, incarnation uint64) error {
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
	var account Account
	account.Nonce = eAccount.Nonce
	account.Balance = eAccount.Balance.ToBig()
	account.Root = common.Hash(eAccount.Root)
	account.CodeHash = eAccount.CodeHash.Bytes()
	return account
}
