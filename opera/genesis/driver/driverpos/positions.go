package driverpos

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Events
var (
	// Topics of Driver contract logs
	Topics = struct {
		UpdateValidatorWeight common.Hash
		UpdateValidatorPubkey common.Hash
		UpdateNetworkRules    common.Hash
		UpdateNetworkVersion  common.Hash
		AdvanceEpochs         common.Hash
	}{
		UpdateValidatorWeight: crypto.Keccak256Hash([]byte("UpdateValidatorWeight(uint256,uint256)")),
		UpdateValidatorPubkey: crypto.Keccak256Hash([]byte("UpdateValidatorPubkey(uint256,bytes)")),
		UpdateNetworkRules:    crypto.Keccak256Hash([]byte("UpdateNetworkRules(bytes)")),
		UpdateNetworkVersion:  crypto.Keccak256Hash([]byte("UpdateNetworkVersion(uint256)")),
		AdvanceEpochs:         crypto.Keccak256Hash([]byte("AdvanceEpochs(uint256)")),
	}
)
