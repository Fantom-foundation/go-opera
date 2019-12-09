package sfc

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/lachesis/genesis/sfc/sfcpos"
	"github.com/Fantom-foundation/go-lachesis/utils"
)

// GetContractBinV1 is SFC contract first implementation bin code
// Must be compiled with bin-runtime flag
func GetContractBinV1() []byte {
	return hexutil.MustDecode("0x00")
}

// ContractAddress is the SFC proxy contract address
var ContractAddress = common.HexToAddress("0xfa00face00fc0000000000000000000000000100")

// the SFC contract first implementation address
//var ContractAddress_v1 = common.HexToAddress("0xfa00beef0fc00000000000000000000000000101")

// AssembleStorage builds the genesis storage for the SFC contract
func AssembleStorage(validators pos.Validators, genesisTime inter.Timestamp, storage map[common.Hash]common.Hash) map[common.Hash]common.Hash {
	if storage == nil {
		storage = make(map[common.Hash]common.Hash)
	}

	// set validators
	for i, validator := range validators.SortedAddresses() { // sort validators to get deterministic stakerIDs
		stakerID := uint64(i + 1)
		stakePos := sfcpos.VStake(stakerID)

		stakeAmount := utils.BigTo256(pos.StakeToBalance(validators.Get(validator)))

		storage[stakePos.StakeAmount()] = stakeAmount
		storage[stakePos.CreatedEpoch()] = utils.U64to256(0)
		storage[stakePos.CreatedTime()] = utils.U64to256(uint64(genesisTime.Unix()))
		storage[stakePos.Address()] = validator.Hash()

		stakerIDPos := sfcpos.VStakerID(validator)
		storage[stakerIDPos] = utils.U64to256(stakerID)
	}

	storage[sfcpos.VStakersNum()] = utils.U64to256(uint64(validators.Len()))
	storage[sfcpos.VStakersLastIdx()] = utils.U64to256(uint64(validators.Len()))
	storage[sfcpos.VStakeTotalAmount()] = utils.BigTo256((pos.StakeToBalance(validators.TotalStake())))

	return storage
}
