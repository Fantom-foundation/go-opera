package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenericNameOfTmpDB(t *testing.T) {
	require := require.New(t)

	for name, exp := range map[string]string{
		"":              "",
		"main":          "main",
		"main-single":   "main-single",
		"lachesis-0":    "lachesises",
		"lachesis-0999": "lachesises",
		"gossip-50":     "gossipes",
		"epoch-1":       "epoches",
		"xxx-1a":        "xxx-1a",
		"123":           "123",
	} {
		got := genericNameOfTmpDB(name)
		require.Equal(exp, got, name)
	}
}
