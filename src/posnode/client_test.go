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
		assert.Equal(conn, wrap)
	})

}

func Test_client_free(t *testing.T) {
	node1 := NewForTests("example.server", nil, nil)
	node1.StartService()
	defer node1.StopService()

	node := NewForTests("client.test", nil, nil)
	node.initClient()

	cli := &node.client
	cli.connectTimeout = time.Millisecond * 500

	t.Run("unused connection", func(t *testing.T) {
		assert := assert.New(t)
		conn, err := cli.get("example.server:55555")
		assert.Nil(err)

		wrap := cli.connections["example.server:55555"]
		cli.maxUnusedDuration = time.Second
		wrap.usedAt = time.Now().Add(-2 * time.Second)

		cli.free(conn)
		assert.Equal(0, len(cli.connections))
	})
}

func Test_client_watchConnections(t *testing.T) {
	node1 := NewForTests("example.server", nil, nil)
	node1.StartService()
	defer node1.StopService()

	node := NewForTests("client.test", nil, nil)
	tickChan := make(chan time.Time)
	ticker := time.Ticker{
		C: tickChan,
	}
	node.initClient()

	cli := &node.client
	cli.connWatchTicker = &ticker
	cli.connectTimeout = time.Millisecond * 500

	t.Run("clear unused on tick", func(t *testing.T) {
		assert := assert.New(t)
		conn, err := cli.get("example.server:55555")
		assert.Nil(err)

		cli.maxUnusedDuration = time.Second
		conn.usedAt = time.Now().Add(-2 * time.Second)

		watchFreeCalledChan := make(chan struct{})
		watchFreeCalled = func() {
			watchFreeCalledChan <- struct{}{}
		}

		tickChan <- time.Now()

		select {
		case <-watchFreeCalledChan:
		case <-time.After(time.Second):
			t.Error("expected to free connection")
		}

		assert.Equal(0, len(cli.connections))
	})
}
