package app

import (
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

// BlocksMissed is information about missed blocks from a staker
type BlocksMissed struct {
	Num    idx.Block
	Period inter.Timestamp
}
