package validatorpk

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

const (
	FakePassword = "fakepassword"
)

type PubKey struct {
	Type uint8
	Raw  []byte
}

var Types = struct {
	Secp256k1 uint8
}{
	Secp256k1: 0xc0,
}

func (pk *PubKey) Empty() bool {
	return len(pk.Raw) == 0 && pk.Type == 0
}

func (pk *PubKey) String() string {
	return "0x" + common.Bytes2Hex(pk.Bytes())
}

func (pk *PubKey) Bytes() []byte {
	return append([]byte{pk.Type}, pk.Raw...)
}

func FromString(str string) (PubKey, error) {
	return FromBytes(common.FromHex(str))
}

func FromBytes(b []byte) (PubKey, error) {
	if len(b) == 0 {
		return PubKey{}, errors.New("empty pubkey")
	}
	return PubKey{b[0], b[1:]}, nil
}

// MarshalText returns the hex representation of a.
func (pk *PubKey) MarshalText() ([]byte, error) {
	return []byte(pk.String()), nil
}

// UnmarshalText parses a hash in hex syntax.
func (pk *PubKey) UnmarshalText(input []byte) error {
	res, err := FromString(string(input))
	if err != nil {
		return err
	}
	*pk = res
	return nil
}
