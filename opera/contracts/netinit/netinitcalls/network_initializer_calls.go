package netinitcall

import (
	"math/big"
	"strings"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-opera/utils"
)

const ContractABI = "[{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"sealedEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalSupply\",\"type\":\"uint256\"},{\"internalType\":\"addresspayable\",\"name\":\"_sfc\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_lib\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_auth\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_driver\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_evmWriter\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"initializeAll\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

var (
	sAbi, _ = abi.JSON(strings.NewReader(ContractABI))
)

// Methods

func InitializeAll(sealedEpoch idx.Epoch, totalSupply *big.Int, sfcAddr common.Address, libAddr common.Address, driverAuthAddr common.Address, driverAddr common.Address, evmWriterAddr common.Address, owner common.Address) []byte {
	data, _ := sAbi.Pack("initializeAll", utils.U64toBig(uint64(sealedEpoch)), totalSupply, sfcAddr, libAddr, driverAuthAddr, driverAddr, evmWriterAddr, owner)
	return data
}
