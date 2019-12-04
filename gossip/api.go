package gossip

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/params"
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

// Etherbase is the address that mining rewards will be send to.
func (api *PublicEthereumAPI) Etherbase() (common.Address, error) {
	return api.s.emitter.GetCoinbase(), nil
}

// Coinbase is the address that mining rewards will be send to (alias for Etherbase)
func (api *PublicEthereumAPI) Coinbase() (common.Address, error) {
	return api.s.emitter.GetCoinbase(), nil
}

// Hashrate returns the POW hashrate
func (api *PublicEthereumAPI) Hashrate() hexutil.Uint64 {
	return hexutil.Uint64(0)
}

// ChainID is the EIP-155 replay-protection chain id for the current ethereum chain config.
func (api *PublicEthereumAPI) ChainId() hexutil.Uint64 {
	return hexutil.Uint64(params.AllEthashProtocolChanges.ChainID.Uint64())
}
