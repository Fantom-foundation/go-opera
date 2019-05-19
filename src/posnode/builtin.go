package posnode

import (
	"sync"
)

// builtin is a set of some built in data.
type builtin struct {
	hosts []string

	sync.Mutex
}

// AddBuiltInPeers appends host names to built in peer list.
func (n *Node) AddBuiltInPeers(hosts ...string) {
	n.builtin.Lock()
	defer n.builtin.Unlock()

	n.builtin.hosts = append(n.builtin.hosts, hosts...)

	n.log.Debugf("built in peer hosts: %v", n.builtin.hosts)
}

// NextBuiltInPeer returns next peer host or empty string.
func (n *Node) NextBuiltInPeer() (host string) {
	n.builtin.Lock()
	defer n.builtin.Unlock()

	if len(n.builtin.hosts) == 0 {
		return
	}

	host = n.builtin.hosts[0]
	n.builtin.hosts = n.builtin.hosts[1:]
	return
}
