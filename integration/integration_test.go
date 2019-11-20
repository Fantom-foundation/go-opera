package integration

import (
	"context"
	"fmt"
	"math/big"
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

	"github.com/Fantom-foundation/go-lachesis/lachesis"
	"github.com/Fantom-foundation/go-lachesis/lachesis/genesis"
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

	// fake net
	network := lachesis.FakeNetConfig(genesis.FakeAccounts(0, count, big.NewInt(0), 10000))

	// register a single gossip service
	services := map[string]adapters.ServiceFunc{
		"gossip": func(ctx *adapters.ServiceContext) (node.Service, error) {
			g := NewIntegration(ctx, network)
			return g, nil
		},
	}
	registerGossip.Do(func() {
		adapters.RegisterServices(services)
	})

	// create the NodeAdapter
	var adapter adapters.NodeAdapter
	adapter = adapters.NewSimAdapter(services)

	// create network
	sim := simulations.NewNetwork(adapter, &simulations.NetworkConfig{
		DefaultService: serviceNames(services)[0],
	})

	// create and start nodes
	nodes := make([]enode.ID, count)
	addrs := network.Genesis.Validators.Addresses()
	for i, addr := range addrs {
		key := network.Genesis.Alloc[addr].PrivateKey
		id := enode.PubkeyToIDV4(&key.PublicKey)
		config := &adapters.NodeConfig{
			ID:         id,
			Name:       fmt.Sprintf("Node-%d", i),
			PrivateKey: key,
			Services:   serviceNames(services),
		}

		_, err := sim.NewNodeWithConfig(config)
		if err != nil {
			panic(err)
		}

		nodes[i] = id
	}

	sim.StartAll()
	defer sim.Shutdown()

	connect(sim, nodes)

	// start
	srv := &http.Server{
		Addr:    ":8888",
		Handler: simulations.NewServer(sim),
	}
	go func() {
		log.Info("Starting simulation server on 0.0.0.0:8888...")
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Crit("Error starting simulation server", "err", err)
		}
	}()

	// stop
	<-time.After(5 * time.Second)

	if err := srv.Shutdown(context.TODO()); err != nil {
		log.Crit("Error stopping simulation server", "err", err)
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
