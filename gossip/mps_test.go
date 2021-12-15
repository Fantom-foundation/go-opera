package gossip

import (
	"testing"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-opera/eventcheck/basiccheck"
	"github.com/Fantom-foundation/go-opera/eventcheck/heavycheck"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/logger"
)

func copyBvs(bvs inter.LlrSignedBlockVotes) inter.LlrSignedBlockVotes {
	cp := make([]hash.Hash, 0, len(bvs.Val.Votes))
	for _, v := range bvs.Val.Votes {
		cp = append(cp, v)
	}
	bvs.Val.Votes = cp
	return bvs
}

func copyMP(mp inter.MisbehaviourProof) inter.MisbehaviourProof {
	if mp.EventsDoublesign != nil {
		cp := *mp.EventsDoublesign
		mp.EventsDoublesign = &cp
	}
	if mp.BlockVoteDoublesign != nil {
		cp := *mp.BlockVoteDoublesign
		mp.BlockVoteDoublesign = &cp
		for i := range mp.BlockVoteDoublesign.Pair {
			mp.BlockVoteDoublesign.Pair[i] = copyBvs(mp.BlockVoteDoublesign.Pair[i])
		}
	}
	if mp.WrongBlockVote != nil {
		cp := *mp.WrongBlockVote
		mp.WrongBlockVote = &cp
		for i := range mp.WrongBlockVote.Pals {
			mp.WrongBlockVote.Pals[i] = copyBvs(mp.WrongBlockVote.Pals[i])
		}
	}
	if mp.EpochVoteDoublesign != nil {
		cp := *mp.EpochVoteDoublesign
		mp.EpochVoteDoublesign = &cp
	}
	if mp.WrongEpochVote != nil {
		cp := *mp.WrongEpochVote
		mp.WrongEpochVote = &cp
	}
	return mp
}

func TestMisbehaviourProofsEventsDoublesign(t *testing.T) {
	logger.SetTestMode(t)
	require := require.New(t)

	const validatorsNum = 3

	startEpoch := idx.Epoch(basiccheck.MaxLiableEpochs)
	env := newTestEnv(startEpoch, validatorsNum)
	defer env.Close()

	// move epoch further
	_, err := env.ApplyTxs(nextEpoch, env.Transfer(1, 1, common.Big0))
	require.NoError(err)
	_, err = env.ApplyTxs(nextEpoch, env.Transfer(1, 1, common.Big0))
	require.NoError(err)
	require.Greater(env.store.GetEpoch(), startEpoch)

	correctMp := inter.MisbehaviourProof{
		EventsDoublesign: &inter.EventsDoublesign{
			Pair: [2]inter.SignedEventLocator{
				{
					Locator: inter.EventLocator{
						Epoch:   startEpoch,
						Seq:     1,
						Lamport: 1,
						Creator: 1,
					},
				},
				{
					Locator: inter.EventLocator{
						Epoch:   startEpoch,
						Seq:     1,
						Lamport: 2,
						Creator: 1,
					},
				},
			},
		},
	}

	// sign
	for i, p := range correctMp.EventsDoublesign.Pair {
		sig, err := env.signer.Sign(env.pubkeys[0], p.Locator.HashToSign().Bytes())
		require.NoError(err)
		copy(correctMp.EventsDoublesign.Pair[i].Sig[:], sig)
	}

	tooLateMp := copyMP(correctMp)
	tooLateMp.EventsDoublesign.Pair[0].Locator.Epoch = 1
	tooLateMp.EventsDoublesign.Pair[1].Locator.Epoch = 1
	err = env.ApplyMPs(nextEpoch, tooLateMp)
	require.ErrorIs(err, basiccheck.ErrMPTooLate)

	wrongEpochMp := copyMP(correctMp)
	wrongEpochMp.EventsDoublesign.Pair[0].Locator.Epoch--
	err = env.ApplyMPs(nextEpoch, wrongEpochMp)
	require.ErrorIs(err, basiccheck.ErrNoCrimeInMP)

	wrongSeqMp := copyMP(correctMp)
	wrongSeqMp.EventsDoublesign.Pair[0].Locator.Seq--
	err = env.ApplyMPs(nextEpoch, wrongSeqMp)
	require.ErrorIs(err, basiccheck.ErrNoCrimeInMP)

	wrongCreatorMp := copyMP(correctMp)
	wrongCreatorMp.EventsDoublesign.Pair[0].Locator.Creator++
	err = env.ApplyMPs(nextEpoch, wrongCreatorMp)
	require.ErrorIs(err, basiccheck.ErrWrongCreatorMP)

	sameEventsMp := copyMP(correctMp)
	sameEventsMp.EventsDoublesign.Pair[0] = sameEventsMp.EventsDoublesign.Pair[1]
	err = env.ApplyMPs(nextEpoch, sameEventsMp)
	require.ErrorIs(err, basiccheck.ErrNoCrimeInMP)

	for i := range correctMp.EventsDoublesign.Pair {
		wrongSigMp := copyMP(correctMp)
		wrongSigMp.EventsDoublesign.Pair[i].Sig[0]++
		err = env.ApplyMPs(nextEpoch, wrongSigMp)
		require.ErrorIs(err, heavycheck.ErrWrongEventSig)
	}

	for i := range correctMp.EventsDoublesign.Pair {
		wrongSigMp := copyMP(correctMp)
		wrongSigMp.EventsDoublesign.Pair[i].Locator.BaseHash = hash.HexToHash("0x10")
		err = env.ApplyMPs(nextEpoch, wrongSigMp)
		require.ErrorIs(err, heavycheck.ErrWrongEventSig)
	}

	wrongAuthEpochMp := copyMP(correctMp)
	wrongAuthEpochMp.EventsDoublesign.Pair[0].Locator.Epoch = startEpoch - 1
	wrongAuthEpochMp.EventsDoublesign.Pair[1].Locator.Epoch = startEpoch - 1
	err = env.ApplyMPs(nextEpoch, wrongAuthEpochMp)
	require.ErrorIs(err, heavycheck.ErrUnknownEpochEventLocator)

	err = env.ApplyMPs(nextEpoch, correctMp)
	require.NoError(err)
	require.Equal(idx.Validator(2), env.store.GetValidators().Len())
	require.False(env.store.GetValidators().Exists(1))
}

