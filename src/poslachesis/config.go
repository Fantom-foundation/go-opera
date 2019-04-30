package lachesis

import (
	"net"
	"strconv"

	"github.com/Fantom-foundation/go-lachesis/src/posnode"
)

// Config of lachesis node.
type Config struct {
	Port     int
	CtrlPort int
	Node     posnode.Config
}

// DefaultConfig returns lachesis default config.
func DefaultConfig() *Config {
	return &Config{
		Port:     55556,
		CtrlPort: 55557,
		Node:     *posnode.DefaultConfig(),
	}
}

// ListenAddr returns listen address from host and configured port.
func (l *Lachesis) ListenAddr() string {
	port := strconv.Itoa(l.conf.Port)
	return net.JoinHostPort(l.host, port)
}

// CtrlListenAddr returns listen address from host and configured port
func (l *Lachesis) CtrlListenAddr() string {
	port := strconv.Itoa(l.conf.CtrlPort)
	return net.JoinHostPort("localhost", port)
}
