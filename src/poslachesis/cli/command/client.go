package command

import (
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/proxy"
	"github.com/Fantom-foundation/go-lachesis/src/proxy/wire"
)

const (
	connTimeout   = 3 * time.Second
	clientTimeout = 3 * time.Second
	ctrlAddr      = "localhost:55557"
)

func newClient() (wire.CtrlClient, error) {
	client, err := proxy.NewCtrlClient(ctrlAddr, connTimeout)
	if err != nil {
		return nil, err
	}

	return client, nil
}
