package proxy

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/andrecronje/lachesis/src/common"
	"github.com/andrecronje/lachesis/src/poset"
	"github.com/andrecronje/lachesis/src/proxy/proto"
)

var (
	timeout    = 2 * time.Second
	errTimeout = "time is over"
)

func TestGrpcCalls(t *testing.T) {
	addr := "127.0.0.1:9993"
	logger := common.NewTestLogger(t)

	s, err := NewGrpcAppProxy(addr, timeout, logger)
	assert.NoError(t, err)

	c, err := NewGrpcLachesisProxy(addr, logger)
	assert.NoError(t, err)

	t.Run("#1 Send tx", func(t *testing.T) {
		assert := assert.New(t)
		gold := []byte("123456")

		err = c.SubmitTx(gold)
		assert.NoError(err)

		select {
		case tx := <-s.SubmitCh():
			assert.Equal(gold, tx)
		case <-time.After(timeout):
			assert.Fail(errTimeout)
		}
	})

	t.Run("#2 Receive block", func(t *testing.T) {
		assert := assert.New(t)
		block := poset.Block{}
		gold := []byte("123456")

		go func() {
			select {
			case event := <-c.CommitCh():
				assert.Equal(block, event.Block)
				event.RespChan <- proto.CommitResponse{
					StateHash: gold,
					Error:     nil,
				}
			case <-time.After(timeout):
				assert.Fail(errTimeout)
			}
		}()

		answ, err := s.CommitBlock(block)
		assert.Nil(err)
		assert.Equal(gold, answ)
	})

	t.Run("#3 Receive snapshot query", func(t *testing.T) {
		assert := assert.New(t)
		index := 1
		gold := []byte("123456")

		go func() {
			select {
			case event := <-c.SnapshotRequestCh():
				assert.Equal(index, event.BlockIndex)
				event.RespChan <- proto.SnapshotResponse{
					Snapshot: gold,
					Error:    nil,
				}
			case <-time.After(timeout):
				assert.Fail(errTimeout)
			}
		}()

		answ, err := s.GetSnapshot(index)
		assert.Nil(err)
		assert.Equal(gold, answ)
	})

	t.Run("#4 Receive restore command", func(t *testing.T) {
		assert := assert.New(t)
		gold := []byte("123456")

		go func() {
			select {
			case event := <-c.RestoreCh():
				assert.Equal(gold, event.Snapshot)
				event.RespChan <- proto.RestoreResponse{
					StateHash: gold,
					Error:     nil,
				}
			case <-time.After(timeout):
				assert.Fail(errTimeout)
			}
		}()

		err := s.Restore(gold)
		assert.Nil(err)
	})

	err = c.Close()
	assert.NoError(t, err)

	err = s.Close()
	assert.NoError(t, err)

}

func TestGrpcReConnection(t *testing.T) {
	addr := "127.0.0.1:9994"
	logger := common.NewTestLogger(t)

	c, err := NewGrpcLachesisProxy(addr, logger)
	assert.Nil(t, c)
	assert.Error(t, err)

	s, err := NewGrpcAppProxy(addr, timeout, logger)
	assert.NoError(t, err)

	c, err = NewGrpcLachesisProxy(addr, logger)
	assert.NoError(t, err)

	checkConnAndStopServer := func(t *testing.T) {
		assert := assert.New(t)
		gold := []byte("123456")

		err := c.SubmitTx(gold)
		assert.NoError(err)

		select {
		case tx := <-s.SubmitCh():
			assert.Equal(gold, tx)
		case <-time.After(timeout):
			assert.Fail(errTimeout)
		}

		err = s.Close()
		assert.NoError(err)
	}

	t.Run("#1 Send tx after connection", checkConnAndStopServer)

	s, err = NewGrpcAppProxy(addr, timeout/2, logger)
	assert.NoError(t, err)

	t.Run("#2 Send tx after reconnection", checkConnAndStopServer)

	err = c.Close()
	assert.NoError(t, err)
}
