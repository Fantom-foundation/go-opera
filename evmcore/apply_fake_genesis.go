// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package evmcore

import (
	"crypto/ecdsa"
	"errors"
	"math"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-opera/inter"
)

var FakeGenesisTime = inter.Timestamp(1608600000 * time.Second)

// ApplyFakeGenesis writes or updates the genesis block in db.
func ApplyFakeGenesis(statedb *state.StateDB, time inter.Timestamp, balances map[common.Address]*big.Int) (*EvmBlock, error) {
	for acc, balance := range balances {
		statedb.SetBalance(acc, balance)
	}

	// initial block
	root, err := flush(statedb, true)
	if err != nil {
		return nil, err
	}
	block := genesisBlock(time, root)

	return block, nil
}

func flush(statedb *state.StateDB, clean bool) (root common.Hash, err error) {
	root, err = statedb.Commit(clean)
	if err != nil {
		return
	}
	err = statedb.Database().TrieDB().Commit(root, false, nil)
	if err != nil {
		return
	}

	if !clean {
		err = statedb.Database().TrieDB().Cap(0)
	}

	return
}

// genesisBlock makes genesis block with pretty hash.
func genesisBlock(time inter.Timestamp, root common.Hash) *EvmBlock {
	block := &EvmBlock{
		EvmHeader: EvmHeader{
			Number:   big.NewInt(0),
			Time:     time,
			GasLimit: math.MaxUint64,
			Root:     root,
			TxHash:   types.EmptyRootHash,
		},
	}

	return block
}

// MustApplyFakeGenesis writes the genesis block and state to db, panicking on error.
func MustApplyFakeGenesis(statedb *state.StateDB, time inter.Timestamp, balances map[common.Address]*big.Int) *EvmBlock {
	block, err := ApplyFakeGenesis(statedb, time, balances)
	if err != nil {
		log.Crit("ApplyFakeGenesis", "err", err)
	}
	return block
}

