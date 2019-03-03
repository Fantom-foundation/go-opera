package fakenet

import (
	"io"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// Network is a fake network.
type Network struct {
	mtx   sync.Mutex
	conns map[string]*Listener
}

// NewNetwork creates a new fake network.
func NewNetwork(listeners ...*Listener) *Network {
	m := make(map[string]*Listener)
	for k := range listeners {
		m[listeners[k].Address] = listeners[k]
	}
	return &Network{conns: m}
}

// CreateListener returns fake listener for a specific address.
func (n *Network) CreateListener(
	network, address string) (net.Listener, error) {
	n.mtx.Lock()
	defer n.mtx.Unlock()

	// If listener exists and not closed, then returns error.
	if lis, ok := n.conns[address]; ok && !lis.isClosed() {
		return nil, ErrAddressAlreadyInUse
	}

	n.conns[address] = NewListener(address)
	return n.conns[address], nil
}

// CreateNetConn returns a fake connection to a fake node.
func (n *Network) CreateNetConn(network,
	address string, timeout time.Duration) (net.Conn, error) {

	// If listener does not exist, returns "connection refused" error.
	n.mtx.Lock()
	if lis, ok := n.conns[address]; !ok || lis.isClosed() {
		n.mtx.Unlock()
		return nil, errors.Errorf(
			"dial tcp %s: connect: connection refused", address)
	}
	n.mtx.Unlock()

	serverRead, clientWrite := io.Pipe()
	clientRead, serverWrite := io.Pipe()
	ownAddr := n.RandomAddress()
	server := &Conn{
		LAddress: address,
		RAddress: ownAddr,
		Reader:   serverRead,
		Writer:   serverWrite,
	}

	client := &Conn{
		LAddress: ownAddr,
		RAddress: address,
		Reader:   clientRead,
		Writer:   clientWrite,
	}

	select {
	case n.conns[address].Input <- server:
	// if a server cannot accept the connection then it returns an error.
	case <-time.After(timeout):
		return nil, errors.Errorf(
			"dial tcp %s: connect: connection refused", address)
	}

	return client, nil
}

func (n *Network) randomIPAddress(gen *rand.Rand) string {
	var octet []string
	for i := 0; i < 4; i++ {
		number := gen.Intn(255)
		octet = append(octet, strconv.Itoa(number))
	}

	return strings.Join(octet, ".")
}

func (n *Network) randomPort(gen *rand.Rand) string {
	return strconv.Itoa(gen.Intn(65534))
}

// RandomAddress creates random network address in the format address:port.
func (n *Network) RandomAddress() string {
	gen := rand.New(rand.NewSource(time.Now().UnixNano()))
	return n.randomIPAddress(gen) + ":" + n.randomPort(gen)
}
