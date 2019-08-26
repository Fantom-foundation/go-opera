// Copyright 2017 The go-ethereum Authors
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

package cryptoaddr

import (
	"testing"

	eth "github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
)

var (
	testmsg     = eth.HexToHash("ce0677bb30baa8cf067c88db9811f4333d131bf8bcf12fe7065d211dce971008")
	testsig     = common.Hex2Bytes("90f27b8b488db00b00606796d2987f6a5f59ae62ea05effe84fef5b8b0e549984a691139ad57a3f0b906637673aa2f63d1f55cb1a69199d4009eea23ceaddc9301")
	testaddress = eth.HexToAddress("5ebcafc66e93c8d4404ab041e3bb1702780f9f49f224ce14674c75950fcb2567")
	testpubkey  = common.Hex2Bytes("04e32df42865e97135acfb65f3bae71bdc86f4d49150ad6a440b6f15878109880a0a2b2667f7e725ceea70c673093bf67663e0312623c8e091b13cf2c0f11ef652")
)

func TestRecoverAddr(t *testing.T) {
	addr, err := RecoverAddr(testmsg, testsig)
	if err != nil {
		t.Fatalf("recover error: %s", err)
	}
	if addr != testaddress {
		t.Errorf("address mismatch: want: %x have: %x", testaddress, addr)
	}
}

func TestVerifySignature(t *testing.T) {
	sig := testsig
	if !VerifySignature(testaddress, testmsg, testsig) {
		t.Errorf("can't verify signature")
	}

	if VerifySignature(eth.Address{}, testmsg, sig) {
		t.Errorf("signature valid with no key")
	}
	if VerifySignature(testaddress, eth.Hash{}, sig) {
		t.Errorf("signature valid with no message")
	}
	if VerifySignature(testaddress, testmsg, nil) {
		t.Errorf("nil signature valid")
	}
	if VerifySignature(testaddress, testmsg, append(common.CopyBytes(sig), 1, 2, 3)) {
		t.Errorf("signature valid with extra bytes at the end")
	}
	if VerifySignature(testaddress, testmsg, sig[:len(sig)-2]) {
		t.Errorf("signature valid even though it's incomplete")
	}
	wrongaddr := common.CopyBytes(testaddress.Bytes())
	wrongaddr[10]++
	if VerifySignature(eth.BytesToAddress(wrongaddr), testmsg, sig) {
		t.Errorf("signature valid with with wrong addr")
	}
}

// This test checks that VerifySignature rejects malleable signatures with s > N/2.
func TestVerifySignatureMalleable(t *testing.T) {
	sig := common.Hex2Bytes("0x638a54215d80a6713c8d523a6adc4e6e73652d859103a36b700851cb0e61b66b8ebfc1a610c57d732ec6e0a8f06a9a7a28df5051ece514702ff9cdff0b11f454")
	addr := eth.HexToAddress("0x03ca634cae0d49acb401d8a4c6b6fe8c55b70d115bf400769cc1400f3258cd3138")
	msg := eth.HexToHash("0xd301ce462d3e639518f482c7f03821fec1e602018630ce621e1e7851c12343a6")
	if VerifySignature(addr, msg, sig) {
		t.Error("VerifySignature returned true for malleable signature")
	}
}

func BenchmarkRecoverSignature(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := RecoverAddr(testmsg, testsig); err != nil {
			b.Fatal("ecrecover error", err)
		}
	}
}

func BenchmarkVerifySignature(b *testing.B) {
	sig := testsig[:len(testsig)-1] // remove recovery id
	for i := 0; i < b.N; i++ {
		if !crypto.VerifySignature(testpubkey, testmsg.Bytes(), sig) {
			b.Fatal("verify error")
		}
	}
}
