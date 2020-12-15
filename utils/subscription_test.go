package utils

import (
	"testing"
	"time"

	notify "github.com/ethereum/go-ethereum/event"
	"github.com/stretchr/testify/require"
)

func TestChanBuffer(t *testing.T) {
	require := require.New(t)
	read := func(from chan int, exp int) {
		readFrom(t, from, exp)
	}

	out := make(chan int)
	cb := NewChanBuffer(out)
	in := cb.InChannel().(chan int)

	N1 := 3
	for i := 0; i < N1; i++ {
		in <- i
	}
	read(out, 0)

	cb.Flush()
	read(out, N1)

	close(in)
	require.Eventually(func() bool {
		_, ok := <-out
		return !ok
	}, 10*time.Second, 10*time.Millisecond)
}

func TestFlushableSubscriptionScope(t *testing.T) {
	read := func(from chan int, exp int) {
		readFrom(t, from, exp)
	}

	var (
		scope FlushableSubscriptionScope
		feed1 notify.Feed
		feed2 notify.Feed
	)

	chan1 := make(chan int)
	cbuf1 := NewChanBuffer(chan1)
	sub1 := scope.Track(feed1.Subscribe(cbuf1.InChannel()), cbuf1.Flush)

	chan2 := make(chan int)
	cbuf2 := NewChanBuffer(chan2)
	_ = scope.Track(feed2.Subscribe(cbuf2.InChannel()), cbuf2.Flush)

	feed1.Send(int(1))
	feed1.Send(int(2))
	read(chan1, 0)
	read(chan2, 0)

	scope.Flush()
	read(chan1, 2)
	read(chan2, 0)

	feed1.Send(int(3))
	feed2.Send(int(4))
	read(chan1, 0)
	read(chan2, 0)

	scope.Flush()
	read(chan1, 1)
	read(chan2, 1)

	sub1.Unsubscribe()
	scope.Close()
}

func readFrom(t *testing.T, ch chan int, exp int) {
	count := 0
	for {
		select {
		case <-ch:
			count++
		case <-time.After(time.Second / 2):
			require.Equal(t, exp, count)
			return
		}
	}
}
