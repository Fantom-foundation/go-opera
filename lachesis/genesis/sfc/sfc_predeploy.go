package sfc

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/lachesis/genesis/sfc/sfcpos"
	"github.com/Fantom-foundation/go-lachesis/utils"
)

// SFC contract first implementation bin code
// Must be compiled with bin-runtime flag
func GetContractBin_v1() []byte {
	return hexutil.MustDecode("0x00")
}

// the SFC proxy contract address
var ContractAddress = common.HexToAddress("0xfa00face00fc0000000000000000000000000100")

// the SFC contract first implementation address
//var ContractAddress_v1 = common.HexToAddress("0xfa00beef0fc00000000000000000000000000101")

// AssembleStorage builds the genesis storage for the SFC contract
func AssembleStorage(validators pos.Validators, genesisTime inter.Timestamp, storage map[common.Hash]common.Hash) map[common.Hash]common.Hash {
	if storage == nil {
		storage = make(map[common.Hash]common.Hash)
	}

	// set validators
	for stakerIdx, validator := range validators.SortedAddresses() { // sort validators to get deterministic stakerIdxs
		position := sfcpos.VStake(validator)

		stakeAmount := common.BytesToHash(pos.StakeToBalance(validators.Get(validator)).Bytes())

		storage[position.StakeAmount()] = stakeAmount
		storage[position.CreatedEpoch()] = utils.U64to256(0)
		storage[position.CreatedTime()] = utils.U64to256(uint64(genesisTime.Unix()))
		storage[position.StakerIdx()] = utils.U64to256(uint64(stakerIdx + 1))
	}

	storage[sfcpos.VStakersNum()] = utils.U64to256(uint64(validators.Len()))
	storage[sfcpos.VStakersLastIdx()] = utils.U64to256(uint64(validators.Len()))
	storage[sfcpos.VStakeTotalAmount()] = common.BytesToHash((pos.StakeToBalance(validators.TotalStake())).Bytes())

	return storage
}
