package posnode

import (
	"net"
	"strconv"
)

// Config is a set of nodes params.
type Config struct {
	// count of event's parents (includes self-parent)
	EventParentsCount int
	// default service port
	Port int

	GossipThreads int
}

// DefaultConfig returns default config.
func DefaultConfig() *Config {
	return &Config{
		EventParentsCount: 3,
		Port:              55555,

		GossipThreads: 4,
	}
}

// NetAddrOf makes listen address from host and configured port.
func (n *Node) NetAddrOf(host string) string {
	port := strconv.Itoa(n.conf.Port)
	return net.JoinHostPort(host, port)
}
