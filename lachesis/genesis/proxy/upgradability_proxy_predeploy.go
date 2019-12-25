package proxy

import (
	"github.com/Fantom-foundation/go-lachesis/lachesis/genesis/proxy/proxypos"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// GetContractBin is SFC contract first implementation bin code for mainnet
// Must be compiled with bin-runtime flag
func GetContractBin() []byte {
	return hexutil.MustDecode("0x00")
}

// AssembleStorage builds genesis storage for the Upgradability contract
func AssembleStorage(admin common.Address, implementation common.Address, storage map[common.Hash]common.Hash) map[common.Hash]common.Hash {
	if storage == nil {
		storage = make(map[common.Hash]common.Hash)
	}
	storage[proxypos.Admin()] = admin.Hash()
	storage[proxypos.Implementation()] = implementation.Hash()
	return storage
}
