package app

import (
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
)

// BlocksMissed is information about missed blocks from a staker
type BlocksMissed struct {
	Num    idx.Block
	Period inter.Timestamp
}
