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
			"enode://969654fc30f10e45903ee2d107cfc165feaff6f9acbd2a9472120a3581f7c717ded76113aa116e35b8389dfcc9129ecfc85a409f6c46fc864ce077cb85316ac8@3.252.226.57:5050",
			"enode://15dde1f31ff775da40d06a5529c15cce5bcdfec6055f8e8782b0df3f2ad3eab30b31fbb91037c6b3dacb8e612e177bdfaf4d82e9080feda31e54412efb56a34a@34.251.109.79:5050",
			"enode://41943a0269fc573355ff6a5e5980195cea68eb7bc703cc9937a3cd0debc9a795396aa9560bce65c8f329a9343f88ee4084daabcca76c5a6a8054aa163973a735@3.10.205.205:5050",
			"enode://b781fa90f63ef497c02ef6da8f32c2e14cc47921403d21083d5ebc69b5a85f77d6f124f084e3abbf3984e07aa5810b250658c04d6d3f3bc76e3735b2512847db@3.73.43.156:5050",
			"enode://fe97ec5b6c2073e4074067f3e513e66cf0a157c449bbb179dc8876803dc9cd3a18b03e735b87393c843cdc3be9c12989c2dd8ddaf88eda2f73a7183f2d65e95e@54.242.2.27:5050",
			"enode://2591357319aaa566897854f225d3a60194a8d193410de28e9ea74df2bb2615af63766351d8950c0582c970e3688ad85a8e454c5a601b9b091f094869fe603350@18.234.231.118:5050",
			"enode://7c6dbe419f79852536f3fc850912eaf657f50f378f6f6bc238bfe16bcf830c06ef3c0358774313eb94d959475a8bfa221bb53a4b0dcf3025ac071396199a3462@13.212.202.99:5050",
			"enode://e6c658978b95bcb29354f5a4a808b3d17d62201bcbd6ac36fa937e134c18b89ca23ee6f9f3893beabf07ec2f3e07b19d02d705ebc67ba671bff9830fa002c635@3.26.147.192:5050",
			"enode://9554ca82fb2cc3a2e3f30b847304e26e0ec6e73993c19610443a9fcd7c2a38796089a0ebe24c2c0bde6fb4155433c04104e049e9c6bd73a758c01e8cb93a7d9d@13.125.201.204:5050",
			"enode://b3699ce7ee4a8c6efb07707256a2a08872576f438664c6437cb263989116747d8d7332fb9cfc35c8f3ba04a024ecc6893cda1782db6fed221d9de2ec11746cfd@18.170.86.32:5050",
		},
		"test": {
			"enode://52c84c99a4cca9524dc626261e932aa8b1c88a103523132a1d6c30fd7d1f3eab0cb105403971baa7255f9eb5eadde9761db23bb277c4c7d6c3eefaf133dcb35f@170.64.156.90:7946",
			"enode://fffb2ee36f5571d3fa640afa38fcf3cd9389a51668f318c42b839cb31432155b34a12e94923a2230d4ff2d786b68ea30a1b9990ad9cefb663ae776279dbb63e5@157.245.202.7:7946",
			"enode://fecedd0b4dd953cb64547e00bf2c8d749911b8ba4870cce20dcbff171897ff2f89561428018892c8b63340a29711a3b271f26f2176fe97261d21dc103010c4af@165.232.96.92:7946",
			"enode://46a8746ac22f8b2b60aaed4820eae914193af873271897476f0b420d50951c479a8fb42591d42d17ec3233f52617646df45e9b6632fe7e418a390d99cc5d53ee@167.71.206.77:7946",
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

		{
			Name:   "Testnet-16200 without history",
			Header: mainnetHeader,
			Hashes: genesis.Hashes{
				genesisstore.EpochsSection(0): hash.HexToHash("0xd72e9bf39c645df8d978955fab8997a7e9cd7cb5812c007e2bb4b51a8c570a90"),
				genesisstore.BlocksSection(0): hash.HexToHash("0x7d651ed0e0f3e92ffd89cb52112598db54afd8bf3050bc083ff0bfe1b98948fd"),
			},
		},
		{
			Name:   "Testnet-16200 without MPT",
			Header: mainnetHeader,
			Hashes: genesis.Hashes{
				genesisstore.EpochsSection(0): hash.HexToHash("0x61040a80f16755b86d67f13880f55c1238d307e2e1c6fc87951eb3bdee0bdff2"),
				genesisstore.BlocksSection(0): hash.HexToHash("0x12010621d8cf4dcd4ea357e98eac61edf9517a6df752cb2d929fca69e56bd8d1"),
				genesisstore.EpochsSection(1): hash.HexToHash("0xd72e9bf39c645df8d978955fab8997a7e9cd7cb5812c007e2bb4b51a8c570a90"),
				genesisstore.BlocksSection(1): hash.HexToHash("0x7d651ed0e0f3e92ffd89cb52112598db54afd8bf3050bc083ff0bfe1b98948fd"),
			},
		},
		{
			Name:   "Testnet-16200 with pruned MPT",
			Header: mainnetHeader,
			Hashes: genesis.Hashes{
				genesisstore.EpochsSection(0): hash.HexToHash("0x61040a80f16755b86d67f13880f55c1238d307e2e1c6fc87951eb3bdee0bdff2"),
				genesisstore.BlocksSection(0): hash.HexToHash("0x12010621d8cf4dcd4ea357e98eac61edf9517a6df752cb2d929fca69e56bd8d1"),
				genesisstore.EpochsSection(1): hash.HexToHash("0xd72e9bf39c645df8d978955fab8997a7e9cd7cb5812c007e2bb4b51a8c570a90"),
				genesisstore.BlocksSection(1): hash.HexToHash("0x7d651ed0e0f3e92ffd89cb52112598db54afd8bf3050bc083ff0bfe1b98948fd"),
				genesisstore.EvmSection(0):    hash.HexToHash("0xbd66dcbbe77881d5aae5091ee9c455d213cebef2cc53c0d4bb356840c7020f7b"),
			},
		},
		{
			Name:   "Testnet-16200 with full MPT",
			Header: testnetHeader,
			Hashes: genesis.Hashes{
				genesisstore.EpochsSection(0): hash.HexToHash("0x61040a80f16755b86d67f13880f55c1238d307e2e1c6fc87951eb3bdee0bdff2"),
				genesisstore.BlocksSection(0): hash.HexToHash("0x12010621d8cf4dcd4ea357e98eac61edf9517a6df752cb2d929fca69e56bd8d1"),
				genesisstore.EpochsSection(1): hash.HexToHash("0xd72e9bf39c645df8d978955fab8997a7e9cd7cb5812c007e2bb4b51a8c570a90"),
				genesisstore.BlocksSection(1): hash.HexToHash("0x7d651ed0e0f3e92ffd89cb52112598db54afd8bf3050bc083ff0bfe1b98948fd"),
				genesisstore.EvmSection(0):    hash.HexToHash("0x9227c80bf56e4af08dc32cb6043cc43672f2be8177d550ab34a7a9f57f4f104b"),
				genesisstore.EvmSection(1):    hash.HexToHash("0x2376016f7ba13123244c6b56088a76e2e8bd5d5795fa92bad65f61488d12c236"),
			},
		},
	}
)

func overrideParams() {
	params.MainnetBootnodes = []string{}
	params.RinkebyBootnodes = []string{}
	params.GoerliBootnodes = []string{}
}
