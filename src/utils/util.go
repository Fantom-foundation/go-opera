package utils

import (
	"encoding/hex"
	"net"
	"strconv"
	"sync/atomic"
	"testing"
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
