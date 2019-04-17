package posnode

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func Test_connPool_get(t *testing.T) {
	cli := newClient(time.Second, make([]grpc.DialOption, 0))
	pool := cli.pool

	t.Run("new conn", func(t *testing.T) {
		assert := assert.New(t)

		conn, err := pool.get("example.server")
		assert.Nil(err)
		assert.NotNil(conn)
	})

	t.Run("conn exists", func(t *testing.T) {
		assert := assert.New(t)

		existingConn := pool.connections["example.server"]
		conn, err := pool.get("example.server")
		assert.Nil(err)
		assert.Equal(conn, existingConn)
	})
}
