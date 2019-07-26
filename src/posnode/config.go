package posnode

import (
	"net"
	"strconv"
	"time"
)

// Config is a set of nodes params.
type Config struct {
	EventParentsCount int // max count of event's parents (includes self-parent)
	Port              int // default service port

	GossipThreads    int           // count of gossiping goroutines
	MinEmitInterval  time.Duration // minimum event emission interval
	MaxEmitInterval  time.Duration // maximum event emission interval
	DiscoveryTimeout time.Duration // how often discovery should try to request

	ConnectTimeout time.Duration // how long dialer will for connection to be established
	ClientTimeout  time.Duration // how long will gRPC client will wait for response

	TopPeersCount int // peers hot cache size
}

// DefaultConfig returns default config.
func DefaultConfig() *Config {
	return &Config{
		EventParentsCount: 3,
		Port:              55555,

		GossipThreads:    4,
		MinEmitInterval:  10 * time.Second,
		MaxEmitInterval:  60 * time.Second,
		DiscoveryTimeout: 5 * time.Second,

		ConnectTimeout: 15 * time.Second,
		ClientTimeout:  15 * time.Second,

		TopPeersCount: 10,
	}
}

// NetAddrOf makes listen address from host and configured port.
func (n *Node) NetAddrOf(host string) string {
	port := strconv.Itoa(n.conf.Port)
	return net.JoinHostPort(host, port)
}
