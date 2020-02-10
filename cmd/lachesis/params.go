package main

import (
	"github.com/ethereum/go-ethereum/params"
)

func overrideParams() {
	// DiscoveryV5Bootnodes are the enode URLs of the P2P bootstrap nodes for the
	// experimental RLPx v5 topic-discovery network.
	params.DiscoveryV5Bootnodes = []string{
		// mainnet
		"enode://a2941866e485442aa6b17d67d77f8a6c4580bb556894cc1618473eff1e18203d8cce50b563cf4c75e408886079b8f067069442ed52e2ac9e556baa3f8fcc525f@3.24.15.66:5050",
		"enode://9ea4e8ad12e1fcfc846d00a626fbf3ca1ee6a5143148b8472dd62a64b876d01031c064a717aaa606de34bcbbb691b50cbdf816acd9e9ee7ce334082b739829b9@52.45.126.213:5050",
		"enode://cba8bd4179052908d069c6bacccf999dd576fdfe6fd29db19b21ae262d96a9ac00a3f5904b772e44b172575b8afbf91c0d1303f1b7b03dfae14e50e234004d36@15.164.136.219:5050",
		"enode://fb904114975d7b238c2d5e46824ac00fdd1133f836e6e5332765e089e70c6dbd4d47c513fcb2ee72774484216634cc9785a54816bfde46d721871d340070b64b@34.245.17.87:5050",
		// testnet
		"enode://563b30428f48357f31c9d4906ca2f3d3815d663b151302c1ba9d58f3428265b554398c6fabf4b806a49525670cd9e031257c805375b9fdbcc015f60a7943e427@3.213.142.230:7946",
		"enode://8b53fe4410cde82d98d28697d56ccb793f9a67b1f8807c523eadafe96339d6e56bc82c0e702757ac5010972e966761b1abecb4935d9a86a9feed47e3e9ba27a6@3.227.34.226:7946",
		"enode://1703640d1239434dcaf010541cafeeb3c4c707be9098954c50aa705f6e97e2d0273671df13f6e447563e7d3a7c7ffc88de48318d8a3cc2cc59d196516054f17e@52.72.222.228:7946",
	}

	// MainnetBootnodes are the enode URLs of the discovery V4 P2P bootstrap nodes running on
	// the main Opera network.
	params.MainnetBootnodes = []string{}

	// TestnetBootnodes are the enode URLs of the discovery V4 P2P bootstrap nodes running on
	//  the test Opera network.
	params.TestnetBootnodes = []string{}

	params.RinkebyBootnodes = []string{}
	params.GoerliBootnodes = []string{}

}
