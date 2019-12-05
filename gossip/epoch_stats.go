package gossip

import (
	"math/big"

	"github.com/Fantom-foundation/go-lachesis/inter"
)

// EpochStats stores general statistics for an epoch
type EpochStats struct {
	Start    inter.Timestamp
	End      inter.Timestamp
	TotalFee *big.Int
}
