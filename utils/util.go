package utils

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

var startBase uint32 = 12000

// GetUnusedNetAddr return array of n unused ports starting with base port
// NB: addresses 1-1024 are reserved for non-root users;
func GetUnusedNetAddr(n int, t testing.TB) []string {
	idx := int(0)
	base := atomic.AddUint32(&startBase, 100)
	addresses := make([]string, n)
	for i := int(base); i < 65536; i++ {
		addrStr := "127.0.0.1:" + strconv.Itoa(i)
		addr, err := net.ResolveTCPAddr("tcp", addrStr)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			continue
		}
		if res := func() []string {
			defer func() {
				if err := l.Close(); err != nil {
					t.Fatal(err)
				}
			}()
			t.Logf("Unused port %s is chosen", addrStr)
			addresses[idx] = addrStr
			idx++
			if idx == n {
				return addresses
			}
			return nil
		}(); res != nil {
			return res
		}
	}
	t.Fatalf("No free port left!!!")
	return addresses
}

// HashFromHex converts hex string to bytes.
func HashFromHex(s string) []byte {
	if len(s) > 1 {
		if s[0:2] == "0x" || s[0:2] == "0X" {
			s = s[2:]
		}
	}
	if len(s)%2 == 1 {
		s = "0" + s
	}
	h, _ := hex.DecodeString(s)
	return h
}

// FreePort gets free network port on host.
func FreePort(network string) (port uint16) {
	addr, err := net.ResolveTCPAddr(network, "localhost:0")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP(network, addr)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := l.Close(); err != nil {
			panic(err)
		}
	}()
	return uint16(l.Addr().(*net.TCPAddr).Port)
}

const (
	// number of bits in a big.Word
	wordBits = 32 << (uint64(^big.Word(0)) >> 63)
	// number of bytes in a big.Word
	wordBytes = wordBits / 8
)

// ReadBits encodes the absolute value of bigint as big-endian bytes. Callers must ensure
// that buf has enough space. If buf is too short the result will be incomplete.
func ReadBits(bigint *big.Int, buf []byte) {
	i := len(buf)
	for _, d := range bigint.Bits() {
		for j := 0; j < wordBytes && i > 0; j++ {
			i--
			buf[i] = byte(d)
			d >>= 8
		}
	}
}

// PaddedBigBytes encodes a big integer as a big-endian byte slice. The length
// of the slice is at least n bytes.
func PaddedBigBytes(bigint *big.Int, n int) []byte {
	if bigint.BitLen()/8 >= n {
		return bigint.Bytes()
	}
	ret := make([]byte, n)
	ReadBits(bigint, ret)
	return ret
}

// NameOf returns human readable string representation.
func NameOf(p idx.StakerID) string {
	if name := hash.GetNodeName(p); len(name) > 0 {
		return name
	}

	return fmt.Sprintf("%d", p)
}