func TestMisbehaviourProofsBlockVoteDoublesign(t *testing.T) {
	logger.SetTestMode(t)
	require := require.New(t)

	const validatorsNum = 3

	startEpoch := idx.Epoch(basiccheck.MaxLiableEpochs)
	env := newTestEnv(basiccheck.MaxLiableEpochs, validatorsNum)
	defer env.Close()

	// move epoch further
	_, err := env.ApplyTxs(nextEpoch, env.Transfer(1, 1, common.Big0))
	require.NoError(err)
	_, err = env.ApplyTxs(nextEpoch, env.Transfer(1, 1, common.Big0))
	require.NoError(err)
	require.Greater(env.store.GetEpoch(), startEpoch)

	correctMp := inter.MisbehaviourProof{
		BlockVoteDoublesign: &inter.BlockVoteDoublesign{
			Block: env.store.GetLatestBlockIndex(),
			Pair: [2]inter.LlrSignedBlockVotes{
				{
					Val: inter.LlrBlockVotes{
						Start: env.store.GetLatestBlockIndex() - 1,
						Epoch: startEpoch,
						Votes: []hash.Hash{
							hash.Zero,
							hash.HexToHash("0x01"),
						},
					},
				},
				{
					Val: inter.LlrBlockVotes{
						Start: env.store.GetLatestBlockIndex(),
						Epoch: startEpoch,
						Votes: []hash.Hash{
							hash.HexToHash("0x02"),
						},
					},
				},
			},
		},
	}

	// sign
	for i, p := range correctMp.BlockVoteDoublesign.Pair {
		e := &inter.MutableEventPayload{}
		e.SetVersion(1)
		e.SetBlockVotes(p.Val)
		e.SetEpoch(env.store.GetEpoch() - idx.Epoch(i))
		e.SetCreator(2)
		e.SetPayloadHash(inter.CalcPayloadHash(e))

		sig, err := env.signer.Sign(env.pubkeys[1], e.HashToSign().Bytes())
		require.NoError(err)
		sSig := inter.Signature{}
		copy(sSig[:], sig)
		e.SetSig(sSig)

		correctMp.BlockVoteDoublesign.Pair[i] = inter.AsSignedBlockVotes(e)
	}

	sameVotesMp := copyMP(correctMp)
	sameVotesMp.BlockVoteDoublesign.Pair[0].Val.Votes[1] = sameVotesMp.BlockVoteDoublesign.Pair[1].Val.Votes[0]
	err = env.ApplyMPs(nextEpoch, sameVotesMp)
	require.ErrorIs(err, basiccheck.ErrNoCrimeInMP)

	differentEpochsMp := sameVotesMp
	differentEpochsMp.BlockVoteDoublesign.Pair[0].Val.Epoch++
	err = env.ApplyMPs(nextEpoch, sameVotesMp)
	require.ErrorIs(err, heavycheck.ErrWrongPayloadHash)

	for i := range correctMp.BlockVoteDoublesign.Pair {
		malformedVotesMp := copyMP(correctMp)
		malformedVotesMp.BlockVoteDoublesign.Pair[i].Val.Votes = []hash.Hash{}
		err = env.ApplyMPs(nextEpoch, malformedVotesMp)
		require.ErrorIs(err, basiccheck.MalformedBVs)
	}
	for i := range correctMp.BlockVoteDoublesign.Pair {
		malformedVotesMp := copyMP(correctMp)
		malformedVotesMp.BlockVoteDoublesign.Pair[i].Val.Start -= 2
		err = env.ApplyMPs(nextEpoch, malformedVotesMp)
		require.ErrorIs(err, basiccheck.ErrWrongMP)
	}
	malformedVotesMp := copyMP(correctMp)
	malformedVotesMp.BlockVoteDoublesign.Pair[1].Val.Start--
	err = env.ApplyMPs(nextEpoch, malformedVotesMp)
	require.ErrorIs(err, basiccheck.ErrWrongMP)

	for i := range correctMp.BlockVoteDoublesign.Pair {
		tooLateMp := copyMP(correctMp)
		tooLateMp.BlockVoteDoublesign.Pair[i].Val.Epoch = 1
		err = env.ApplyMPs(nextEpoch, tooLateMp)
		require.ErrorIs(err, basiccheck.ErrMPTooLate)
	}

	for i := range correctMp.BlockVoteDoublesign.Pair {
		futureEpochMp := copyMP(correctMp)
		futureEpochMp.BlockVoteDoublesign.Pair[i].Signed.Locator.Epoch = futureEpochMp.BlockVoteDoublesign.Pair[i].Val.Epoch - 1
		err = env.ApplyMPs(nextEpoch, futureEpochMp)
		require.ErrorIs(err, basiccheck.FutureBVsEpoch)
	}

	wrongCreatorMp := copyMP(correctMp)
	wrongCreatorMp.BlockVoteDoublesign.Pair[0].Signed.Locator.Creator++
	err = env.ApplyMPs(nextEpoch, wrongCreatorMp)
	require.ErrorIs(err, basiccheck.ErrWrongCreatorMP)

	sameEventsMp := copyMP(correctMp)
	sameEventsMp.BlockVoteDoublesign.Pair[0] = sameEventsMp.BlockVoteDoublesign.Pair[1]
	err = env.ApplyMPs(nextEpoch, sameEventsMp)
	require.ErrorIs(err, basiccheck.ErrNoCrimeInMP)

	for i := range correctMp.BlockVoteDoublesign.Pair {
		wrongSigMp := copyMP(correctMp)
		wrongSigMp.BlockVoteDoublesign.Pair[i].Signed.Sig[0]++
		err = env.ApplyMPs(nextEpoch, wrongSigMp)
		require.ErrorIs(err, heavycheck.ErrWrongEventSig)
	}

	for i := range correctMp.BlockVoteDoublesign.Pair {
		wrongSigMp := copyMP(correctMp)
		wrongSigMp.BlockVoteDoublesign.Pair[i].Signed.Locator.BaseHash = hash.HexToHash("0x10")
		err = env.ApplyMPs(nextEpoch, wrongSigMp)
		require.ErrorIs(err, heavycheck.ErrWrongEventSig)
	}

	wrongAuthEpochMp := copyMP(correctMp)
	wrongAuthEpochMp.BlockVoteDoublesign.Pair[0].Signed.Locator.Epoch = startEpoch - 1
	wrongAuthEpochMp.BlockVoteDoublesign.Pair[0].Val.Epoch = startEpoch - 1
	wrongAuthEpochMp.BlockVoteDoublesign.Pair[1].Signed.Locator.Epoch = startEpoch - 1
	wrongAuthEpochMp.BlockVoteDoublesign.Pair[1].Val.Epoch = startEpoch - 1
	err = env.ApplyMPs(nextEpoch, wrongAuthEpochMp)
	require.ErrorIs(err, heavycheck.ErrUnknownEpochBVs)

	err = env.ApplyMPs(nextEpoch, correctMp)
	require.NoError(err)
	require.Equal(idx.Validator(2), env.store.GetValidators().Len())
	require.False(env.store.GetValidators().Exists(2))
}

