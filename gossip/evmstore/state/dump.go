// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package state

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	//"github.com/ethereum/go-ethereum/rlp"
	//"github.com/ethereum/go-ethereum/trie"
	ecommon "github.com/ledgerwatch/erigon/common"
	"github.com/ledgerwatch/erigon/common/dbutils"

	"github.com/ledgerwatch/erigon/core/types/accounts"
	"github.com/ledgerwatch/erigon/crypto"

	"github.com/ledgerwatch/erigon-lib/kv"
)

// DumpConfig is a set of options to control what portions of the statewill be
// iterated and collected.
type DumpConfig struct {
	SkipCode          bool
	SkipStorage       bool
	OnlyWithAddresses bool
	Start             []byte
	Max               uint64
}

// DumpCollector interface which the state trie calls during iteration
type DumpCollector interface {
	// OnRoot is called with the state root
	OnRoot(common.Hash)
	// OnAccount is called once for each account in the trie
	OnAccount(common.Address, DumpAccount)
}

// DumpAccount represents an account in the state.
type DumpAccount struct {
	Balance   string                 `json:"balance"`
	Nonce     uint64                 `json:"nonce"`
	Root      hexutil.Bytes          `json:"root"`
	CodeHash  hexutil.Bytes          `json:"codeHash"`
	Code      hexutil.Bytes          `json:"code,omitempty"`
	Storage   map[common.Hash]string `json:"storage,omitempty"`
	Address   *common.Address        `json:"address,omitempty"` // Address only present in iterative (line-by-line) mode
	SecureKey hexutil.Bytes          `json:"key,omitempty"`     // If we don't have address, we can output the key

	Raw Account `json:"-"`
}

// Dump represents the full dump in a collected format, as one large map.
type Dump struct {
	Root     string                         `json:"root"`
	Accounts map[common.Address]DumpAccount `json:"accounts"`
}

// OnRoot implements DumpCollector interface
func (d *Dump) OnRoot(root common.Hash) {
	d.Root = fmt.Sprintf("%x", root)
}

// OnAccount implements DumpCollector interface
func (d *Dump) OnAccount(addr common.Address, account DumpAccount) {
	d.Accounts[addr] = account
}

// IteratorDump is an implementation for iterating over data.
type IteratorDump struct {
	Root     string                         `json:"root"`
	Accounts map[common.Address]DumpAccount `json:"accounts"`
	Next     []byte                         `json:"next,omitempty"` // nil if no more accounts
}

// OnRoot implements DumpCollector interface
func (d *IteratorDump) OnRoot(root common.Hash) {
	d.Root = fmt.Sprintf("%x", root)
}

// OnAccount implements DumpCollector interface
func (d *IteratorDump) OnAccount(addr common.Address, account DumpAccount) {
	d.Accounts[addr] = account
}

// iterativeDump is a DumpCollector-implementation which dumps output line-by-line iteratively.
type iterativeDump struct {
	*json.Encoder
}

// OnAccount implements DumpCollector interface
func (d iterativeDump) OnAccount(addr common.Address, account DumpAccount) {
	dumpAccount := &DumpAccount{
		Balance:   account.Balance,
		Nonce:     account.Nonce,
		Root:      account.Root,
		CodeHash:  account.CodeHash,
		Code:      account.Code,
		Storage:   account.Storage,
		SecureKey: account.SecureKey,
		Address:   nil,
	}
	if addr != (common.Address{}) {
		dumpAccount.Address = &addr
	}
	d.Encode(dumpAccount)
}

// OnRoot implements DumpCollector interface
func (d iterativeDump) OnRoot(root common.Hash) {
	d.Encode(struct {
		Root common.Hash `json:"root"`
	}{root})
}

