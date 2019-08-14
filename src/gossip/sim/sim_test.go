package sim

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/simulations"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/gossip"
)

type topology func(net *simulations.Network, nodes []enode.ID)

func TestStar(t *testing.T) {
	testSim(t, topologyStar)
}

func TestRing(t *testing.T) {
	testSim(t, topologyRing)
}

var registerGossip sync.Once

func testSim(t *testing.T, connect topology) {
	const count = 3

	// set the log level to Trace
	log.Root().SetHandler(log.LvlFilterHandler(
		log.LvlTrace,
		log.StreamHandler(os.Stderr, log.TerminalFormat(false))))

	// register a single gossip service
	services := map[string]adapters.ServiceFunc{
		"gossip": func(ctx *adapters.ServiceContext) (node.Service, error) {
			s := gossip.NewMemStore()
			gs := gossip.NewService(s, nil)
			return gs, nil
		},
	}
	registerGossip.Do(func() {
		adapters.RegisterServices(services)
	})

	// create the NodeAdapter
	var adapter adapters.NodeAdapter
	adapter = adapters.NewSimAdapter(services)

	// create network
	network := simulations.NewNetwork(adapter, &simulations.NetworkConfig{
		DefaultService: serviceNames(services)[0],
	})

	// create and start nodes
	nodes := make([]enode.ID, count)
	for i := 0; i < count; i++ {
		n := i + 1
		key := crypto.GenerateFakeKey(n)
		id := enode.PubkeyToIDV4(&key.PublicKey)
		config := &adapters.NodeConfig{
			ID:         id,
			Name:       fmt.Sprintf("Node-%d", n),
			PrivateKey: (*ecdsa.PrivateKey)(key),
			Services:   serviceNames(services),
		}

		_, err := network.NewNodeWithConfig(config)
		if err != nil {
			panic(err)
		}

		nodes[i] = id
	}

	network.StartAll()
	defer network.Shutdown()

	connect(network, nodes)

	// start
	srv := &http.Server{
		Addr:    ":8888",
		Handler: simulations.NewServer(network),
	}
	go func() {
		log.Info("starting simulation server on 0.0.0.0:8888...")
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Crit("error starting simulation server", "err", err)
		}
	}()

	// stop
	<-time.After(5 * time.Second)

	if err := srv.Shutdown(context.TODO()); err != nil {
		log.Crit("error stopping simulation server", "err", err)
	}
}

func topologyStar(net *simulations.Network, nodes []enode.ID) {
	if len(nodes) < 2 {
		return
	}
	err := net.ConnectNodesStar(nodes, nodes[0])
	if err != nil {
		panic(err)
	}
}

func topologyRing(net *simulations.Network, nodes []enode.ID) {
	if len(nodes) < 2 {
		return
	}
	err := net.ConnectNodesRing(nodes)
	if err != nil {
		panic(err)
	}
}

func serviceNames(services map[string]adapters.ServiceFunc) []string {
	names := make([]string, 0, len(services))
	for name := range services {
		names = append(names, name)
	}

	return names
}