func TestMisbehaviourProofsWrongBlockVote(t *testing.T) {
	logger.SetTestMode(t)
	require := require.New(t)

	const validatorsNum = 3

	startEpoch := idx.Epoch(basiccheck.MaxLiableEpochs)
	env := newTestEnv(startEpoch, validatorsNum)
	defer env.Close()

	// move epoch further
	_, err := env.ApplyTxs(nextEpoch, env.Transfer(1, 1, common.Big0))
	require.NoError(err)
	_, err = env.ApplyTxs(nextEpoch, env.Transfer(1, 1, common.Big0))
	require.NoError(err)
	require.Greater(env.store.GetEpoch(), startEpoch)

	correctMp := inter.MisbehaviourProof{
		WrongBlockVote: &inter.WrongBlockVote{
			Block: env.store.GetLatestBlockIndex(),
			Pals: [2]inter.LlrSignedBlockVotes{
				{
					Val: inter.LlrBlockVotes{
						Start: env.store.GetLatestBlockIndex() - 1,
						Epoch: startEpoch,
						Votes: []hash.Hash{
							hash.Zero,
							hash.HexToHash("0x01"),
						},
					},
				},
				{
					Val: inter.LlrBlockVotes{
						Start: env.store.GetLatestBlockIndex(),
						Epoch: startEpoch + 1,
						Votes: []hash.Hash{
							hash.HexToHash("0x01"),
						},
					},
				},
			},
			WrongEpoch: false,
		},
	}

	sign := func(mp *inter.MisbehaviourProof) {
		// sign
		for i, p := range mp.WrongBlockVote.Pals {
			e := &inter.MutableEventPayload{}
			e.SetVersion(1)
			e.SetBlockVotes(p.Val)
			e.SetEpoch(env.store.GetEpoch() - idx.Epoch(i))
			e.SetCreator(idx.ValidatorID(i + 1))
			e.SetPayloadHash(inter.CalcPayloadHash(e))

			sig, err := env.signer.Sign(env.pubkeys[i], e.HashToSign().Bytes())
			require.NoError(err)
			sSig := inter.Signature{}
			copy(sSig[:], sig)
			e.SetSig(sSig)

			mp.WrongBlockVote.Pals[i] = inter.AsSignedBlockVotes(e)
		}
	}
	sign(&correctMp)

	differentVotesMp := copyMP(correctMp)
	differentVotesMp.WrongBlockVote.Pals[0].Val.Votes[1] = hash.HexToHash("0x02")
	err = env.ApplyMPs(nextEpoch, differentVotesMp)
	require.ErrorIs(err, basiccheck.ErrNoCrimeInMP)

	for i := range correctMp.WrongBlockVote.Pals {
		malformedVotesMp := copyMP(correctMp)
		malformedVotesMp.WrongBlockVote.Pals[i].Val.Votes = []hash.Hash{}
		err = env.ApplyMPs(nextEpoch, malformedVotesMp)
		require.ErrorIs(err, basiccheck.MalformedBVs)
	}
	for i := range correctMp.WrongBlockVote.Pals {
		malformedVotesMp := copyMP(correctMp)
		malformedVotesMp.WrongBlockVote.Pals[i].Val.Start -= 2
		err = env.ApplyMPs(nextEpoch, malformedVotesMp)
		require.ErrorIs(err, basiccheck.ErrWrongMP)
	}
	malformedVotesMp := copyMP(correctMp)
	malformedVotesMp.WrongBlockVote.Pals[1].Val.Start--
	err = env.ApplyMPs(nextEpoch, malformedVotesMp)
	require.ErrorIs(err, basiccheck.ErrWrongMP)

	for i := range correctMp.WrongBlockVote.Pals {
		tooLateMp := copyMP(correctMp)
		tooLateMp.WrongBlockVote.Pals[i].Val.Epoch = 1
		err = env.ApplyMPs(nextEpoch, tooLateMp)
		require.ErrorIs(err, basiccheck.ErrMPTooLate)
	}

	for i := range correctMp.WrongBlockVote.Pals {
		futureEpochMp := copyMP(correctMp)
		futureEpochMp.WrongBlockVote.Pals[i].Signed.Locator.Epoch = futureEpochMp.WrongBlockVote.Pals[i].Val.Epoch - 1
		err = env.ApplyMPs(nextEpoch, futureEpochMp)
		require.ErrorIs(err, basiccheck.FutureBVsEpoch)
	}

	sameCreatorMp := copyMP(correctMp)
	sameCreatorMp.WrongBlockVote.Pals[0].Signed.Locator.Creator = sameCreatorMp.WrongBlockVote.Pals[1].Signed.Locator.Creator
	err = env.ApplyMPs(nextEpoch, sameCreatorMp)
	require.ErrorIs(err, basiccheck.ErrWrongCreatorMP)

	for i := range correctMp.WrongBlockVote.Pals {
		wrongSigMp := copyMP(correctMp)
		wrongSigMp.WrongBlockVote.Pals[i].Signed.Sig[0]++
		err = env.ApplyMPs(nextEpoch, wrongSigMp)
		require.ErrorIs(err, heavycheck.ErrWrongEventSig)
	}

	for i := range correctMp.WrongBlockVote.Pals {
		wrongSigMp := copyMP(correctMp)
		wrongSigMp.WrongBlockVote.Pals[i].Signed.Locator.BaseHash = hash.HexToHash("0x10")
		err = env.ApplyMPs(nextEpoch, wrongSigMp)
		require.ErrorIs(err, heavycheck.ErrWrongEventSig)
	}

	wrongAuthEpochMp := copyMP(correctMp)
	wrongAuthEpochMp.WrongBlockVote.Pals[0].Signed.Locator.Epoch = startEpoch - 1
	wrongAuthEpochMp.WrongBlockVote.Pals[0].Val.Epoch = startEpoch - 1
	wrongAuthEpochMp.WrongBlockVote.Pals[1].Signed.Locator.Epoch = startEpoch - 1
	wrongAuthEpochMp.WrongBlockVote.Pals[1].Val.Epoch = startEpoch - 1
	err = env.ApplyMPs(nextEpoch, wrongAuthEpochMp)
	require.ErrorIs(err, heavycheck.ErrUnknownEpochBVs)

	goodVotesMp := copyMP(correctMp)
	goodVotesMp.WrongBlockVote.Pals[0].Val.Votes[1] = env.store.GetFullBlockRecord(goodVotesMp.WrongBlockVote.Block).Hash()
	goodVotesMp.WrongBlockVote.Pals[1].Val.Votes[0] = env.store.GetFullBlockRecord(goodVotesMp.WrongBlockVote.Block).Hash()
	err = env.ApplyMPs(nextEpoch, goodVotesMp)
	require.ErrorIs(err, heavycheck.ErrWrongPayloadHash)
	sign(&goodVotesMp)
	err = env.ApplyMPs(nextEpoch, goodVotesMp)
	require.NoError(err)
	require.Equal(idx.Validator(3), env.store.GetValidators().Len())

	err = env.ApplyMPs(nextEpoch, correctMp)
	require.NoError(err)
	require.Equal(idx.Validator(1), env.store.GetValidators().Len())
	require.False(env.store.GetValidators().Exists(1))
	require.False(env.store.GetValidators().Exists(2))
}

