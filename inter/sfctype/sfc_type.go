package sfctype

import (
	"math/big"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"

	"github.com/Fantom-foundation/go-opera/inter/validator"
)

var (
	// DoublesignBit is set if validator has a confirmed pair of fork events
	DoublesignBit = uint64(1 << 7)
	OkStatus      = uint64(0)
)

// SfcValidator is the node-side representation of SFC validator
type SfcValidator struct {
	Weight *big.Int
	PubKey validator.PubKey
}

// SfcValidatorAndID is pair SfcValidator + ValidatorID
type SfcValidatorAndID struct {
	ValidatorID idx.ValidatorID
	Validator   SfcValidator
}
