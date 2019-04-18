package posnode

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_client_get(t *testing.T) {
	node1 := NewForTests("example.server", nil, nil)
	node1.StartService()
	defer node1.StopService()

	node := NewForTests("client.test", nil, nil)
	node.initClient()

	cli := &node.client
	cli.connectTimeout = time.Millisecond * 500

	t.Run("new conn", func(t *testing.T) {
		assert := assert.New(t)

		conn, err := cli.get("example.server:55555")
		assert.Nil(err)
		assert.NotNil(conn)
	})

	t.Run("conn exists", func(t *testing.T) {
		assert := assert.New(t)

		wrap := cli.connections["example.server:55555"]
		conn, err := cli.get("example.server:55555")
		assert.Nil(err)
		assert.Equal(conn, wrap.ClientConn)
	})

	t.Run("conn max life", func(t *testing.T) {
		assert := assert.New(t)

		wrap := cli.connections["example.server:55555"]
		cli.maxLifeDuration = time.Second
		wrap.initiatedAt = time.Now().Add(-2 * time.Second)

		conn, err := cli.get("example.server:55555")
		assert.Nil(err)
		assert.NotEqual(conn, wrap.ClientConn)
	})
}