func TestMisbehaviourProofsWrongBlockEpoch(t *testing.T) {
	logger.SetTestMode(t)
	require := require.New(t)

	const validatorsNum = 3

	startEpoch := idx.Epoch(basiccheck.MaxLiableEpochs)
	env := newTestEnv(startEpoch, validatorsNum)
	defer env.Close()

	// move epoch further
	_, err := env.ApplyTxs(nextEpoch, env.Transfer(1, 1, common.Big0))
	require.NoError(err)
	_, err = env.ApplyTxs(nextEpoch, env.Transfer(1, 1, common.Big0))
	require.NoError(err)
	require.Greater(env.store.GetEpoch(), startEpoch)

	correctMp := inter.MisbehaviourProof{
		WrongBlockVote: &inter.WrongBlockVote{
			Block: env.store.GetLatestBlockIndex(),
			Pals: [2]inter.LlrSignedBlockVotes{
				{
					Val: inter.LlrBlockVotes{
						Start: env.store.GetLatestBlockIndex() - 1,
						Epoch: startEpoch,
						Votes: []hash.Hash{
							hash.Zero,
							hash.HexToHash("0x01"),
						},
					},
				},
				{
					Val: inter.LlrBlockVotes{
						Start: env.store.GetLatestBlockIndex(),
						Epoch: startEpoch,
						Votes: []hash.Hash{
							hash.HexToHash("0x02"),
						},
					},
				},
			},
			WrongEpoch: true,
		},
	}
	require.Greater(env.store.GetEpoch(), correctMp.WrongBlockVote.Pals[0].Val.Epoch)

	// sign
	sign := func(mp *inter.MisbehaviourProof) {
		for i, p := range mp.WrongBlockVote.Pals {
			e := &inter.MutableEventPayload{}
			e.SetVersion(1)
			e.SetBlockVotes(p.Val)
			e.SetEpoch(env.store.GetEpoch() - idx.Epoch(i))
			e.SetCreator(idx.ValidatorID(i + 1))
			e.SetPayloadHash(inter.CalcPayloadHash(e))

			sig, err := env.signer.Sign(env.pubkeys[i], e.HashToSign().Bytes())
			require.NoError(err)
			sSig := inter.Signature{}
			copy(sSig[:], sig)
			e.SetSig(sSig)

			mp.WrongBlockVote.Pals[i] = inter.AsSignedBlockVotes(e)
		}
	}
	sign(&correctMp)

	// it allows same votes, as it failed verification only on last step
	sameVotesMp := copyMP(correctMp)
	sameVotesMp.WrongBlockVote.Pals[0].Val.Votes[1] = sameVotesMp.WrongBlockVote.Pals[1].Val.Votes[0]
	err = env.ApplyMPs(nextEpoch, sameVotesMp)
	require.ErrorIs(err, heavycheck.ErrWrongPayloadHash)

	differentEpochsMp := copyMP(correctMp)
	differentEpochsMp.WrongBlockVote.Pals[0].Val.Epoch++
	err = env.ApplyMPs(nextEpoch, differentEpochsMp)
	require.ErrorIs(err, basiccheck.ErrNoCrimeInMP)

	for i := range correctMp.WrongBlockVote.Pals {
		malformedVotesMp := copyMP(correctMp)
		malformedVotesMp.WrongBlockVote.Pals[i].Val.Votes = []hash.Hash{}
		err = env.ApplyMPs(nextEpoch, malformedVotesMp)
		require.ErrorIs(err, basiccheck.MalformedBVs)
	}
	for i := range correctMp.WrongBlockVote.Pals {
		malformedVotesMp := copyMP(correctMp)
		malformedVotesMp.WrongBlockVote.Pals[i].Val.Start -= 2
		err = env.ApplyMPs(nextEpoch, malformedVotesMp)
		require.ErrorIs(err, basiccheck.ErrWrongMP)
	}
	malformedVotesMp := copyMP(correctMp)
	malformedVotesMp.WrongBlockVote.Pals[1].Val.Start--
	err = env.ApplyMPs(nextEpoch, malformedVotesMp)
	require.ErrorIs(err, basiccheck.ErrWrongMP)

	for i := range correctMp.WrongBlockVote.Pals {
		tooLateMp := copyMP(correctMp)
		tooLateMp.WrongBlockVote.Pals[i].Val.Epoch = 1
		err = env.ApplyMPs(nextEpoch, tooLateMp)
		require.ErrorIs(err, basiccheck.ErrMPTooLate)
	}

	for i := range correctMp.WrongBlockVote.Pals {
		futureEpochMp := copyMP(correctMp)
		futureEpochMp.WrongBlockVote.Pals[i].Signed.Locator.Epoch = futureEpochMp.WrongBlockVote.Pals[i].Val.Epoch - 1
		err = env.ApplyMPs(nextEpoch, futureEpochMp)
		require.ErrorIs(err, basiccheck.FutureBVsEpoch)
	}

	sameCreatorMp := copyMP(correctMp)
	sameCreatorMp.WrongBlockVote.Pals[0].Signed.Locator.Creator = sameCreatorMp.WrongBlockVote.Pals[1].Signed.Locator.Creator
	err = env.ApplyMPs(nextEpoch, sameCreatorMp)
	require.ErrorIs(err, basiccheck.ErrWrongCreatorMP)

	for i := range correctMp.WrongBlockVote.Pals {
		wrongSigMp := copyMP(correctMp)
		wrongSigMp.WrongBlockVote.Pals[i].Signed.Sig[0]++
		err = env.ApplyMPs(nextEpoch, wrongSigMp)
		require.ErrorIs(err, heavycheck.ErrWrongEventSig)
	}

	for i := range correctMp.WrongBlockVote.Pals {
		wrongSigMp := copyMP(correctMp)
		wrongSigMp.WrongBlockVote.Pals[i].Signed.Locator.BaseHash = hash.HexToHash("0x10")
		err = env.ApplyMPs(nextEpoch, wrongSigMp)
		require.ErrorIs(err, heavycheck.ErrWrongEventSig)
	}

	wrongAuthEpochMp := copyMP(correctMp)
	wrongAuthEpochMp.WrongBlockVote.Pals[0].Signed.Locator.Epoch = startEpoch - 1
	wrongAuthEpochMp.WrongBlockVote.Pals[0].Val.Epoch = startEpoch - 1
	wrongAuthEpochMp.WrongBlockVote.Pals[1].Signed.Locator.Epoch = startEpoch - 1
	wrongAuthEpochMp.WrongBlockVote.Pals[1].Val.Epoch = startEpoch - 1
	err = env.ApplyMPs(nextEpoch, wrongAuthEpochMp)
	require.ErrorIs(err, heavycheck.ErrUnknownEpochBVs)

	goodEpochMp := copyMP(correctMp)
	goodEpochMp.WrongBlockVote.Pals[0].Val.Epoch = env.store.FindBlockEpoch(goodEpochMp.WrongBlockVote.Block)
	goodEpochMp.WrongBlockVote.Pals[1].Val.Epoch = env.store.FindBlockEpoch(goodEpochMp.WrongBlockVote.Block)
	err = env.ApplyMPs(nextEpoch, goodEpochMp)
	require.ErrorIs(err, heavycheck.ErrWrongPayloadHash)
	sign(&goodEpochMp)
	err = env.ApplyMPs(nextEpoch, goodEpochMp)
	require.NoError(err)
	require.Equal(idx.Validator(3), env.store.GetValidators().Len())

	err = env.ApplyMPs(nextEpoch, correctMp)
	require.NoError(err)
	require.Equal(idx.Validator(1), env.store.GetValidators().Len())
	require.False(env.store.GetValidators().Exists(1))
	require.False(env.store.GetValidators().Exists(2))
}

