package posnode

import (
	"sync"

	"github.com/asaskevich/govalidator"
)

// builtin is a set of some built in data.
type builtin struct {
	last  int
	hosts []string

	sync.Mutex
}

// AddBuiltInPeers checks in hosts and saves valid
// ones for futher peer discovery.
func (n *Node) AddBuiltInPeers(hosts ...string) {
	n.builtin.Lock()
	defer n.builtin.Unlock()

	for _, host := range hosts {
		if govalidator.IsHost(host) {
			n.builtin.hosts = append(n.builtin.hosts, host)
		}
	}

	n.log.Debugf("built in peer hosts: %v", n.builtin.hosts)
}

// NextBuiltInPeer infinitely returns one of the hosts
// in the order.
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
