package iep

import (
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/ier"
)

type LlrEpochPack struct {
	Votes  []inter.LlrSignedEpochVote
	Record ier.LlrIdxFullEpochRecord
}