func TestMisbehaviourProofsEpochVoteDoublesign(t *testing.T) {
	logger.SetTestMode(t)
	require := require.New(t)

	const validatorsNum = 3

	startEpoch := idx.Epoch(basiccheck.MaxLiableEpochs)
	env := newTestEnv(startEpoch, validatorsNum)
	defer env.Close()

	// move epoch further
	_, err := env.ApplyTxs(nextEpoch, env.Transfer(1, 1, common.Big0))
	require.NoError(err)
	_, err = env.ApplyTxs(nextEpoch, env.Transfer(1, 1, common.Big0))
	require.NoError(err)
	require.Greater(env.store.GetEpoch(), startEpoch)

	correctMp := inter.MisbehaviourProof{
		EpochVoteDoublesign: &inter.EpochVoteDoublesign{
			Pair: [2]inter.LlrSignedEpochVote{
				{
					Val: inter.LlrEpochVote{
						Epoch: startEpoch + 1,
						Vote:  hash.HexToHash("0x01"),
					},
				},
				{
					Val: inter.LlrEpochVote{
						Epoch: startEpoch + 1,
						Vote:  hash.HexToHash("0x02"),
					},
				},
			},
		},
	}

	// sign
	for i, p := range correctMp.EpochVoteDoublesign.Pair {
		e := &inter.MutableEventPayload{}
		e.SetVersion(1)
		e.SetEpochVote(p.Val)
		e.SetEpoch(env.store.GetEpoch() - idx.Epoch(i))
		e.SetCreator(3)
		e.SetPayloadHash(inter.CalcPayloadHash(e))

		sig, err := env.signer.Sign(env.pubkeys[2], e.HashToSign().Bytes())
		require.NoError(err)
		sSig := inter.Signature{}
		copy(sSig[:], sig)
		e.SetSig(sSig)

		correctMp.EpochVoteDoublesign.Pair[i] = inter.AsSignedEpochVote(e)
	}

	sameVotesMp := copyMP(correctMp)
	sameVotesMp.EpochVoteDoublesign.Pair[0].Val.Vote = sameVotesMp.EpochVoteDoublesign.Pair[1].Val.Vote
	err = env.ApplyMPs(nextEpoch, sameVotesMp)
	require.ErrorIs(err, basiccheck.ErrNoCrimeInMP)

	differentEpochsMp := copyMP(correctMp)
	differentEpochsMp.EpochVoteDoublesign.Pair[0].Val.Epoch++
	err = env.ApplyMPs(nextEpoch, differentEpochsMp)
	require.ErrorIs(err, basiccheck.ErrNoCrimeInMP)

	malformedVotesMp := copyMP(correctMp)
	malformedVotesMp.EpochVoteDoublesign.Pair[1].Val.Epoch = 0
	err = env.ApplyMPs(nextEpoch, malformedVotesMp)
	require.ErrorIs(err, basiccheck.MalformedEV)

	tooLateMp := copyMP(correctMp)
	tooLateMp.EpochVoteDoublesign.Pair[0].Val.Epoch = 1
	tooLateMp.EpochVoteDoublesign.Pair[1].Val.Epoch = 1
	err = env.ApplyMPs(nextEpoch, tooLateMp)
	require.ErrorIs(err, basiccheck.ErrMPTooLate)

	for i := range correctMp.EpochVoteDoublesign.Pair {
		futureEpochMp := copyMP(correctMp)
		futureEpochMp.EpochVoteDoublesign.Pair[i].Signed.Locator.Epoch = futureEpochMp.EpochVoteDoublesign.Pair[i].Val.Epoch - 1
		err = env.ApplyMPs(nextEpoch, futureEpochMp)
		require.ErrorIs(err, basiccheck.FutureEVEpoch)
	}

	wrongCreatorMp := copyMP(correctMp)
	wrongCreatorMp.EpochVoteDoublesign.Pair[0].Signed.Locator.Creator++
	err = env.ApplyMPs(nextEpoch, wrongCreatorMp)
	require.ErrorIs(err, basiccheck.ErrWrongCreatorMP)

	sameEventsMp := copyMP(correctMp)
	sameEventsMp.EpochVoteDoublesign.Pair[0] = sameEventsMp.EpochVoteDoublesign.Pair[1]
	err = env.ApplyMPs(nextEpoch, sameEventsMp)
	require.ErrorIs(err, basiccheck.ErrNoCrimeInMP)

	for i := range correctMp.EpochVoteDoublesign.Pair {
		wrongSigMp := copyMP(correctMp)
		wrongSigMp.EpochVoteDoublesign.Pair[i].Signed.Sig[0]++
		err = env.ApplyMPs(nextEpoch, wrongSigMp)
		require.ErrorIs(err, heavycheck.ErrWrongEventSig)
	}

	for i := range correctMp.EpochVoteDoublesign.Pair {
		wrongSigMp := copyMP(correctMp)
		wrongSigMp.EpochVoteDoublesign.Pair[i].Signed.Locator.BaseHash = hash.HexToHash("0x10")
		err = env.ApplyMPs(nextEpoch, wrongSigMp)
		require.ErrorIs(err, heavycheck.ErrWrongEventSig)
	}

	wrongAuthEpochMp := copyMP(correctMp)
	wrongAuthEpochMp.EpochVoteDoublesign.Pair[0].Signed.Locator.Epoch = startEpoch - 1
	wrongAuthEpochMp.EpochVoteDoublesign.Pair[0].Val.Epoch = startEpoch - 1
	wrongAuthEpochMp.EpochVoteDoublesign.Pair[1].Signed.Locator.Epoch = startEpoch - 1
	wrongAuthEpochMp.EpochVoteDoublesign.Pair[1].Val.Epoch = startEpoch - 1
	err = env.ApplyMPs(nextEpoch, wrongAuthEpochMp)
	require.ErrorIs(err, heavycheck.ErrUnknownEpochEV)

	err = env.ApplyMPs(nextEpoch, correctMp)
	require.NoError(err)
	require.Equal(idx.Validator(2), env.store.GetValidators().Len())
	require.False(env.store.GetValidators().Exists(3))
}

