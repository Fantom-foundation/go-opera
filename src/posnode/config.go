package posnode

import (
	"strconv"
)

type Config struct {
	// count of event's parents (includes self-parent)
	EventParentsCount int
	// default service port
	Port int
}

func DefaultConfig() *Config {
	return &Config{
		EventParentsCount: 3,
		Port:              55555,
	}
}

func (n *Node) NetAddrOf(host string) string {
	return host + ":" + strconv.Itoa(n.conf.Port)
}
