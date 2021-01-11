package driverpos

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Events
var (
	// Topics of Driver contract logs
	Topics = struct {
		IncBalance            common.Hash
		SetBalance            common.Hash
		SubBalance            common.Hash
		SetCode               common.Hash
		SwapCode              common.Hash
		SetStorage            common.Hash
		UpdateValidatorWeight common.Hash
		UpdateValidatorPubkey common.Hash
		UpdateNetworkRules    common.Hash
	}{
		IncBalance:            crypto.Keccak256Hash([]byte("IncBalance(address,uint256)")),
		SetBalance:            crypto.Keccak256Hash([]byte("SetBalance(address,uint256)")),
		SubBalance:            crypto.Keccak256Hash([]byte("SubBalance(address,uint256)")),
		SetCode:               crypto.Keccak256Hash([]byte("SetCode(address,address)")),
		SwapCode:              crypto.Keccak256Hash([]byte("SwapCode(address,address)")),
		SetStorage:            crypto.Keccak256Hash([]byte("SetStorage(address,uint256,uint256)")),
		UpdateValidatorWeight: crypto.Keccak256Hash([]byte("UpdateValidatorWeight(uint256,uint256)")),
		UpdateValidatorPubkey: crypto.Keccak256Hash([]byte("UpdateValidatorPubkey(uint256,bytes)")),
		UpdateNetworkRules:    crypto.Keccak256Hash([]byte("UpdateNetworkRules(bytes)")),
	}
)
