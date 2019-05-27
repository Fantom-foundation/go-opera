package posnode

import (
	"sync"
)

// builtin is a set of built in peers to find network.
type builtin struct {
	last  int
	hosts []string

	sync.Mutex
}

// AddBuiltInPeers sets hosts for futher peer discovery.
func (n *Node) AddBuiltInPeers(hosts ...string) {
	n.builtin.Lock()
	defer n.builtin.Unlock()

	n.builtin.hosts = append(n.builtin.hosts, hosts...)

	n.Debugf("built in peer hosts: %v", n.builtin.hosts)
}

// NextBuiltInPeer returns one of builtin hosts.
func (n *Node) NextBuiltInPeer() (host string) {
	n.builtin.Lock()
	defer n.builtin.Unlock()

	if len(n.builtin.hosts) == 0 {
		return
	}

	host = n.builtin.hosts[n.builtin.last]
	n.builtin.last = (n.builtin.last + 1) % len(n.builtin.hosts)
	return
}
