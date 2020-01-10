package main

import (
	"github.com/ethereum/go-ethereum/params"
)

func overrideParams() {
	// DiscoveryV5Bootnodes are the enode URLs of the P2P bootstrap nodes for the
	// experimental RLPx v5 topic-discovery network.
	params.DiscoveryV5Bootnodes = []string{
		"enode://a2941866e485442aa6b17d67d77f8a6c4580bb556894cc1618473eff1e18203d8cce50b563cf4c75e408886079b8f067069442ed52e2ac9e556baa3f8fcc525f@3.24.15.66:5050",
		"enode://9ea4e8ad12e1fcfc846d00a626fbf3ca1ee6a5143148b8472dd62a64b876d01031c064a717aaa606de34bcbbb691b50cbdf816acd9e9ee7ce334082b739829b9@52.45.126.213:5050",
		"enode://cba8bd4179052908d069c6bacccf999dd576fdfe6fd29db19b21ae262d96a9ac00a3f5904b772e44b172575b8afbf91c0d1303f1b7b03dfae14e50e234004d36@15.164.136.219:5050",
		"enode://fb904114975d7b238c2d5e46824ac00fdd1133f836e6e5332765e089e70c6dbd4d47c513fcb2ee72774484216634cc9785a54816bfde46d721871d340070b64b@34.245.17.87:5050",
	}

	// MainnetBootnodes are the enode URLs of the discovery V4 P2P bootstrap nodes running on
	// the main Opera network.
	params.MainnetBootnodes = []string{}

	// TestnetBootnodes are the enode URLs of the discovery V4 P2P bootstrap nodes running on the
	// Testnet test network.
	params.TestnetBootnodes = []string{}

	params.RinkebyBootnodes = []string{}
	params.GoerliBootnodes = []string{}
}
