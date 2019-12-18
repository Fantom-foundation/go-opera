package sfc

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/lachesis/genesis/sfc/sfcpos"
	"github.com/Fantom-foundation/go-lachesis/utils"
)

// GetMainContractBinV1 is SFC contract first implementation bin code for mainnet
// Must be compiled with bin-runtime flag
func GetMainContractBinV1() []byte {
	return hexutil.MustDecode("0x00")
}

// GetTestContractBinV1 is SFC contract first implementation bin code for testnet
// Must be compiled with bin-runtime flag
func GetTestContractBinV1() []byte {
	return hexutil.MustDecode("0x00")
}

// ContractAddress is the SFC proxy contract address
var ContractAddress = common.HexToAddress("0xfa00face00fc0000000000000000000000000100")

// the SFC contract first implementation address
//var ContractAddress_v1 = common.HexToAddress("0xfa00beef0fc00000000000000000000000000101")

// AssembleStorage builds the genesis storage for the SFC contract
func AssembleStorage(validators pos.GValidators, genesisTime inter.Timestamp, storage map[common.Hash]common.Hash) map[common.Hash]common.Hash {
	if storage == nil {
		storage = make(map[common.Hash]common.Hash)
	}

	// set validators
	maxStakerID := idx.StakerID(0)
	for _, validator := range validators {
		stakePos := sfcpos.Staker(validator.ID)

		stakeAmount := utils.BigTo256(pos.StakeToBalance(validator.Stake))

		storage[stakePos.StakeAmount()] = stakeAmount
		storage[stakePos.CreatedEpoch()] = utils.U64to256(0)
		storage[stakePos.CreatedTime()] = utils.U64to256(uint64(genesisTime.Unix()))
		storage[stakePos.Address()] = validator.Address.Hash()

		stakerIDPos := sfcpos.StakerID(validator.Address)
		storage[stakerIDPos] = utils.U64to256(uint64(validator.ID))

		if maxStakerID < validator.ID {
			maxStakerID = validator.ID
		}
	}

	storage[sfcpos.StakersNum()] = utils.U64to256(uint64(len(validators)))
	storage[sfcpos.StakersLastIdx()] = utils.U64to256(uint64(maxStakerID))
	storage[sfcpos.StakeTotalAmount()] = utils.BigTo256((pos.StakeToBalance(validators.Validators().TotalStake())))

	return storage
}
