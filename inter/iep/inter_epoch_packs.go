package iep

import (
	"github.com/cyberbono3/go-opera/inter"
	"github.com/cyberbono3/go-opera/inter/ier"
)

type LlrEpochPack struct {
	Votes  []inter.LlrSignedEpochVote
	Record ier.LlrIdxFullEpochRecord
}
