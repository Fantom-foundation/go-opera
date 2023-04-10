package main

import (
	"crypto/ecdsa"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"

	"github.com/Fantom-foundation/go-opera/cmd/opera/launcher"
)

func init() {
	glogger := log.NewGlogHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(false)))
	//glogger.Verbosity(log.LvlTrace)
	glogger.Verbosity(log.LvlDebug)
	//glogger.Verbosity(log.LvlInfo)
	log.Root().SetHandler(glogger)
}

const (
	genesisFile = "../blockchains/mainnet-109331-pruned-mpt.g"
	// genesisFile = "../blockchains/testnet-6226-full-mpt.g"
)

func main() {
	backend := NewProbeBackend()
	defer backend.Close()
	backend.LoadGenesis(genesisFile)

	s := newServer(backend)
	err := s.Start()
	if err != nil {
		panic(err)
	}
	defer s.Stop()

	wait()
}

func wait() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	<-sigs
}

func newServer(backend *ProbeBackend) *p2p.Server {
	var cfg = launcher.NodeDefaultConfig.P2P

	cfg.PrivateKey = anyKey()
	cfg.Protocols = ProbeProtocols(backend)

	for _, url := range launcher.Bootnodes[backend.Opera.Name] {
		cfg.BootstrapNodesV5 = append(cfg.BootstrapNodesV5, eNode(url))
	}
	/*
		cfg.BootstrapNodesV5 = append(cfg.BootstrapNodesV5[0:0],
			eNode("2ab8900c54a13cafebaace0cee178d696ef7f86fc09284445fa2df997c1a7f51@65.21.206.66:37523"),  // go-opera/v1.1.1 epoch=355299 block=89228301 atropos=197306:2379:605cb8
			eNode("e6d69102009c04e5fc6a9ca0be3d418f9678141c47e9a907f6c7b8ec3012fa5a@178.63.14.233:37523"), // go-opera/v1.1.1 epoch=356345 block=89437501 atropos=197306:2379:605cb8
			eNode("37c764c3785790380389f2892797e9b3825955be8d6eca77be09723d55ecfa8@65.21.229.175:37523"),  // go-opera/v1.1.1 epoch=357044 block=89577301 atropos=197306:2379:605cb8
		)
	*/
	cfg.BootstrapNodesV5 = append(cfg.BootstrapNodesV5[0:0],
		eNode("013d3dff53cd085ed494d1b9ede359746543f7457da157656fa1be13a7e22672cc6ad5af7a8d407b5aeaada7bbb8bd0b662e81348ac1db1fc5334d79be380d10@65.21.206.66:37523"),  // 362152 90598901 197306:2379:605cb8 (go-opera/v1.1.1)
		eNode("5ebd040b10b5493018b1f873e3e5734bced236b8d8156c049229f1580c46dd91ef95e7a6aea10c3c91c626698bd426de112e1926895e982a9ef88a9300ee45a6@65.21.229.175:37523"), // 362275 90623501 197306:2379:605cb8 (go-opera/v1.1.1)
		eNode("6729e74f461ef9dcd5c97d1e55479bfe28f210b5afda0988a33e0dda0a56f9e9d56365d0416e50887c96efedfc0144892b3073dffd6dc6d8e426735b71e62562@178.63.14.233:37523"), // 364737 91115901 197306:2379:605cb8 (go-opera/v1.1.1)
	)
	cfg.BootstrapNodes = cfg.BootstrapNodesV5

	return &p2p.Server{
		Config: cfg,
	}
}

func anyKey() *ecdsa.PrivateKey {
	key, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	return key
}

func eNode(url string) *enode.Node {
	if !strings.HasPrefix(url, "enode://") {
		url = "enode://" + url
	}
	n, err := enode.Parse(enode.ValidSchemes, url)
	if err != nil {
		panic(err)
	}
	return n
}
