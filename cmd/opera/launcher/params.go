package launcher

import (
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/ethereum/go-ethereum/params"

	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/opera/genesis"
	"github.com/Fantom-foundation/go-opera/opera/genesisstore"
)

var (
	Bootnodes = map[string][]string{
		"main": {
			"enode://03c70d4597d731ef182678b7664f2a4a3add07056f23d4e01aba86f066080d18fa13abbd2e13e9d4ea762a2715a983b5ac6151162d05ee0434f1847da1a626e9@34.242.220.16:5050",
			"enode://01c64d1a9dd8a65c56f2d4e373795eb6efd27b714b2b5999363a42a0edc39d7417a431416ceb5c67b1a170983af109e8a15d0c2d44a2ac41ecfb5c23c1a1a48a@3.35.200.210:5050",
			"enode://7044c88daa5df059e2f7a2667471a8149a5cf66e68643dcb86f399d48c4ff6475b73ee91486ea830d225f7f78a2fdf955208673da51c6852230c3a90a3701c06@3.1.103.70:5050",
			"enode://594d26c2338566daca9391d73f1b1821bb0b454e6f3d48715116bf42f320924d569534c143b640feec8a8eaa137a0b822426fb62b52a90162270ea5868bdc37c@18.138.254.181:5050",
			"enode://339e331912e5239a9e13eb82b47be58ea4d3946e91caa2992103a8d4f0226c1e86f9134822d5b238f25c9cbdd473f806caa8e4f8ef1748a6c66395f4bf0dd569@54.66.206.151:5050",
		},
		"test": {
			"enode://563b30428f48357f31c9d4906ca2f3d3815d663b151302c1ba9d58f3428265b554398c6fabf4b806a49525670cd9e031257c805375b9fdbcc015f60a7943e427@3.213.142.230:7946",
			"enode://8b53fe4410cde82d98d28697d56ccb793f9a67b1f8807c523eadafe96339d6e56bc82c0e702757ac5010972e966761b1abecb4935d9a86a9feed47e3e9ba27a6@3.227.34.226:7946",
			"enode://1703640d1239434dcaf010541cafeeb3c4c707be9098954c50aa705f6e97e2d0273671df13f6e447563e7d3a7c7ffc88de48318d8a3cc2cc59d196516054f17e@52.72.222.228:7946",
		},
	}

	mainnetHeader = genesis.Header{
		GenesisID:   hash.HexToHash("0x4a53c5445584b3bfc20dbfb2ec18ae20037c716f3ba2d9e1da768a9deca17cb4"),
		NetworkID:   opera.MainNetworkID,
		NetworkName: "main",
	}

	testnetHeader = genesis.Header{
		GenesisID:   hash.HexToHash("0xc4a5fc96e575a16a9a0c7349d44dc4d0f602a54e0a8543360c2fee4c3937b49e"),
		NetworkID:   opera.TestNetworkID,
		NetworkName: "test",
	}

	AllowedOperaGenesis = []GenesisTemplate{
		{
			Name:   "Mainnet-5577 with pruned MPT",
			Header: mainnetHeader,
			Hashes: genesis.Hashes{
				genesisstore.EpochsSection(0): hash.HexToHash("0x945d8084b4e6e1e78cfe9472fefca3f6ecc7041765dfed24f64e9946252f569a"),
				genesisstore.BlocksSection(0): hash.HexToHash("0xe3ec041f3cca79928aa4abef588b48e96ff3cfa3908b2268af3ac5496c722fec"),
				genesisstore.EvmSection(0):    hash.HexToHash("0x12dd52ac21fee5d76b47a64386e73187d5260e448e8044f38c6c73eaa627e4b5"),
			},
		},
		{
			Name:   "Mainnet-5577 with full MPT",
			Header: mainnetHeader,
			Hashes: genesis.Hashes{
				genesisstore.EpochsSection(0): hash.HexToHash("0x945d8084b4e6e1e78cfe9472fefca3f6ecc7041765dfed24f64e9946252f569a"),
				genesisstore.BlocksSection(0): hash.HexToHash("0xe3ec041f3cca79928aa4abef588b48e96ff3cfa3908b2268af3ac5496c722fec"),
				genesisstore.EvmSection(0):    hash.HexToHash("0x54614c9475963ed706f3e654bee0faf9ca21e29c588ad4070fd5b5897c8e0b5d"),
			},
		},
		{
			Name:   "Mainnet-109331 without history",
			Header: mainnetHeader,
			Hashes: genesis.Hashes{
				genesisstore.EpochsSection(0): hash.HexToHash("0xf0bef59de85dde7772bf43c62267eb312b0dd2c412ec5f04d96b6ea55178e901"),
				genesisstore.BlocksSection(0): hash.HexToHash("0x80fb348f77f65f0c357f69e29ca123a4c8f1ba60ff445510474be952a1e28d7a"),
			},
		},
		{
			Name:   "Mainnet-109331 without MPT",
			Header: mainnetHeader,
			Hashes: genesis.Hashes{
				genesisstore.EpochsSection(0): hash.HexToHash("0x71cdb819c2745a4853016bbb9690053b70fac679b168cd9a4999bf2a3dfb5578"),
				genesisstore.BlocksSection(0): hash.HexToHash("0xb7394f84b73528423a5b634bfb3cec8ab0a015b387bf6cbe70b378b08e9253bd"),
			},
		},
		{
			Name:   "Mainnet-109331 with pruned MPT",
			Header: mainnetHeader,
			Hashes: genesis.Hashes{
				genesisstore.EpochsSection(0): hash.HexToHash("0x71cdb819c2745a4853016bbb9690053b70fac679b168cd9a4999bf2a3dfb5578"),
				genesisstore.BlocksSection(0): hash.HexToHash("0xb7394f84b73528423a5b634bfb3cec8ab0a015b387bf6cbe70b378b08e9253bd"),
				genesisstore.EvmSection(0):    hash.HexToHash("0x617b8c4d74d1598f7d3914ba4c7cd46b7c98d5e044987c6c8d023cc59e849df7"),
			},
		},
		{
			Name:   "Mainnet-109331 with full MPT",
			Header: mainnetHeader,
			Hashes: genesis.Hashes{
				genesisstore.EpochsSection(0): hash.HexToHash("0x71cdb819c2745a4853016bbb9690053b70fac679b168cd9a4999bf2a3dfb5578"),
				genesisstore.BlocksSection(0): hash.HexToHash("0xb7394f84b73528423a5b634bfb3cec8ab0a015b387bf6cbe70b378b08e9253bd"),
				genesisstore.EvmSection(0):    hash.HexToHash("0xef0e1b833321a8de98aaaa1a3946378c78d66ab16b39eb0ad56636d5f7f9f2c5"),
			},
		},

		{
			Name:   "Mainnet-171200 without history",
			Header: mainnetHeader,
			Hashes: genesis.Hashes{
				genesisstore.EpochsSection(0): hash.HexToHash("0xb1954a0ac01a35a9ff6026d239e0be659fb4c37c356a94318bbe9201b9b2f3bf"),
				genesisstore.BlocksSection(0): hash.HexToHash("0x48ea3ccd2e2ff819386aa6d5ba86b12abb92f2f2e3b405e903f71e3f33f1d258"),
			},
		},
		{
			Name:   "Mainnet-171200 without MPT",
			Header: mainnetHeader,
			Hashes: genesis.Hashes{
				genesisstore.EpochsSection(0): hash.HexToHash("0x71cdb819c2745a4853016bbb9690053b70fac679b168cd9a4999bf2a3dfb5578"),
				genesisstore.BlocksSection(0): hash.HexToHash("0xb7394f84b73528423a5b634bfb3cec8ab0a015b387bf6cbe70b378b08e9253bd"),
				genesisstore.EpochsSection(1): hash.HexToHash("0xda430371772ee2fefd1caa342b6a5cb188041a01730f681099dd241bc57a3f77"),
				genesisstore.BlocksSection(1): hash.HexToHash("0x14b8b9c3b47cc174ae5c36599cebdef551ad35032ed29c087abb814ac5559619"),
			},
		},
		{
			Name:   "Mainnet-171200 with pruned MPT",
			Header: mainnetHeader,
			Hashes: genesis.Hashes{
				genesisstore.EpochsSection(0): hash.HexToHash("0x71cdb819c2745a4853016bbb9690053b70fac679b168cd9a4999bf2a3dfb5578"),
				genesisstore.BlocksSection(0): hash.HexToHash("0xb7394f84b73528423a5b634bfb3cec8ab0a015b387bf6cbe70b378b08e9253bd"),
				genesisstore.EpochsSection(1): hash.HexToHash("0xda430371772ee2fefd1caa342b6a5cb188041a01730f681099dd241bc57a3f77"),
				genesisstore.BlocksSection(1): hash.HexToHash("0x14b8b9c3b47cc174ae5c36599cebdef551ad35032ed29c087abb814ac5559619"),
				genesisstore.EvmSection(0):    hash.HexToHash("0x2a685df416eeca50f4b725117ae88deb35f05e3c51f34e9555ff6ffc62e75d14"),
			},
		},

		{
			Name:   "Testnet-2458 with pruned MPT",
			Header: testnetHeader,
			Hashes: genesis.Hashes{
				genesisstore.EpochsSection(0): hash.HexToHash("0x4a5caf86d7f5a31dad91f2cbd44db052c602f515e5319f828adb585a7a6723d6"),
				genesisstore.BlocksSection(0): hash.HexToHash("0x07eadb81c1e2a1b5c444c8c2430c6873380f447de64790b25abe9e7fa6874f65"),
				genesisstore.EvmSection(0):    hash.HexToHash("0xa96e006ae17d15e1244c3e7ff4d556e5a3849e70df7a81704787f3273f37c9b1"),
			},
		},
		{
			Name:   "Testnet-2458 with full MPT",
			Header: testnetHeader,
			Hashes: genesis.Hashes{
				genesisstore.EpochsSection(0): hash.HexToHash("0x4a5caf86d7f5a31dad91f2cbd44db052c602f515e5319f828adb585a7a6723d6"),
				genesisstore.BlocksSection(0): hash.HexToHash("0x07eadb81c1e2a1b5c444c8c2430c6873380f447de64790b25abe9e7fa6874f65"),
				genesisstore.EvmSection(0):    hash.HexToHash("0x3c635232f82cabdfc76405fd03c58134fb00ff9fc0ad080a8a8ae40a7a6fe604"),
			},
		},
		{
			Name:   "Testnet-6226 without history",
			Header: testnetHeader,
			Hashes: genesis.Hashes{
				genesisstore.EpochsSection(0): hash.HexToHash("0x5527a0c5e45b84c1350ccc77a9644eb797afd1e87887c35526d7622b12881b22"),
				genesisstore.BlocksSection(0): hash.HexToHash("0xf209c98aa5d3473dd71164165152e8802fb95b71d9dbfe394a0addcf81808d5c"),
			},
		},
		{
			Name:   "Testnet-6226 without MPT",
			Header: testnetHeader,
			Hashes: genesis.Hashes{
				genesisstore.EpochsSection(0): hash.HexToHash("0x61040a80f16755b86d67f13880f55c1238d307e2e1c6fc87951eb3bdee0bdff2"),
				genesisstore.BlocksSection(0): hash.HexToHash("0x12010621d8cf4dcd4ea357e98eac61edf9517a6df752cb2d929fca69e56bd8d1"),
			},
		},
		{
			Name:   "Testnet-6226 with pruned MPT",
			Header: testnetHeader,
			Hashes: genesis.Hashes{
				genesisstore.EpochsSection(0): hash.HexToHash("0x61040a80f16755b86d67f13880f55c1238d307e2e1c6fc87951eb3bdee0bdff2"),
				genesisstore.BlocksSection(0): hash.HexToHash("0x12010621d8cf4dcd4ea357e98eac61edf9517a6df752cb2d929fca69e56bd8d1"),
				genesisstore.EvmSection(0):    hash.HexToHash("0x86ec3c7938ab053fc84bbbc8f5259bc81885ec424df91272c553f371464840fc"),
			},
		},
		{
			Name:   "Testnet-6226 with full MPT",
			Header: testnetHeader,
			Hashes: genesis.Hashes{
				genesisstore.EpochsSection(0): hash.HexToHash("0x61040a80f16755b86d67f13880f55c1238d307e2e1c6fc87951eb3bdee0bdff2"),
				genesisstore.BlocksSection(0): hash.HexToHash("0x12010621d8cf4dcd4ea357e98eac61edf9517a6df752cb2d929fca69e56bd8d1"),
				genesisstore.EvmSection(0):    hash.HexToHash("0x9227c80bf56e4af08dc32cb6043cc43672f2be8177d550ab34a7a9f57f4f104b"),
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