func TestMisbehaviourProofsWrongVote(t *testing.T) {
	logger.SetTestMode(t)
	require := require.New(t)

	const validatorsNum = 3

	startEpoch := idx.Epoch(basiccheck.MaxLiableEpochs)
	env := newTestEnv(startEpoch, validatorsNum)
	defer env.Close()

	// move epoch further
	_, err := env.ApplyTxs(nextEpoch, env.Transfer(1, 1, common.Big0))
	require.NoError(err)
	_, err = env.ApplyTxs(nextEpoch, env.Transfer(1, 1, common.Big0))
	require.NoError(err)
	require.Greater(env.store.GetEpoch(), startEpoch)

	correctMp := inter.MisbehaviourProof{
		WrongEpochVote: &inter.WrongEpochVote{
			Pals: [2]inter.LlrSignedEpochVote{
				{
					Val: inter.LlrEpochVote{
						Epoch: startEpoch + 1,
						Vote:  hash.HexToHash("0x01"),
					},
				},
				{
					Val: inter.LlrEpochVote{
						Epoch: startEpoch + 1,
						Vote:  hash.HexToHash("0x01"),
					},
				},
			},
		},
	}

	// sign
	sign := func(mp *inter.MisbehaviourProof) {
		for i, p := range mp.WrongEpochVote.Pals {
			e := &inter.MutableEventPayload{}
			e.SetVersion(1)
			e.SetEpochVote(p.Val)
			e.SetEpoch(env.store.GetEpoch() - idx.Epoch(i))
			e.SetCreator(idx.ValidatorID(i + 1))
			e.SetPayloadHash(inter.CalcPayloadHash(e))

			sig, err := env.signer.Sign(env.pubkeys[i], e.HashToSign().Bytes())
			require.NoError(err)
			sSig := inter.Signature{}
			copy(sSig[:], sig)
			e.SetSig(sSig)

			mp.WrongEpochVote.Pals[i] = inter.AsSignedEpochVote(e)
		}
	}
	sign(&correctMp)

	differentVotesMp := copyMP(correctMp)
	differentVotesMp.WrongEpochVote.Pals[0].Val.Vote = hash.HexToHash("0x02")
	err = env.ApplyMPs(nextEpoch, differentVotesMp)
	require.ErrorIs(err, basiccheck.ErrNoCrimeInMP)

	differentEpochsMp := copyMP(correctMp)
	differentEpochsMp.WrongEpochVote.Pals[0].Val.Epoch++
	err = env.ApplyMPs(nextEpoch, differentEpochsMp)
	require.ErrorIs(err, basiccheck.ErrNoCrimeInMP)

	malformedVotesMp := copyMP(correctMp)
	malformedVotesMp.WrongEpochVote.Pals[1].Val.Epoch = 0
	err = env.ApplyMPs(nextEpoch, malformedVotesMp)
	require.ErrorIs(err, basiccheck.MalformedEV)

	tooLateMp := copyMP(correctMp)
	tooLateMp.WrongEpochVote.Pals[0].Val.Epoch = 1
	tooLateMp.WrongEpochVote.Pals[1].Val.Epoch = 1
	err = env.ApplyMPs(nextEpoch, tooLateMp)
	require.ErrorIs(err, basiccheck.ErrMPTooLate)

	for i := range correctMp.WrongEpochVote.Pals {
		futureEpochMp := copyMP(correctMp)
		futureEpochMp.WrongEpochVote.Pals[i].Signed.Locator.Epoch = futureEpochMp.WrongEpochVote.Pals[i].Val.Epoch - 1
		err = env.ApplyMPs(nextEpoch, futureEpochMp)
		require.ErrorIs(err, basiccheck.FutureEVEpoch)
	}

	sameCreatorMp := copyMP(correctMp)
	sameCreatorMp.WrongEpochVote.Pals[0].Signed.Locator.Creator = sameCreatorMp.WrongEpochVote.Pals[1].Signed.Locator.Creator
	err = env.ApplyMPs(nextEpoch, sameCreatorMp)
	require.ErrorIs(err, basiccheck.ErrWrongCreatorMP)

	for i := range correctMp.WrongEpochVote.Pals {
		wrongSigMp := copyMP(correctMp)
		wrongSigMp.WrongEpochVote.Pals[i].Signed.Sig[0]++
		err = env.ApplyMPs(nextEpoch, wrongSigMp)
		require.ErrorIs(err, heavycheck.ErrWrongEventSig)
	}

	for i := range correctMp.WrongEpochVote.Pals {
		wrongSigMp := copyMP(correctMp)
		wrongSigMp.WrongEpochVote.Pals[i].Signed.Locator.BaseHash = hash.HexToHash("0x10")
		err = env.ApplyMPs(nextEpoch, wrongSigMp)
		require.ErrorIs(err, heavycheck.ErrWrongEventSig)
	}

	wrongAuthEpochMp := copyMP(correctMp)
	wrongAuthEpochMp.WrongEpochVote.Pals[0].Signed.Locator.Epoch = startEpoch - 1
	wrongAuthEpochMp.WrongEpochVote.Pals[0].Val.Epoch = startEpoch - 1
	wrongAuthEpochMp.WrongEpochVote.Pals[1].Signed.Locator.Epoch = startEpoch - 1
	wrongAuthEpochMp.WrongEpochVote.Pals[1].Val.Epoch = startEpoch - 1
	err = env.ApplyMPs(nextEpoch, wrongAuthEpochMp)
	require.ErrorIs(err, heavycheck.ErrUnknownEpochEV)

	goodVotesMp := copyMP(correctMp)
	goodVotesMp.WrongEpochVote.Pals[0].Val.Vote = env.store.GetFullEpochRecord(goodVotesMp.WrongEpochVote.Pals[0].Val.Epoch).Hash()
	goodVotesMp.WrongEpochVote.Pals[1].Val.Vote = env.store.GetFullEpochRecord(goodVotesMp.WrongEpochVote.Pals[1].Val.Epoch).Hash()
	err = env.ApplyMPs(nextEpoch, goodVotesMp)
	require.ErrorIs(err, heavycheck.ErrWrongPayloadHash)
	sign(&goodVotesMp)
	err = env.ApplyMPs(nextEpoch, goodVotesMp)
	require.NoError(err)
	require.Equal(idx.Validator(3), env.store.GetValidators().Len())

	err = env.ApplyMPs(nextEpoch, correctMp)
	require.NoError(err)
	require.Equal(idx.Validator(1), env.store.GetValidators().Len())
	require.False(env.store.GetValidators().Exists(1))
	require.False(env.store.GetValidators().Exists(2))
}
