package lachesis

import (
	"net"
	"strconv"

	"github.com/Fantom-foundation/go-lachesis/src/posnode"
)

// Config of lachesis node.
type Config struct {
	AppPort  int
	CtrlPort int
	Node     posnode.Config
	LogLevel string
}

// DefaultConfig returns lachesis default config.
func DefaultConfig() *Config {
	return &Config{
		AppPort:  55556,
		CtrlPort: 55557,
		Node:     *posnode.DefaultConfig(),
		LogLevel: "debug",
	}
}

// AppListenAddr returns listen address for application connections.
func (l *Lachesis) AppListenAddr() string {
	port := strconv.Itoa(l.conf.AppPort)
	return net.JoinHostPort(l.host, port)
}

// CtrlListenAddr returns listen address for control connections.
func (l *Lachesis) CtrlListenAddr() string {
	port := strconv.Itoa(l.conf.CtrlPort)
	return net.JoinHostPort(l.host, port)
}
