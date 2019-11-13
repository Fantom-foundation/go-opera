package gossip

import (
	"math/big"

	"github.com/Fantom-foundation/go-lachesis/inter"
)

type EpochStats struct {
	Start    inter.Timestamp
	End      inter.Timestamp
	TotalFee *big.Int
}
