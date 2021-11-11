package inter

import (
	"crypto/sha256"

	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
)

type LlrBlockVotes struct {
	Start idx.Block
	Epoch idx.Epoch
	Votes []hash.Hash
}

func (bvs LlrBlockVotes) LastBlock() idx.Block {
	return bvs.Start + idx.Block(len(bvs.Votes)) - 1
}

type LlrEpochVote struct {
	Epoch idx.Epoch
	Vote  hash.Hash
}

type LlrSignedBlockVotes struct {
	Signed                       SignedEventLocator
	TxsAndMisbehaviourProofsHash hash.Hash
	EpochVoteHash                hash.Hash
	Val                          LlrBlockVotes
}

type LlrSignedEpochVote struct {
	Signed                       SignedEventLocator
	TxsAndMisbehaviourProofsHash hash.Hash
	BlockVotesHash               hash.Hash
	Val                          LlrEpochVote
}

func AsSignedBlockVotes(e EventPayloadI) LlrSignedBlockVotes {
	return LlrSignedBlockVotes{
		Signed:                       AsSignedEventLocator(e),
		TxsAndMisbehaviourProofsHash: hash.Of(CalcTxHash(e.Txs()).Bytes(), CalcMisbehaviourProofsHash(e.MisbehaviourProofs()).Bytes()),
		EpochVoteHash:                e.EpochVote().Hash(),
		Val:                          e.BlockVotes(),
	}
}

func AsSignedEpochVote(e EventPayloadI) LlrSignedEpochVote {
	return LlrSignedEpochVote{
		Signed:                       AsSignedEventLocator(e),
		TxsAndMisbehaviourProofsHash: hash.Of(CalcTxHash(e.Txs()).Bytes(), CalcMisbehaviourProofsHash(e.MisbehaviourProofs()).Bytes()),
		BlockVotesHash:               e.BlockVotes().Hash(),
		Val:                          e.EpochVote(),
	}
}

func (r SignedEventLocator) Size() uint64 {
	return uint64(len(r.Sig)) + 3*32 + 4*4
}

func (bvs LlrSignedBlockVotes) Size() uint64 {
	return bvs.Signed.Size() + uint64(len(bvs.Val.Votes))*32 + 32*2 + 8 + 4
}

func (ers LlrEpochVote) Hash() hash.Hash {
	hasher := sha256.New()
	hasher.Write(ers.Epoch.Bytes())
	hasher.Write(ers.Vote.Bytes())
	return hash.BytesToHash(hasher.Sum(nil))
}

func (bvs LlrBlockVotes) Hash() hash.Hash {
	hasher := sha256.New()
	hasher.Write(bvs.Start.Bytes())
	hasher.Write(bvs.Epoch.Bytes())
	hasher.Write(bigendian.Uint32ToBytes(uint32(len(bvs.Votes))))
	for _, bv := range bvs.Votes {
		hasher.Write(bv.Bytes())
	}
	return hash.BytesToHash(hasher.Sum(nil))
}

func (bvs LlrSignedBlockVotes) CalcPayloadHash() hash.Hash {
	return hash.Of(bvs.TxsAndMisbehaviourProofsHash.Bytes(), hash.Of(bvs.EpochVoteHash.Bytes(), bvs.Val.Hash().Bytes()).Bytes())
}

func (ev LlrSignedEpochVote) CalcPayloadHash() hash.Hash {
	return hash.Of(ev.TxsAndMisbehaviourProofsHash.Bytes(), hash.Of(ev.Val.Hash().Bytes(), ev.BlockVotesHash.Bytes()).Bytes())
}

func (ev LlrSignedEpochVote) Size() uint64 {
	return ev.Signed.Size() + 32 + 32*2 + 4 + 4
}
