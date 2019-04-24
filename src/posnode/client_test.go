package posnode

import (
	"context"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/posnode/api"
)

func TestClient(t *testing.T) {
	server := NewForTests("server.fake", nil, nil)
	server.StartService()
	defer server.StopService()
	peer := server.AsPeer()

	node := NewForTests("client.fake", nil, nil)
	node.initClient()

	ping := func(t *testing.T) (proto.Message, error) {
		client, free, fail, err := node.ConnectTo(peer)
		if !assert.NoError(t, err) {
			return nil, err
		}
		defer free()

		ctx, cancel := context.WithTimeout(context.Background(), node.conf.ClientTimeout)
		defer cancel()

		resp, err := client.GetPeerInfo(ctx, &api.PeerRequest{})
		if err != nil {
			fail(err)
		}

		return resp, err
	}

	t.Run("1st-connection", func(t *testing.T) {
		assert := assert.New(t)

		resp, err := ping(t)
		_ = assert.NoError(err) && assert.NotNil(resp)
	})

	t.Run("2nd-connection", func(t *testing.T) {
		assert := assert.New(t)

		resp, err := ping(t)
		_ = assert.NoError(err) && assert.NotNil(resp)
	})

	t.Run("Re-connection 1", func(t *testing.T) {
		assert := assert.New(t)

		server.StopService()
		server.StartService()
		resp, err := ping(t)
		_ = assert.Error(err) && assert.Nil(resp)
	})

	t.Run("Re-connection 2", func(t *testing.T) {
		assert := assert.New(t)

		resp, err := ping(t)
		_ = assert.NoError(err) && assert.NotNil(resp)
	})

	// TODO: test the all situations.
}