// FakeKey gets n-th fake private key.
func FakeKey(n int) *ecdsa.PrivateKey {
	var key00, _ = crypto.ToECDSA(hexutil.MustDecode("0x163f5f0f9a621d72fedd85ffca3d08d131ab4e812181e0d30ffd1c885d20aac7"))
	var key01, _ = crypto.ToECDSA(hexutil.MustDecode("0x3144c0aa4ced56dc15c79b045bc5559a5ac9363d98db6df321fe3847a103740f"))
	var key02, _ = crypto.ToECDSA(hexutil.MustDecode("0x04a531f967898df5dbe223b67989b248e23c1c356a3f6717775cccb7fe53482c"))
	var key03, _ = crypto.ToECDSA(hexutil.MustDecode("0x00ca81d4fe11c23fae8b5e5b06f9fe952c99ca46abaec8bda70a678cd0314dde"))
	var key04, _ = crypto.ToECDSA(hexutil.MustDecode("0x532d9b2ce282fad94efefcf076fdfbe5befe558b145f4cc97f953bcabf087aeb"))
	var key05, _ = crypto.ToECDSA(hexutil.MustDecode("0x6e50dbd3e81b22424cb230133b87bc9ef0f17c584a2a5dc4b212d2b83b5ee084"))
	var key06, _ = crypto.ToECDSA(hexutil.MustDecode("0x2215aaee06a2d64ca32b201e1fb9d1e3c7a25d45a6d8b0de6300ba3a20e42ef5"))
	var key07, _ = crypto.ToECDSA(hexutil.MustDecode("0x1cd6fdfc633c0fa73bd306c46eecd23096365b44ab75f0e6fa04dc2adbea9583"))
	var key08, _ = crypto.ToECDSA(hexutil.MustDecode("0x2fc91d5829f44650c32ba92c8b29d511511446b91badf03b1fd0f808b91a4b5b"))
	var key09, _ = crypto.ToECDSA(hexutil.MustDecode("0x6aeeb7f09e757baa9d3935a042c3d0d46a2eda19e9b676283dce4eaf32e29dc9"))
	var key10, _ = crypto.ToECDSA(hexutil.MustDecode("0x7d51a817ee07c3f28581c47a5072142193337fdca4d7911e58c5af2d03895d1a"))
	var key11, _ = crypto.ToECDSA(hexutil.MustDecode("0x59963733b8a6fb1c6eeb1ce51c7e6046e652a9bcacd4cbaa3f6f26dafe7f79f7"))
	var key12, _ = crypto.ToECDSA(hexutil.MustDecode("0x4cf757812428b0764a871e94b02ba026a5d3738e69f7d1d4f9f93b43ed00e820"))
	var key13, _ = crypto.ToECDSA(hexutil.MustDecode("0xa80a59dc6a9be8003a696ed08a4d37d5046f66201912b40c224d4fe96b515231"))
	var key14, _ = crypto.ToECDSA(hexutil.MustDecode("0xa2ef6534312d205b045a94ec2e9d49191a6d17702671d51dd88a9e2837b612ce"))
	var key15, _ = crypto.ToECDSA(hexutil.MustDecode("0x9512765baac04484c19491feb59fe8ef8ba557e29e00100f3159c8ee35c89038"))
	var key16, _ = crypto.ToECDSA(hexutil.MustDecode("0x700717777c4b7ccdc8c79d6823cb3ea0356ac5e3822accdfa8539cf833caae15"))
	var key17, _ = crypto.ToECDSA(hexutil.MustDecode("0x838ab204f288b4673bbc603ac52c167e8b1c1392cdd96bc02b8fcfadec98cc26"))
	var key18, _ = crypto.ToECDSA(hexutil.MustDecode("0xbf6ba360590e69d1495ea8c0ab2f4a18ebbed7c4bbbe2d823a57719cd40df94f"))
	var key19, _ = crypto.ToECDSA(hexutil.MustDecode("0xd2f091785e9ca0ea2388cf90a046e87552e5cbb4492a9702b83aa32dddf821ac"))
	var key20, _ = crypto.ToECDSA(hexutil.MustDecode("0xad8b51bf6a35a934587413394ab453df559603f885ae3ff0e90c1d90c78153bd"))
	var key21, _ = crypto.ToECDSA(hexutil.MustDecode("0xa1ae301b83bfdd9e4a6ffb277896e5b4438725844fd44b5f733219f9f0a1402b"))
	var key22, _ = crypto.ToECDSA(hexutil.MustDecode("0x9bf39f28aa39153777677711b8ca8a733ffcdad9ec8831713a01d71fb3dbe184"))
	var key23, _ = crypto.ToECDSA(hexutil.MustDecode("0xee948e4413ce4e82ecb51fb6669f82d5af9b0ca4c31924514e6e844e8da46051"))
	var key24, _ = crypto.ToECDSA(hexutil.MustDecode("0xe9a94ddcec56059cffb6dd699011f2bb323293f90613385c8624839296b3d182"))
	var key25, _ = crypto.ToECDSA(hexutil.MustDecode("0xbccc8d4364e82a04ea2dc840ad6eeec6a2c35a51fb01943d58728da7bd4364dc"))
	var key26, _ = crypto.ToECDSA(hexutil.MustDecode("0xd8af1e1f98a3628e91e46888b02cb34b00fd72aee1946409a3435ea806f1ace8"))
	var key27, _ = crypto.ToECDSA(hexutil.MustDecode("0x0ba54f6d7c269ae7d115a17446abe7ba52293997de821d262a3d113fd694d85a"))
	var key28, _ = crypto.ToECDSA(hexutil.MustDecode("0x2666d00809c1ce11da2c7598d3ab54e1bb75263d9e25d8209568a1d5e7cf9cb7"))
	var key29, _ = crypto.ToECDSA(hexutil.MustDecode("0x204b21603d4a076bcdd34db298229f935198ea695964a3e156289728290e6240"))
	var key30, _ = crypto.ToECDSA(hexutil.MustDecode("0xd30f524750dbbef5833dddbcdcdaf6c7f4e43c777d5c468d124170838e83c59d"))
	var key31, _ = crypto.ToECDSA(hexutil.MustDecode("0xee3002a37a510360b0de793b45dd56a4a7a1df843e04cd991441854978b5154e"))
	var key32, _ = crypto.ToECDSA(hexutil.MustDecode("0x410605310cdc3bb8ffe3d472ebf183545e6a09f3b211616156d42d8ad2ee1218"))
	var key33, _ = crypto.ToECDSA(hexutil.MustDecode("0x3d47844c536f73c3558bf2e2238b13b327be2890cad6de60a3940337b8afa774"))
	var key34, _ = crypto.ToECDSA(hexutil.MustDecode("0x114a976408f9a71c581871a6b68f5006df44a178e86d0bc7659d591bb4e56da6"))
	var key35, _ = crypto.ToECDSA(hexutil.MustDecode("0x2c1108cc823ae0c5f496ad61eaab90d0677875ab1b2e0a55a89ecd87388fa9b3"))
	var key36, _ = crypto.ToECDSA(hexutil.MustDecode("0x5fb2462733c28810a8bc68712a08c201ed2b89e822bc7309834476cfa1857acc"))
	var key37, _ = crypto.ToECDSA(hexutil.MustDecode("0x595726750f55bc28a9a2e50f92a6c5fab42e738409cab0008299039c9966e0fe"))
	var key38, _ = crypto.ToECDSA(hexutil.MustDecode("0x6cb89990a3ecf4930470351f1d76a72525d2f702e47d490cc0cc087893d2664a"))
	var key39, _ = crypto.ToECDSA(hexutil.MustDecode("0x275b9ee8df6f2c2d02cd1fb5c343f139867104d5da6f8d82afc771e2d11a28e4"))
	var key40, _ = crypto.ToECDSA(hexutil.MustDecode("0x3a5abe2f6ee961774f0466fca8f986fb0a53c5560b0f239d2a7ce0c8cdb3e1d1"))
	var key41, _ = crypto.ToECDSA(hexutil.MustDecode("0x97bb9f4bb472e89fc5405dd5786ea1de87c5d79758020abb0bfcbf4c48daf9a2"))
	var key42, _ = crypto.ToECDSA(hexutil.MustDecode("0x8ae00c99180aa1f9a8fe2ee7e53aaaedc0e55ff71785f26afa725295d4de78ff"))
	var key43, _ = crypto.ToECDSA(hexutil.MustDecode("0x65e35bf4a097d543df427ec67c192765f6edcbdda54e1b5c0bb5e99507f6a269"))
	var key44, _ = crypto.ToECDSA(hexutil.MustDecode("0x5fa4c34c434e0ddf1f8b3917c3e9ecfcbc09df91237012ff8105bcba275e4b7a"))
	var key45, _ = crypto.ToECDSA(hexutil.MustDecode("0x52ebc273f1da45483d5c6d884f0b65dda891ffee0ea6cdb0c6af90e852984e96"))
	var key46, _ = crypto.ToECDSA(hexutil.MustDecode("0xadec518fdc716a50ffc89c1caf6d2861ffaf02f06656d9275bd5e5d25696c840"))
	var key47, _ = crypto.ToECDSA(hexutil.MustDecode("0xc08f211d4803a2ab2c6c3c0745671decba5564dbebf9d339ae3344d055fd8e1d"))
	var key48, _ = crypto.ToECDSA(hexutil.MustDecode("0x7cf47f78fc8a5a550eae7bc77fb2943dbf8b37dfc381b35ccc6970684ac6cbee"))
	var key49, _ = crypto.ToECDSA(hexutil.MustDecode("0x90659790bafc92adea25095ebfabaffaa5c4bf1d1cc375ab3eac825832181398"))
	var key50, _ = crypto.ToECDSA(hexutil.MustDecode("0xebcae9b7ee8dc6b813fd7aa567f94d9986a7d39a4997ebea3b09db85941cedb5"))
	var key51, _ = crypto.ToECDSA(hexutil.MustDecode("0xde2da353e4200f22614ce98b03e7af8e3f79afa4dcd40666b679525103301606"))
	var key52, _ = crypto.ToECDSA(hexutil.MustDecode("0xd850eca0a7ac46bc01962bcff3cd27fff5e32d401a4a4db3883a3f0e0bdf0933"))
	var key53, _ = crypto.ToECDSA(hexutil.MustDecode("0xabd5c47c0be87b046603c7016258e65c4345aab4a65dde1daf60f2fb3f6c7b0c"))
	var key54, _ = crypto.ToECDSA(hexutil.MustDecode("0xa6c6c5d4c176336017fe715861750fe674b6552755010bd1e8f24cbee19b9b59"))
	var key55, _ = crypto.ToECDSA(hexutil.MustDecode("0xf90b1f7c5e046b894c4c70d93ed74367c4ec260a5ee0051876c929b0a7e98dcb"))
	var key56, _ = crypto.ToECDSA(hexutil.MustDecode("0x15316697d6979fd22c5e3127e030484d94e4e7d78419200e669c3c2d0d9aa2e4"))
	var key57, _ = crypto.ToECDSA(hexutil.MustDecode("0xe86120a57411c86be886b6df0e80ee2061ddf322043ef565b47b548c8072ae31"))
	var key58, _ = crypto.ToECDSA(hexutil.MustDecode("0xe3e66700a59d00d5778c6b732d0e5f90b1516881a76ee4ad232aa7d06ba11e62"))
	var key59, _ = crypto.ToECDSA(hexutil.MustDecode("0xdd876a98e12334ef52ca2d8e149a20b5e085e7e8c6281c2aa60736915073f53f"))
	var key60, _ = crypto.ToECDSA(hexutil.MustDecode("0x30ab7160e3c2ec3884117f91e5189ca1c16af03af36a75cc0169b5f2e8163a88"))
	var key61, _ = crypto.ToECDSA(hexutil.MustDecode("0x2b2a8abbbc4624e33f737bbcf8d864999619e7eb2e92630c2ce3a773c438fba5"))
	var key62, _ = crypto.ToECDSA(hexutil.MustDecode("0xffcad1487800293d08dbe6355f60c170f41ae93906293f2a30c00568f6fb8717"))
	var key63, _ = crypto.ToECDSA(hexutil.MustDecode("0x1b70a964ce916046ba1d3def8fc7d004f213028113e2751e3cf0a12307a21e9f"))
	var key64, _ = crypto.ToECDSA(hexutil.MustDecode("0x4eb12e7c8a1d99a00dc99df7f8c162a929894ab2a638048627a08d9913c02efd"))
	var key65, _ = crypto.ToECDSA(hexutil.MustDecode("0x6944e7e33cafcd099c1b2a88e87e8f57b3fc48a0002c4d168737f55bca9dce6e"))
	var key66, _ = crypto.ToECDSA(hexutil.MustDecode("0x3c2bf03d642d85932ef2f6cc23259f8cc8782c60043c9d7ae58b096a02f9007b"))
	var key67, _ = crypto.ToECDSA(hexutil.MustDecode("0x161c258dc7afadecfe8f8106ab619690ac01f52f669a3b1f453540bf82c78b14"))
	var key68, _ = crypto.ToECDSA(hexutil.MustDecode("0x2941eea6ed3ee2166a0a8ce17f4a7e571cd8fd23ca270cc72839d7bafa955845"))
	var key69, _ = crypto.ToECDSA(hexutil.MustDecode("0x85449300aed219707b7801669597c082dd9e4c74633472610c0009e79422da53"))
	var key70, _ = crypto.ToECDSA(hexutil.MustDecode("0x79254268cf6352f9910405bb8c545aeeee8fbb61293e62663f81355b0ba3d86b"))
	var key71, _ = crypto.ToECDSA(hexutil.MustDecode("0x548c31c260958764c20b417b416223ab8623e8364d84fc8806f665eeb084d6d9"))
	var key72, _ = crypto.ToECDSA(hexutil.MustDecode("0x67edd0d0682a4e7e52575884db12d03b325bd6f8a5e18fb65b143f9f25df04aa"))
	var key73, _ = crypto.ToECDSA(hexutil.MustDecode("0xa262f0ea06bc87c7829dc6fe38d83f99c326976a13c5a0e94da13adc5d136307"))
	var key74, _ = crypto.ToECDSA(hexutil.MustDecode("0x9c655e84994bbe21ac8bade72dd6ebe37491c2986a8e4eb6ab2d007e3f130270"))
	var key75, _ = crypto.ToECDSA(hexutil.MustDecode("0xaf06f1eb84c1f3c9e14e35e6aed34703803fea21dcf628f2bc178325b33148c1"))
	var key76, _ = crypto.ToECDSA(hexutil.MustDecode("0x6aad2dc6d0442b14cfbdcc0ae207c3b88d31e1294606c287d1a6b0cc58670a5f"))
	var key77, _ = crypto.ToECDSA(hexutil.MustDecode("0x7e4e0e156ae1532e98bfda00d850fb476f43b81f09af080446fece7b7d8ac388"))
	var key78, _ = crypto.ToECDSA(hexutil.MustDecode("0xdaf1bfaeb6700798569f6d4815dff9fa6590856c27e6d9aa112cb06d5f21b525"))
	var key79, _ = crypto.ToECDSA(hexutil.MustDecode("0xcd76b0089ef74032756bf06fdd5903c8787957e943e0622f9b35ba84185cb675"))
	var key80, _ = crypto.ToECDSA(hexutil.MustDecode("0xa7b791cf6aae777a7954961b6b5c66b9326ee35ce379f17f554a55bd7cc91d5f"))
	var key81, _ = crypto.ToECDSA(hexutil.MustDecode("0xbb76f0697e3baca59a9b7e5ad449c912ad568d371a1c579bc92b8043479606fc"))
	var key82, _ = crypto.ToECDSA(hexutil.MustDecode("0x95f963b910d735ffd9dfd784ca683cc790aa40acdc5fa9d27fe8e84b6900528d"))
	var key83, _ = crypto.ToECDSA(hexutil.MustDecode("0xe8fa32934f5e6c895f1586114cd3c84caf04f01efab47bcd8bc2efbd743e6dfb"))
	var key84, _ = crypto.ToECDSA(hexutil.MustDecode("0x030076ec9dd721e30c743bea5f0019e80428a97fe900b7245b968c6a6f774313"))
	var key85, _ = crypto.ToECDSA(hexutil.MustDecode("0xb7001594545d584baab7d6b056d3a39c325db2b450787329ed916dd26fa32260"))
	var key86, _ = crypto.ToECDSA(hexutil.MustDecode("0xd2a748ae6cfa9156c94754320b22b35b295dfe779887098386b2cea72f4d0dd2"))
	var key87, _ = crypto.ToECDSA(hexutil.MustDecode("0x258bb5bc5a87c5aebdadb7f83f39029242f0650bc52fc4cc482c89412532db2b"))
	var key88, _ = crypto.ToECDSA(hexutil.MustDecode("0x20ecd5168010ed59a184fd5b5ac528b9fff70ed103ea58c6f35f0137854969b8"))
	var key89, _ = crypto.ToECDSA(hexutil.MustDecode("0x1a517bb0a54425b29b032d4cc28423224963cda64b50ef9731b429660c8da129"))
	var key90, _ = crypto.ToECDSA(hexutil.MustDecode("0xee61ccd710d6dabb81849a1ca2a9369c5484b65517ea80c7772135c55aa9f147"))
	var key91, _ = crypto.ToECDSA(hexutil.MustDecode("0xe906f9f17d6511279053a95feca4a55ef59f97f2596790163ade48067d20238f"))
	var key92, _ = crypto.ToECDSA(hexutil.MustDecode("0x3c8cd94010f44ac08efaae8a31e726bb4fc95564ee7c2e03e7a43ee43f31d6ed"))
	var key93, _ = crypto.ToECDSA(hexutil.MustDecode("0x58ed1a9bde738f0a089ae0c7e18bf77d2489e75f282c1af6e16e0b86fc30c41e"))
	var key94, _ = crypto.ToECDSA(hexutil.MustDecode("0x0be230f443fdc623254292438f0c0f86ba0b799c89d15c9383aa3a64d99628e7"))
	var key95, _ = crypto.ToECDSA(hexutil.MustDecode("0x26a53f1c2b8fff8cfc6ca691777d1c84eb6bef60a3b3c26233c1b12faf653584"))
	var key96, _ = crypto.ToECDSA(hexutil.MustDecode("0x79a6be35f01db3d6a2b659b49f04505a0ce1abf3770494a7a24a83642ae8a675"))
	var key97, _ = crypto.ToECDSA(hexutil.MustDecode("0x54cd7ec556aeeb6f538fe1b7523a86425e520e288be516c95909582e79012bc3"))
	var key98, _ = crypto.ToECDSA(hexutil.MustDecode("0x675e4d1f766213db0791a274e748e858d091460fb4a9222e4b75380f8ecabcdc"))
	var key99, _ = crypto.ToECDSA(hexutil.MustDecode("0x4203c438d4e94bf4a595794b5f5c2882f959face730abb7a7b8acb462c8e138d"))

	var keys = [100]*ecdsa.PrivateKey{
		key00, key01, key02, key03, key04, key05, key06, key07, key08, key09,
		key10, key11, key12, key13, key14, key15, key16, key17, key18, key19,
		key20, key21, key22, key23, key24, key25, key26, key27, key28, key29,
		key30, key31, key32, key33, key34, key35, key36, key37, key38, key39,
		key40, key41, key42, key43, key44, key45, key46, key47, key48, key49,
		key50, key51, key52, key53, key54, key55, key56, key57, key58, key59,
		key60, key61, key62, key63, key64, key65, key66, key67, key68, key69,
		key70, key71, key72, key73, key74, key75, key76, key77, key78, key79,
		key80, key81, key82, key83, key84, key85, key86, key87, key88, key89,
		key90, key91, key92, key93, key94, key95, key96, key97, key98, key99,
	}

	if n > len(keys) {
		panic(errors.New("validator num is too large"))
	}

	return keys[n-1]
}
