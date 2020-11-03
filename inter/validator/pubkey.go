package validator

import (
	"encoding/hex"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

const (
	FakePassword = "fakepassword"
)

type PubKey struct {
	Raw  []byte
	Type string
}

func (pk *PubKey) Empty() bool {
	return len(pk.Raw) == 0 && len(pk.Type) == 0
}

func (pk *PubKey) String() string {
	return common.Bytes2Hex(pk.Raw) + "@" + pk.Type
}

func (pk *PubKey) Bytes() []byte {
	return []byte(string(pk.Raw) + pk.Type)
}

func PubKeyFromString(str string) (PubKey, error) {
	parts := strings.Split(str, "@")
	if len(parts) != 2 {
		return PubKey{}, errors.New("malformed pubkey: expected hex@type")
	}
	raw, err := hex.DecodeString(parts[0])
	if err != nil {
		return PubKey{}, errors.Wrap(err, "pubkey decoding error")
	}
	return PubKey{raw, parts[1]}, nil
}

// MarshalText returns the hex representation of a.
func (pk *PubKey) MarshalText() ([]byte, error) {
	return []byte(pk.String()), nil
}

// UnmarshalText parses a hash in hex syntax.
func (pk *PubKey) UnmarshalText(input []byte) error {
	res, err := PubKeyFromString(string(input))
	if err != nil {
		return err
	}
	*pk = res
	return nil
}
