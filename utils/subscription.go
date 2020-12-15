package utils

import (
	"errors"
	"reflect"
	"sync"

	notify "github.com/ethereum/go-ethereum/event"
)

// ChanBuffer buffers the channel until flushed.
type ChanBuffer struct {
	in    interface{}
	buff  []interface{}
	out   reflect.SelectCase
	flush chan struct{}
}

// NewChanBuffer constructor.
func NewChanBuffer(channel interface{}) *ChanBuffer {
	chanval := reflect.ValueOf(channel)
	chantyp := chanval.Type()
	if chantyp.Kind() != reflect.Chan || chantyp.ChanDir()&reflect.SendDir == 0 {
		panic(errors.New("event: Subscribe argument does not have sendable channel type"))
	}

	chantyp = reflect.ChanOf(reflect.BothDir, chantyp.Elem())
	cb := &ChanBuffer{
		out:   reflect.SelectCase{Dir: reflect.SelectSend, Chan: chanval},
		in:    reflect.MakeChan(chantyp, 0).Interface(),
		flush: make(chan struct{}),
	}
	go cb.loop()
	return cb
}

// In returns an input channel.
func (cb *ChanBuffer) InChannel() interface{} {
	return cb.in
}

// Flush the buffer async.
func (cb *ChanBuffer) Flush() {
	cb.flush <- struct{}{}
}

func (cb *ChanBuffer) loop() {
	caseIn := reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(cb.in)}
	caseFlush := reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(cb.flush)}
	for {
		chosen, val, ok := reflect.Select([]reflect.SelectCase{caseIn, caseFlush})
		switch chosen {
		case 0:
			if !ok {
				cb.out.Chan.Close()
				return
			}
			cb.buff = append(cb.buff, val.Interface())
		case 1:
			for _, v := range cb.buff {
				rvalue := reflect.ValueOf(v)
				cb.out.Chan.Send(rvalue)
			}
			cb.buff = cb.buff[:0]
		}
	}
}

// FlushableSubscriptionScope provides a facility to flush and unsubscribe multiple subscriptions at once.
// The zero value is ready to use.
type FlushableSubscriptionScope struct {
	mu     sync.Mutex
	subs   map[*scopeSub]struct{}
	closed bool
}

type scopeSub struct {
	scope *FlushableSubscriptionScope
	sub   notify.Subscription
	flush func()
}

// Track starts tracking a subscription. If the scope is closed, Track returns nil. The
// returned subscription is a wrapper. Unsubscribing the wrapper removes it from the
// scope.
func (sc *FlushableSubscriptionScope) Track(s notify.Subscription, flush func()) notify.Subscription {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.closed {
		return nil
	}
	if sc.subs == nil {
		sc.subs = make(map[*scopeSub]struct{})
	}

	ss := &scopeSub{
		scope: sc,
		sub:   s,
		flush: flush,
	}
	sc.subs[ss] = struct{}{}

	return ss
}

// Flush sends the all delayed notification.
func (sc *FlushableSubscriptionScope) Flush() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.closed {
		return
	}

	for s := range sc.subs {
		s.flush()
	}
}

// Close calls Unsubscribe on all tracked subscriptions and prevents further additions to
// the tracked set. Calls to Track after Close return nil.
func (sc *FlushableSubscriptionScope) Close() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.closed {
		return
	}
	sc.closed = true

	for s := range sc.subs {
		s.sub.Unsubscribe()
	}
	sc.subs = nil
}

// Count returns the number of tracked subscriptions.
// It is meant to be used for debugging.
func (sc *FlushableSubscriptionScope) Count() int {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	return len(sc.subs)
}

func (s *scopeSub) Unsubscribe() {
	s.scope.mu.Lock()
	defer s.scope.mu.Unlock()

	s.sub.Unsubscribe()
	delete(s.scope.subs, s)
}

func (s *scopeSub) Err() <-chan error {
	return s.sub.Err()
}