// DumpToCollector iterates the state according to the given options and inserts
// the items into a collector for aggregation or serialization.
func (s *StateDB) DumpToCollector(dc DumpCollector, conf *DumpConfig, tx kv.RwTx) (nextKey []byte) {

	var emptyCodeHash = crypto.Keccak256Hash(nil)
	var emptyHash = common.Hash{}
	var accountList []*DumpAccount
	var addrList []common.Address
	var incarnationList []uint64

	// addr -> compositeKey -> value
	addrCompKeyValueMap := make(map[ecommon.Address]map[string][]byte)

	c, err := tx.Cursor(kv.PlainState)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	start := time.Now()
	storageCount, accountCount := 0, 0
	log.Info("PlainState dumping started")
	logEvery := time.NewTicker(20 * time.Second)
	defer logEvery.Stop()

	dc.OnRoot(emptyHash) // We do not calculate the root

	for k, v, e := c.First(); k != nil; k, v, e = c.Next() {
		if e != nil {
			panic(e)
		}

		select {
		case <-logEvery.C:
			log.Info("Plainstate dumping in progress", "at", k, "accounts", accountCount, "storage", storageCount,
				"elapsed", common.PrettyDuration(time.Since(start)))
		default:
		}

		switch {
		case len(k) == ecommon.AddressLength+ecommon.IncarnationLength+ecommon.HashLength:
			// handle account storage
			addr, _, _ := dbutils.PlainParseCompositeStorageKey(k)
			addrCompKeyValueMap[addr] = make(map[string][]byte)
			addrCompKeyValueMap[addr][string(k)] = v
			storageCount++
		case len(k) == ecommon.AddressLength:
			// handle non contract account
			acc := accounts.NewAccount()
			if err := acc.DecodeForStorage(v); err != nil {
				panic(err)
			}

			// TODO add incarnation
			account := DumpAccount{
				Balance:  acc.Balance.ToBig().String(),
				Nonce:    acc.Nonce,
				Root:     hexutil.Bytes(emptyHash[:]), // We cannot provide historical storage hash
				CodeHash: hexutil.Bytes(emptyCodeHash[:]),
				Storage:  make(map[common.Hash]string),
			}

			accountList = append(accountList, &account)
			addrList = append(addrList, common.BytesToAddress(k))
			incarnationList = append(incarnationList, acc.Incarnation)

			accountCount++
		default:
			panic("key is corrupt")
		}
	}

	for i, addr := range addrList {
		account := accountList[i]
		incarnation := incarnationList[i]
		storagePrefix := dbutils.PlainGenerateStoragePrefix(addr[:], incarnation)

		// handle contract code
		if incarnation > 0 {
			codeHash, err := s.db.GetOne(kv.PlainContractCode, storagePrefix)
			if err != nil {
				panic(err)
			}
			if codeHash != nil {
				account.CodeHash = codeHash
			} else {
				account.CodeHash = emptyCodeHash[:]
			}

			if codeHash != nil && !bytes.Equal(codeHash, emptyCodeHash[:]) {
				var code []byte
				if code, err = s.db.GetOne(kv.Code, codeHash); err != nil {
					panic(err)
				}
				account.Code = code
			}
		}

		// storage := address + inc + key -> value
		//TODO think about account.Root
		// handle contract storage
		if compositeKeyValMap, ok := addrCompKeyValueMap[ecommon.Address(addr)]; ok {
			for compositeKey, storageValue := range compositeKeyValMap {
				_, _, storageKey := dbutils.PlainParseCompositeStorageKey([]byte(compositeKey))
				account.Storage[common.Hash(storageKey)] = string(storageValue)
			}
		}

		dc.OnAccount(addr, *account)
	}

	log.Info("Plainstate dumping complete", "accounts", accountCount, "storageCount", storageCount,
		"elapsed", common.PrettyDuration(time.Since(start)))

	return
}

// open PlainState cursor for reading with wha

// dependign on len of key

/*
	if conf == nil {
		conf = new(DumpConfig)
	}
	var (
		missingPreimages int
		accounts         uint64
		start            = time.Now()
		logged           = time.Now()
	)
	log.Info("Trie dumping started", "root", s.trie.Hash())
	c.OnRoot(s.trie.Hash())

	it := trie.NewIterator(s.trie.NodeIterator(conf.Start))
	for it.Next() {
		var data Account
		if err := rlp.DecodeBytes(it.Value, &data); err != nil {
			panic(err)
		}
		account := DumpAccount{
			Balance:   data.Balance.String(),
			Nonce:     data.Nonce,
			Root:      data.Root[:],
			CodeHash:  data.CodeHash,
			SecureKey: it.Key,
		}
		addrBytes := s.trie.GetKey(it.Key)
		if addrBytes == nil {
			// Preimage missing
			missingPreimages++
			if conf.OnlyWithAddresses {
				continue
			}
			account.SecureKey = it.Key
		}
		addr := common.BytesToAddress(addrBytes)
		obj := newObject(s, addr, data)
		if !conf.SkipCode {
			account.Code = obj.Code()
		}
		if !conf.SkipStorage {
			account.Storage = make(map[common.Hash]string)
			storageIt := trie.NewIterator(obj.getTrie(s.db).NodeIterator(nil))
			for storageIt.Next() {
				_, content, _, err := rlp.Split(storageIt.Value)
				if err != nil {
					log.Error("Failed to decode the value returned by iterator", "error", err)
					continue
				}
				account.Storage[common.BytesToHash(s.trie.GetKey(storageIt.Key))] = common.Bytes2Hex(content)
			}
		}
		c.OnAccount(addr, account)
		accounts++
		if time.Since(logged) > 8*time.Second {
			log.Info("Trie dumping in progress", "at", it.Key, "accounts", accounts,
				"elapsed", common.PrettyDuration(time.Since(start)))
			logged = time.Now()
		}
		if conf.Max > 0 && accounts >= conf.Max {
			if it.Next() {
				nextKey = it.Key
			}
			break
		}
	}
	if missingPreimages > 0 {
		log.Warn("Dump incomplete due to missing preimages", "missing", missingPreimages)
	}
	log.Info("Trie dumping complete", "accounts", accounts,
		"elapsed", common.PrettyDuration(time.Since(start)))

	return nextKey
}
*/

// RawDump returns the entire state an a single large object
func (s *StateDB) RawDump(opts *DumpConfig) Dump {
	dump := &Dump{
		Accounts: make(map[common.Address]DumpAccount),
	}
	s.DumpToCollector(dump, opts, nil)
	return *dump
}

// Dump returns a JSON string representing the entire state as a single json-object
func (s *StateDB) Dump(opts *DumpConfig) []byte {
	dump := s.RawDump(opts)
	json, err := json.MarshalIndent(dump, "", "    ")
	if err != nil {
		fmt.Println("Dump err", err)
	}
	return json
}

// IterativeDump dumps out accounts as json-objects, delimited by linebreaks on stdout
func (s *StateDB) IterativeDump(opts *DumpConfig, output *json.Encoder) {
	s.DumpToCollector(iterativeDump{output}, opts, nil)
}

// IteratorDump dumps out a batch of accounts starts with the given start key
func (s *StateDB) IteratorDump(opts *DumpConfig) IteratorDump {
	iterator := &IteratorDump{
		Accounts: make(map[common.Address]DumpAccount),
	}
	iterator.Next = s.DumpToCollector(iterator, opts, nil)
	return *iterator
}
