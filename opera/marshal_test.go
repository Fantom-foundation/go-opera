package opera

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpdateRules(t *testing.T) {
	require := require.New(t)

	var exp Rules
	exp.Epochs.MaxEpochGas = 99

	exp.Dag.MaxParents = 5
	exp.Economy.BlockMissedSlack = 7
	exp.Blocks.MaxBlockGas = 1000
	got, err := UpdateRules(exp, []byte(`{"Dag":{"MaxParents":5},"Economy":{"BlockMissedSlack":7},"Blocks":{"MaxBlockGas":1000}}`))
	require.NoError(err)
	require.Equal(exp, got, "mutate fields")

	exp.Dag.MaxParents = 0
	got, err = UpdateRules(exp, []byte(`{"Name":"xxx","NetworkID":1,"Dag":{"MaxParents":0}}`))
	require.NoError(err)
	require.Equal(exp, got, "readonly fields")

	got, err = UpdateRules(exp, []byte(`{}`))
	require.NoError(err)
	require.Equal(exp, got, "empty diff")

	_, err = UpdateRules(exp, []byte(`}{`))
	require.Error(err)
}
