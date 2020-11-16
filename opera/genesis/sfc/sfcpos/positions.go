package sfcpos

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Events

var (
	// Topics of SFC contract logs
	Topics = struct {
		IncBalance             common.Hash
		UpdatedValidatorWeight common.Hash
		UpdatedValidatorPubkey common.Hash
	}{
		IncBalance:             crypto.Keccak256Hash([]byte("IncBalance(address,uint256)")),
		UpdatedValidatorWeight: crypto.Keccak256Hash([]byte("UpdatedValidatorWeight(uint256,uint256)")),
		UpdatedValidatorPubkey: crypto.Keccak256Hash([]byte("UpdatedValidatorPubkey(uint256,bytes)")),
	}
)
