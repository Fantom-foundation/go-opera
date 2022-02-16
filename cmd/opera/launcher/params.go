package launcher

import (
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/ethereum/go-ethereum/params"

	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/opera/genesis"
)

var (
	Bootnodes = []string{
		"enode://03c70d4597d731ef182678b7664f2a4a3add07056f23d4e01aba86f066080d18fa13abbd2e13e9d4ea762a2715a983b5ac6151162d05ee0434f1847da1a626e9@34.242.220.16:5050",
		"enode://01c64d1a9dd8a65c56f2d4e373795eb6efd27b714b2b5999363a42a0edc39d7417a431416ceb5c67b1a170983af109e8a15d0c2d44a2ac41ecfb5c23c1a1a48a@3.35.200.210:5050",
		"enode://7044c88daa5df059e2f7a2667471a8149a5cf66e68643dcb86f399d48c4ff6475b73ee91486ea830d225f7f78a2fdf955208673da51c6852230c3a90a3701c06@3.1.103.70:5050",
		"enode://594d26c2338566daca9391d73f1b1821bb0b454e6f3d48715116bf42f320924d569534c143b640feec8a8eaa137a0b822426fb62b52a90162270ea5868bdc37c@18.138.254.181:5050",
		"enode://339e331912e5239a9e13eb82b47be58ea4d3946e91caa2992103a8d4f0226c1e86f9134822d5b238f25c9cbdd473f806caa8e4f8ef1748a6c66395f4bf0dd569@54.66.206.151:5050",
		"enode://563b30428f48357f31c9d4906ca2f3d3815d663b151302c1ba9d58f3428265b554398c6fabf4b806a49525670cd9e031257c805375b9fdbcc015f60a7943e427@3.213.142.230:7946",
		"enode://8b53fe4410cde82d98d28697d56ccb793f9a67b1f8807c523eadafe96339d6e56bc82c0e702757ac5010972e966761b1abecb4935d9a86a9feed47e3e9ba27a6@3.227.34.226:7946",
		"enode://1703640d1239434dcaf010541cafeeb3c4c707be9098954c50aa705f6e97e2d0273671df13f6e447563e7d3a7c7ffc88de48318d8a3cc2cc59d196516054f17e@52.72.222.228:7946",
	}

	//mainnetHeader = genesis.Header{
	//	GenesisID:   hash.HexToHash("0x4a53c5445584b3bfc20dbfb2ec18ae20037c716f3ba2d9e1da768a9deca17cb4"),
	//	NetworkID:   opera.MainNetworkID,
	//	NetworkName: "main",
	//}

	testnetHeader = genesis.Header{
		GenesisID:   hash.HexToHash("0xc4a5fc96e575a16a9a0c7349d44dc4d0f602a54e0a8543360c2fee4c3937b49e"),
		NetworkID:   opera.TestNetworkID,
		NetworkName: "test",
	}

	AllowedOperaGenesis = []GenesisTemplate{
		{
			Name:   "Testnet-2457 with pruned MPT",
			Header: testnetHeader,
			Hashes: genesis.Hashes{
				Epochs:      hash.Hashes{hash.HexToHash("0x19833ee6780afe4c946f9b3645cae8776be5e040966c431b947b31b9ac495000")},
				Blocks:      hash.Hashes{hash.HexToHash("0x764a3ae2984127cb1076be6327dd2ae56fa4583deae189b481396d914efdc077")},
				RawEvmItems: hash.Hashes{hash.HexToHash("0x2756b0c5f69876a2b315ef7940ac6f36463d2bf71d59589534de32672dbb0602")},
			},
		},
		{
			Name:   "Testnet-2457 with full MPT",
			Header: testnetHeader,
			Hashes: genesis.Hashes{
				Epochs:      hash.Hashes{hash.HexToHash("0x19833ee6780afe4c946f9b3645cae8776be5e040966c431b947b31b9ac495000")},
				Blocks:      hash.Hashes{hash.HexToHash("0x764a3ae2984127cb1076be6327dd2ae56fa4583deae189b481396d914efdc077")},
				RawEvmItems: hash.Hashes{hash.HexToHash("0x7b263279ff8b76130fdc659578a10995a230d80f180722a9020c9cec9fb9d493")},
			},
		},
		{
			Name:   "Testnet-6226 without history",
			Header: testnetHeader,
			Hashes: genesis.Hashes{
				Epochs:      hash.Hashes{hash.HexToHash("0xd277c6912f948cfdf29351b9c58b5ee2c5ffe295df217d8cc22658bac7ea6f64")},
				Blocks:      hash.Hashes{hash.HexToHash("0x99680e633477b012ae01d0eb67eac2edeb2b8146539d079bfdc6de4346e51556")},
				RawEvmItems: hash.Hashes{},
			},
		},
		{
			Name:   "Testnet-6226 without MPT",
			Header: testnetHeader,
			Hashes: genesis.Hashes{
				Epochs:      hash.Hashes{hash.HexToHash("0x8f97e4d9906fccd253a4c82e24fa31fcb02bb099946dcebb520972d4d5470efe")},
				Blocks:      hash.Hashes{hash.HexToHash("0x1ebf436f5a9af11d60af99e1f629422bc72f81e57ee2cdc8b1f753d3dae39eb1")},
				RawEvmItems: hash.Hashes{},
			},
		},
		{
			Name:   "Testnet-6226 with pruned MPT",
			Header: testnetHeader,
			Hashes: genesis.Hashes{
				Epochs:      hash.Hashes{hash.HexToHash("0x8f97e4d9906fccd253a4c82e24fa31fcb02bb099946dcebb520972d4d5470efe")},
				Blocks:      hash.Hashes{hash.HexToHash("0x1ebf436f5a9af11d60af99e1f629422bc72f81e57ee2cdc8b1f753d3dae39eb1")},
				RawEvmItems: hash.Hashes{hash.HexToHash("0x6e620d8a6655a4dfb88e964a0b1b4fc86452ca2d5bcb8c57cc81ed9ce08e6ddb")},
			},
		},
		{
			Name:   "Testnet-6226 with full MPT",
			Header: testnetHeader,
			Hashes: genesis.Hashes{
				Epochs:      hash.Hashes{hash.HexToHash("0x8f97e4d9906fccd253a4c82e24fa31fcb02bb099946dcebb520972d4d5470efe")},
				Blocks:      hash.Hashes{hash.HexToHash("0x1ebf436f5a9af11d60af99e1f629422bc72f81e57ee2cdc8b1f753d3dae39eb1")},
				RawEvmItems: hash.Hashes{hash.HexToHash("0xe8ddfc1a271d9700f48e788d8fa928e29170805a53289d74fedef2a7bc53fba0")},
			},
		},
	}
)

func overrideParams() {
	params.MainnetBootnodes = []string{}
	params.RopstenBootnodes = []string{}
	params.RinkebyBootnodes = []string{}
	params.GoerliBootnodes = []string{}
}
