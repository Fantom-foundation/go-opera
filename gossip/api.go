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

// Etherbase returns the zero address for web3 compatibility
func (api *PublicEthereumAPI) Etherbase() (common.Address, error) {
	return common.Address{}, nil
}

// Coinbase returns the zero address for web3 compatibility
func (api *PublicEthereumAPI) Coinbase() (common.Address, error) {
	return common.Address{}, nil
}

// Hashrate returns the zero POW hashrate for web3 compatibility
func (api *PublicEthereumAPI) Hashrate() hexutil.Uint64 {
	return hexutil.Uint64(0)
}

// ChainId is the EIP-155 replay-protection chain id for the current ethereum chain config.
func (api *PublicEthereumAPI) ChainId() hexutil.Uint64 {
	return hexutil.Uint64(api.s.store.GetRules().EvmChainConfig().ChainID.Uint64())
}
