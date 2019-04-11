package network

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
)

var (
	media     = make(map[Addr]*Listener)
	syncMedia sync.RWMutex
)

func listenFreeAddr(desired Addr) (res *Listener, err error) {
	for desired == "" {
		free := Addr(fmt.Sprintf(":%d", rand.Intn(1<<16-1)+1))
		if res, err = listenFreeAddr(free); err == nil {
			return
		}
	}

	syncMedia.Lock()
	defer syncMedia.Unlock()

	if _, exists := media[desired]; exists {
		err = &net.AddrError{
			Err:  "net address already in use",
			Addr: desired.String(),
		}
		return
	}
	res = NewListener(desired)
	media[desired] = res
	return
}

func findListener(addr Addr) *Listener {
	syncMedia.RLock()
	defer syncMedia.RUnlock()

	return media[addr]
}

func removeListener(l *Listener) {
	syncMedia.Lock()
	defer syncMedia.Unlock()

	delete(media, l.NetAddr)
}
