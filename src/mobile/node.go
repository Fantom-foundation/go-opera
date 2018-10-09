package mobile

import (
	"fmt"

	"github.com/andrecronje/lachesis/src/crypto"
	"github.com/andrecronje/lachesis/src/lachesis"
	"github.com/andrecronje/lachesis/src/node"
	"github.com/andrecronje/lachesis/src/peers"
	"github.com/andrecronje/lachesis/src/proxy"
	"github.com/sirupsen/logrus"
)

type Node struct {
	nodeID int
	node   *node.Node
	proxy  proxy.AppProxy
	logger *logrus.Logger
}

// New initializes Node struct
func New(privKey string,
	nodeAddr string,
	participants *peers.Peers,
	commitHandler CommitHandler,
	exceptionHandler ExceptionHandler,
	config *MobileConfig) *Node {

	lachesisConfig := lachesis.NewDefaultConfig()

	lachesisConfig.Logger.WithFields(logrus.Fields{
		"nodeAddr": nodeAddr,
		"peers":    participants,
		"config":   fmt.Sprintf("%v", config),
	}).Debug("New Mobile Node")

	//Check private key
	pemKey := &crypto.PemKey{}

	key, err := pemKey.ReadKeyFromBuf([]byte(privKey))

	if err != nil {
		exceptionHandler.OnException(fmt.Sprintf("Failed to read private key: %s", err))

		return nil
	}

	lachesisConfig.Key = key

	// There should be at least two peers
	if participants.Len() < 2 {
		exceptionHandler.OnException(fmt.Sprintf("Should define at least two peers"))

		return nil
	}

	lachesisConfig.Proxy = newMobileAppProxy(commitHandler, exceptionHandler, lachesisConfig.Logger)
	lachesisConfig.LoadPeers = false

	engine := lachesis.NewLachesis(lachesisConfig)

	engine.Peers = participants

	if err := engine.Init(); err != nil {
		exceptionHandler.OnException(fmt.Sprintf("Cannot initialize engine: %s", err))

		return nil
	}

	return &Node{
		node:   engine.Node,
		proxy:  lachesisConfig.Proxy,
		nodeID: engine.Node.ID(),
		logger: lachesisConfig.Logger,
	}
}

func (n *Node) Run(async bool) {
	if async {
		n.node.RunAsync(true)
	} else {
		n.node.Run(true)
	}
}

func (n *Node) Shutdown() {
	n.node.Shutdown()
}

func (n *Node) SubmitTx(tx []byte) {
	//have to make a copy or the tx will be garbage collected and weird stuff
	//happens in transaction pool
	t := make([]byte, len(tx), len(tx))
	copy(t, tx)
	n.proxy.SubmitCh() <- t
}
