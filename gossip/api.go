package gossip

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// PublicEthereumAPI provides an API to access Ethereum-like information.
// It is a github.com/ethereum/go-ethereum/eth simulation for console.
type PublicEthereumAPI struct {
	s *Service
}

// NewPublicEthereumAPI creates a new Ethereum protocol API for gossip.
func NewPublicEthereumAPI(s *Service) *PublicEthereumAPI {
	return &PublicEthereumAPI{s}
}

// Etherbase is the validator address
func (api *PublicEthereumAPI) Etherbase() (common.Address, error) {
	return api.Validator()
}

// Coinbase is the validator address
func (api *PublicEthereumAPI) Coinbase() (common.Address, error) {
	return api.Validator()
}

// Validator is the validator address
func (api *PublicEthereumAPI) Validator() (common.Address, error) {
	_, addr := api.s.emitter.GetValidator()
	return addr, nil
}

// Hashrate returns the POW hashrate
func (api *PublicEthereumAPI) Hashrate() hexutil.Uint64 {
	return hexutil.Uint64(0)
}

// ChainId is the EIP-155 replay-protection chain id for the current ethereum chain config.
func (api *PublicEthereumAPI) ChainId() hexutil.Uint64 {
	return hexutil.Uint64(api.s.config.Net.EvmChainConfig().ChainID.Uint64())
}
