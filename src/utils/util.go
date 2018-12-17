package utils

import (
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"testing"
)

// source: https://gist.github.com/montanaflynn/b59c058ce2adc18f31d6
func GetUnusedNetAddr(t testing.TB) string {
	// Create a new server without specifying a port
	// which will result in an open port being chosen
	server, err := net.Listen("tcp", ":0")
	// If there's an error it likely means no ports
	// are available or something else prevented finding
	// an open port
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer server.Close()
	hostString := server.Addr().String()
	// Split the host from the port
	_, portString, err := net.SplitHostPort(hostString)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	// Return the port as an int
	port, err := strconv.Atoi(portString)
	return fmt.Sprintf("127.0.0.1:%d", port)
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
