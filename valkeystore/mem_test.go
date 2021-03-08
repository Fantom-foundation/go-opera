package valkeystore

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMemKeystoreAdd(t *testing.T) {
	require := require.New(t)
	keystore := NewMemKeystore()

	key, err := keystore.Get(pubkey1, "auth1")
	require.EqualError(err, ErrNotFound.Error())
	require.Nil(key)

	err = keystore.Add(pubkey1, key1, "auth1")
	require.NoError(err)

	testGet(t, keystore, pubkey1, key1, "auth1")

	err = keystore.Add(pubkey2, key2, "auth2")
	require.NoError(err)

	testGet(t, keystore, pubkey1, key1, "auth1")
	testGet(t, keystore, pubkey2, key2, "auth2")

	err = keystore.Add(pubkey2, key2, "auth1")
	require.Error(err, ErrAlreadyExists.Error())

	testGet(t, keystore, pubkey2, key2, "auth2")
}
