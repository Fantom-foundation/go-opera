package main

import (
	"github.com/ethereum/go-ethereum/params"
)

func overrideParams() {
	// MainnetBootnodes are the enode URLs of the P2P bootstrap nodes running on
	// the main Ethereum network.
	params.MainnetBootnodes = []string{}

	// TestnetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
	// Ropsten test network.
	params.TestnetBootnodes = []string{}

	// RinkebyBootnodes are the enode URLs of the P2P bootstrap nodes running on the
	// Rinkeby test network.
	params.RinkebyBootnodes = []string{}

	// GoerliBootnodes are the enode URLs of the P2P bootstrap nodes running on the
	// Görli test network.
	params.GoerliBootnodes = []string{}

	// DiscoveryV5Bootnodes are the enode URLs of the P2P bootstrap nodes for the
	// experimental RLPx v5 topic-discovery network.
	params.DiscoveryV5Bootnodes = []string{
		/*
			"enode://06051a5573c81934c9554ef2898eb13b33a34b94cf36b202b69fde139ca17a85051979867720d4bdae4323d4943ddf9aeeb6643633aa656e0be843659795007a@35.177.226.168:30303",
			"enode://0cc5f5ffb5d9098c8b8c62325f3797f56509bff942704687b6530992ac706e2cb946b90a34f1f19548cd3c7baccbcaea354531e5983c7d1bc0dee16ce4b6440b@40.118.3.223:30304",
			"enode://1c7a64d76c0334b0418c004af2f67c50e36a3be60b5e4790bdac0439d21603469a85fad36f2473c9a80eb043ae60936df905fa28f1ff614c3e5dc34f15dcd2dc@40.118.3.223:30306",
			"enode://85c85d7143ae8bb96924f2b54f1b3e70d8c4d367af305325d30a61385a432f247d2c75c45c6b4a60335060d072d7f5b35dd1d4c45f76941f62a4f83b6e75daaf@40.118.3.223:30307",
		*/
	}
}
