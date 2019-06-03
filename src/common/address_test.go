package common

import (
	"bytes"
	"testing"
)

var (
	// expected length of human readable address
	expected_address_len = 11
)

func TestAddress(t *testing.T) {
	address := FakeAddress()
	str := address.String()
	if len(str) != expected_address_len {
		t.Errorf("Expected length of human readable address to be %d but got %d! - '%s'",
			expected_address_len, len(str), str)
	}

	abytes := []byte{0x55, 0xAA, 0xDD}
	address = BytesToAddress(abytes)
	if !bytes.Equal(abytes, address.Bytes()[:3]) {
		t.Errorf("abytes is not equal to address.Bytes()!")
	}
}
