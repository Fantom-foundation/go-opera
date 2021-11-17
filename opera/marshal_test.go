package opera

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/require"
)

func TestUpdateRules(t *testing.T) {
	require := require.New(t)

	var exp Rules
	exp.Epochs.MaxEpochGas = 99

	exp.Dag.MaxParents = 5
	exp.Economy.MinGasPrice = big.NewInt(7)
	exp.Blocks.MaxBlockGas = 1000
	got, err := UpdateRules(exp, []byte(`{"Dag":{"MaxParents":5},"Economy":{"MinGasPrice":7},"Blocks":{"MaxBlockGas":1000}}`))
	require.NoError(err)
	require.Equal(exp.String(), got.String(), "mutate fields")

	exp.Dag.MaxParents = 0
	got, err = UpdateRules(exp, []byte(`{"Name":"xxx","NetworkID":1,"Dag":{"MaxParents":0}}`))
	require.NoError(err)
	require.Equal(exp.String(), got.String(), "readonly fields")

	got, err = UpdateRules(exp, []byte(`{}`))
	require.NoError(err)
	require.Equal(exp.String(), got.String(), "empty diff")

	_, err = UpdateRules(exp, []byte(`}{`))
	require.Error(err)
}

func TestMainNetRulesRLP(t *testing.T) {
	rules := MainNetRules()
	require := require.New(t)

	b, err := rlp.EncodeToBytes(rules)
	require.NoError(err)

	decodedRules := Rules{}
	require.NoError(rlp.DecodeBytes(b, &decodedRules))

	require.Equal(rules.String(), decodedRules.String())
}

func TestRulesBerlinRLP(t *testing.T) {
	rules := MainNetRules()
	rules.Upgrades.Berlin = true
	require := require.New(t)

	b, err := rlp.EncodeToBytes(rules)
	require.NoError(err)

	decodedRules := Rules{}
	require.NoError(rlp.DecodeBytes(b, &decodedRules))

	require.Equal(rules.String(), decodedRules.String())
	require.True(decodedRules.Upgrades.Berlin)
}

func TestRulesLondonRLP(t *testing.T) {
	rules := MainNetRules()
	rules.Upgrades.London = true
	rules.Upgrades.Berlin = true
	require := require.New(t)

	b, err := rlp.EncodeToBytes(rules)
	require.NoError(err)

	decodedRules := Rules{}
	require.NoError(rlp.DecodeBytes(b, &decodedRules))

	require.Equal(rules.String(), decodedRules.String())
	require.True(decodedRules.Upgrades.Berlin)
	require.True(decodedRules.Upgrades.London)
}

func TestRulesBerlinCompatibilityRLP(t *testing.T) {
	require := require.New(t)

	b1, err := rlp.EncodeToBytes(Upgrades{
		Berlin: true,
	})
	require.NoError(err)

	b2, err := rlp.EncodeToBytes(struct {
		Berlin bool
	}{true})
	require.NoError(err)

	require.Equal(b2, b1)
}

func TestGasRulesLLRCompatibilityRLP(t *testing.T) {
	require := require.New(t)

	b1, err := rlp.EncodeToBytes(GasRules{
		MaxEventGas:          1,
		EventGas:             2,
		ParentGas:            3,
		ExtraDataGas:         4,
		BlockVotesBaseGas:    0,
		BlockVoteGas:         0,
		EpochVoteGas:         0,
		MisbehaviourProofGas: 0,
	})
	require.NoError(err)

	b2, err := rlp.EncodeToBytes(struct {
		MaxEventGas  uint64
		EventGas     uint64
		ParentGas    uint64
		ExtraDataGas uint64
	}{1, 2, 3, 4})
	require.NoError(err)

	require.Equal(b2, b1)
}
