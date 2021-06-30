package genesisstore

import (
	"crypto/sha256"
	"math/big"

	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/opera/genesis"
	"github.com/Fantom-foundation/go-opera/opera/genesis/gpos"
)

type (
	Metadata struct {
		Validators    gpos.Validators
		FirstEpoch    idx.Epoch
		Time          inter.Timestamp
		PrevEpochTime inter.Timestamp
		ExtraData     []byte
		DriverOwner   common.Address
		TotalSupply   *big.Int
	}
	Accounts struct {
		Raw kvdb.Iteratee
	}
	Storage struct {
		Raw kvdb.Iteratee
	}
	Delegations struct {
		Raw kvdb.Iteratee
	}
	Blocks struct {
		Raw kvdb.Iteratee
	}
)

func (s *Store) EvmAccounts() genesis.Accounts {
	return &Accounts{s.table.EvmAccounts}
}

func (s *Store) SetEvmAccount(addr common.Address, acc genesis.Account) {
	s.rlp.Set(s.table.EvmAccounts, addr.Bytes(), &acc)
}

func (s *Store) GetEvmAccount(addr common.Address) genesis.Account {
	w, ok := s.rlp.Get(s.table.EvmAccounts, addr.Bytes(), &genesis.Account{}).(*genesis.Account)
	if !ok {
		return genesis.Account{
			Code:    []byte{},
			Balance: new(big.Int),
			Nonce:   0,
		}
	}
	return *w
}

func (s *Store) EvmStorage() genesis.Storage {
	return &Storage{s.table.EvmStorage}
}

func (s *Store) SetEvmState(addr common.Address, key common.Hash, value common.Hash) {
	err := s.table.EvmStorage.Put(append(addr.Bytes(), key.Bytes()...), value.Bytes())
	if err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

func (s *Store) GetEvmState(addr common.Address, key common.Hash) common.Hash {
	valBytes, err := s.table.EvmStorage.Get(append(addr.Bytes(), key.Bytes()...))
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if len(valBytes) == 0 {
		return common.Hash{}
	}
	return common.BytesToHash(valBytes)
}

func (s *Store) Delegations() genesis.Delegations {
	return &Delegations{s.table.Delegations}
}

func (s *Store) SetDelegation(addr common.Address, toValidatorID idx.ValidatorID, delegation genesis.Delegation) {
	s.rlp.Set(s.table.Delegations, append(addr.Bytes(), toValidatorID.Bytes()...), &delegation)
}

func (s *Store) GetDelegation(addr common.Address, toValidatorID idx.ValidatorID) genesis.Delegation {
	w, ok := s.rlp.Get(s.table.Delegations, append(addr.Bytes(), toValidatorID.Bytes()...), &genesis.Delegation{}).(*genesis.Delegation)
	if !ok {
		return genesis.Delegation{
			Stake:              new(big.Int),
			Rewards:            new(big.Int),
			LockedStake:        new(big.Int),
			LockupFromEpoch:    0,
			LockupEndTime:      0,
			LockupDuration:     0,
			EarlyUnlockPenalty: new(big.Int),
		}
	}
	return *w
}

func (s *Store) Blocks() genesis.Blocks {
	return &Blocks{s.table.Blocks}
}

func (s *Store) SetBlock(index idx.Block, block genesis.Block) {
	s.rlp.Set(s.table.Blocks, index.Bytes(), &block)
}

func (s *Store) SetRawEvmItem(key, value []byte) {
	err := s.table.RawEvmItems.Put(key, value)
	if err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

func (s *Store) GetMetadata() Metadata {
	metadata := s.rlp.Get(s.table.Metadata, []byte("m"), &Metadata{}).(*Metadata)
	return *metadata
}

func (s *Store) SetMetadata(metadata Metadata) {
	s.rlp.Set(s.table.Metadata, []byte("m"), &metadata)
}

func (s *Store) GetRules() opera.Rules {
	cfg := s.rlp.Get(s.table.Rules, []byte("c"), &opera.Rules{}).(*opera.Rules)
	return *cfg
}

func (s *Store) SetRules(cfg opera.Rules) {
	s.rlp.Set(s.table.Rules, []byte("c"), &cfg)
}

func (s *Store) GetGenesis() opera.Genesis {
	meatadata := s.GetMetadata()
	return opera.Genesis{
		GenesisHeader: opera.GenesisHeader{
			Validators:    meatadata.Validators,
			FirstEpoch:    meatadata.FirstEpoch,
			PrevEpochTime: meatadata.PrevEpochTime,
			Time:          meatadata.Time,
			ExtraData:     meatadata.ExtraData,
			TotalSupply:   meatadata.TotalSupply,
			DriverOwner:   meatadata.DriverOwner,
			Rules:         s.GetRules(),
			Hash:          s.Hash,
		},
		Accounts:    s.EvmAccounts(),
		Storage:     s.EvmStorage(),
		Delegations: s.Delegations(),
		Blocks:      s.Blocks(),
		RawEvmItems: s.table.RawEvmItems,
	}
}

func (s *Accounts) ForEach(fn func(common.Address, genesis.Account)) {
	it := s.Raw.NewIterator(nil, nil)
	defer it.Release()
	for it.Next() {
		addr := common.BytesToAddress(it.Key())
		acc := genesis.Account{}
		err := rlp.DecodeBytes(it.Value(), &acc)
		if err != nil {
			log.Crit("Genesis accounts error", "err", err)
		}
		fn(addr, acc)
	}
}

func (s *Storage) ForEach(fn func(common.Address, common.Hash, common.Hash)) {
	it := s.Raw.NewIterator(nil, nil)
	defer it.Release()
	for it.Next() {
		addr := common.BytesToAddress(it.Key()[:20])
		key := common.BytesToHash(it.Key()[20:])
		val := common.BytesToHash(it.Value())
		fn(addr, key, val)
	}
}

func (s *Delegations) ForEach(fn func(common.Address, idx.ValidatorID, genesis.Delegation)) {
	it := s.Raw.NewIterator(nil, nil)
	defer it.Release()
	for it.Next() {
		addr := common.BytesToAddress(it.Key()[:20])
		to := idx.BytesToValidatorID(it.Key()[20:])
		delegation := genesis.Delegation{}
		err := rlp.DecodeBytes(it.Value(), &delegation)
		if err != nil {
			log.Crit("Genesis delegations error", "err", err)
		}
		fn(addr, to, delegation)
	}
}

func (s *Blocks) ForEach(fn func(idx.Block, genesis.Block)) {
	it := s.Raw.NewIterator(nil, nil)
	defer it.Release()
	for it.Next() {
		index := idx.BytesToBlock(it.Key())
		block := genesis.Block{}
		err := rlp.DecodeBytes(it.Value(), &block)
		if err != nil {
			log.Crit("Genesis blocks error", "err", err)
		}
		fn(index, block)
	}
}

func (s *Store) Hash() hash.Hash {
	hasher := sha256.New()
	it := s.db.NewIterator(nil, nil)
	defer it.Release()
	for it.Next() {
		k := it.Key()
		v := it.Value()
		hasher.Write(bigendian.Uint32ToBytes(uint32(len(k))))
		hasher.Write(k)
		hasher.Write(bigendian.Uint32ToBytes(uint32(len(v))))
		hasher.Write(v)
	}
	return hash.BytesToHash(hasher.Sum(nil))
}
