package oldtests // TODO: Re-write the tests after refactor poset & move it back to node package.

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/dummy"
	"github.com/Fantom-foundation/go-lachesis/src/peer"
	"github.com/Fantom-foundation/go-lachesis/src/peer/fakenet"
	"github.com/Fantom-foundation/go-lachesis/src/peers"
	"github.com/Fantom-foundation/go-lachesis/src/poset"
)

func initPeers(
	number int, network *fakenet.Network) ([]*ecdsa.PrivateKey, *peers.Peers, []string) {
	var keys []*ecdsa.PrivateKey
	ps := peers.NewPeers()
	var adds []string

	for i := 0; i < number; i++ {
		key, _ := crypto.GenerateECDSAKey()
		keys = append(keys, key)
		addr := network.RandomAddress()
		adds = append(adds, addr)

		ps.AddPeer(peers.NewPeer(
			fmt.Sprintf("0x%X", crypto.FromECDSAPub(&keys[i].PublicKey)),
			addr,
		))
	}

	return keys, ps, adds
}

func transportClose(t *testing.T, syncPeer peer.SyncPeer) {
	if err := syncPeer.Close(); err != nil {
		t.Fatal(err)
	}
}

func checkSyncResponse(t *testing.T, exp, got *peer.SyncResponse) {
	if exp.FromID != got.FromID {
		t.Fatalf("expected %d, got %d", exp.FromID, got.FromID)
	}

	if l := len(got.Events); l != len(exp.Events) {
		t.Fatalf("expected %d items, got %d", len(exp.Events), l)
	}

	for i, e := range exp.Events {
		ex := got.Events[i]
		if !reflect.DeepEqual(e.Body, ex.Body) {
			t.Fatalf("expected %v, got %v", e.Body, ex.Body)
		}
	}

	if !reflect.DeepEqual(exp.Known, got.Known) {
		t.Fatalf("expected %v, got %v", exp.Known, got.Known)
	}
}

func checkForceSyncResponse(t *testing.T, exp, got *peer.ForceSyncResponse) {
	if exp.Success != got.Success {
		t.Fatalf("expected %v, got %v", exp.Success, got.Success)
	}

	if exp.FromID != got.FromID {
		t.Fatalf("expected %v, got %v", exp.FromID, got.FromID)
	}
}

func createTransport(t *testing.T, logger logrus.FieldLogger,
	backConf *peer.BackendConfig, addr string, poolSize int,
	clientFu peer.CreateSyncClientFunc,
	listenerFu peer.CreateListenerFunc) peer.SyncPeer {
	producer1 := peer.NewProducer(poolSize, time.Second, clientFu)
	backend1 := peer.NewBackend(backConf, logger, listenerFu)
	if err := backend1.ListenAndServe(peer.TCP, addr); err != nil {
		t.Fatal(err)
	}
	return peer.NewTransport(logger, producer1, backend1)
}

func runNode(t *testing.T, logger *logrus.Logger, config *Config,
	id uint64, key *ecdsa.PrivateKey, participants *peers.Peers,
	trans peer.SyncPeer, run bool) *Node {
	db := poset.NewInmemStore(participants, config.CacheSize, nil)
	app := dummy.NewInmemDummyApp(logger)
	node := NewNode(config, id, key, participants, db, trans, app)
	if err := node.Init(); err != nil {
		t.Fatal(err)
	}
	go node.Run(run)
	return node
}

func createNetwork() (*fakenet.Network, peer.CreateSyncClientFunc) {
	network := fakenet.NewNetwork()
	createFu := func(target string,
		timeout time.Duration) (peer.SyncClient, error) {
		rpcCli, err := peer.NewRPCClient(
			peer.TCP, target, time.Second, network.CreateNetConn)
		if err != nil {
			return nil, err
		}

		return peer.NewClient(rpcCli)
	}

	return network, createFu
}

func submitTransaction(done chan struct{}, n *Node, tx []byte) {
	select {
	case n.proxy.SubmitCh() <- tx:
	case <-done:
	}
}

func TestProcessSync(t *testing.T) {
	poolSize := 2
	logger := common.NewTestLogger(t)
	config := TestConfig(t)
	backConfig := peer.NewBackendConfig()
	network, createFu := createNetwork()
	keys, p, adds := initPeers(2, network)
	ps := p.ToPeerSlice()

	trans1 := createTransport(t, logger, backConfig, adds[0],
		poolSize, createFu, network.CreateListener)
	defer transportClose(t, trans1)

	trans2 := createTransport(t, logger, backConfig, adds[1],
		poolSize, createFu, network.CreateListener)
	defer transportClose(t, trans2)

	node1 := runNode(t, logger, config, ps[0].ID, keys[0], p, trans1, false)
	defer node1.Shutdown()

	node2 := runNode(t, logger, config, ps[1].ID, keys[1], p, trans2, false)
	defer node2.Shutdown()

	node0KnownEvents := node1.core.KnownEvents()
	node1KnownEvents := node2.core.KnownEvents()

	unknownEvents, err := node2.core.EventDiff(node0KnownEvents)
	if err != nil {
		t.Fatal(err)
	}

	unknownWireEvents, err := node2.core.ToWire(unknownEvents)
	if err != nil {
		t.Fatal(err)
	}

	req := &peer.SyncRequest{
		FromID: node1.id,
		Known:  node0KnownEvents,
	}

	exp := &peer.SyncResponse{
		FromID: node2.id,
		Events: unknownWireEvents,
		Known:  node1KnownEvents,
	}

	result := &peer.SyncResponse{}
	if err := trans1.Sync(context.Background(),
		adds[1], req, result); err != nil {
		t.Fatal(err)
	}

	checkSyncResponse(t, exp, result)
}

func TestProcessEagerSync(t *testing.T) {
	poolSize := 2
	logger := common.NewTestLogger(t)
	config := TestConfig(t)
	backConfig := peer.NewBackendConfig()
	network, createFu := createNetwork()
	keys, p, adds := initPeers(2, network)
	ps := p.ToPeerSlice()

	trans1 := createTransport(t, logger, backConfig, adds[0],
		poolSize, createFu, network.CreateListener)
	defer transportClose(t, trans1)

	trans2 := createTransport(t, logger, backConfig, adds[1],
		poolSize, createFu, network.CreateListener)
	defer transportClose(t, trans2)

	node1 := runNode(t, logger, config, ps[0].ID, keys[0], p, trans1, false)
	defer node1.Shutdown()

	node2 := runNode(t, logger, config, ps[1].ID, keys[1], p, trans2, false)
	defer node2.Shutdown()

	node1KnownEvents := node1.core.KnownEvents()

	unknownEvents, err := node1.core.EventDiff(node1KnownEvents)
	if err != nil {
		t.Fatal(err)
	}

	unknownWireEvents, err := node1.core.ToWire(unknownEvents)
	if err != nil {
		t.Fatal(err)
	}

	req := peer.ForceSyncRequest{
		FromID: node1.id,
		Events: unknownWireEvents,
	}
	exp := &peer.ForceSyncResponse{
		FromID:  node2.id,
		Success: true,
	}

	result := &peer.ForceSyncResponse{}
	if err := trans1.ForceSync(context.Background(),
		adds[1], &req, result); err != nil {
		t.Fatal(err)
	}

	checkForceSyncResponse(t, exp, result)
}

// TODO: async Sync
//func TestAddTransaction(t *testing.T) {
//	poolSize := 2
//	logger := common.NewTestLogger(t)
//	config := TestConfig(t)
//	backConfig := peer.NewBackendConfig()
//	network, createFu := createNetwork()
//	keys, p, adds := initPeers(2, network)
//	ps := p.ToPeerSlice()
//
//	trans1 := createTransport(t, logger, backConfig, adds[0],
//		poolSize, createFu, network.CreateListener)
//	defer transportClose(t, trans1)
//
//	trans2 := createTransport(t, logger, backConfig, adds[1],
//		poolSize, createFu, network.CreateListener)
//	defer transportClose(t, trans2)
//
//	node1 := runNode(t, logger, config, ps[0].ID, keys[0], p, trans1, false)
//	defer node1.Shutdown()
//
//	node2 := runNode(t, logger, config, ps[1].ID, keys[1], p, trans2, false)
//	defer node2.Shutdown()
//
//
//	message := "Hello World!"
//	submitTransaction(nil, node1, []byte(message))
//
//	node0KnownEvents := node1.core.KnownEvents()
//	args := peer.SyncRequest{
//		FromID: node1.id,
//		Known:  node0KnownEvents,
//	}
//
//	out := &peer.SyncResponse{}
//	if err := trans1.Sync(context.Background(),
//		adds[1], &args, out); err != nil {
//		t.Fatal(err)
//	}
//
//	if err := node1.sync(out.Events); err != nil {
//		t.Fatal(err)
//	}
//
//	if l := len(node1.core.transactionPool); l > 0 {
//		t.Fatalf("expected %d, got %d",0, l)
//	}
//
//	node1Head, err := node1.core.GetHead()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	if l := len(node1Head.Transactions()); l != 1 {
//		t.Fatalf("expected %d, got %d", 1, l)
//	}
//
//	if m := string(node1Head.Transactions()[0]); m != message {
//		t.Fatalf("expected message %s, got %s", message, m)
//	}
//}

func TestGossip(t *testing.T) {
	poolSize := 2
	logger := common.NewTestLogger(t)
	backConfig := peer.NewBackendConfig()
	backConfig.IdleTimeout = time.Second * 4
	backConfig.ProcessTimeout = time.Second * 2
	backConfig.ReceiveTimeout = time.Second * 2
	var target = 100

	config := NewConfig(
		5*time.Millisecond,
		time.Second,
		10000,
		1000,
		logger,
	)

	network, createFu := createNetwork()
	keys, p, adds := initPeers(4, network)
	ps := p.ToPeerSlice()

	trans1 := createTransport(t, logger, backConfig, adds[0],
		poolSize, createFu, network.CreateListener)
	defer transportClose(t, trans1)

	trans2 := createTransport(t, logger, backConfig, adds[1],
		poolSize, createFu, network.CreateListener)
	defer transportClose(t, trans2)

	trans3 := createTransport(t, logger, backConfig, adds[2],
		poolSize, createFu, network.CreateListener)
	defer transportClose(t, trans3)

	trans4 := createTransport(t, logger, backConfig, adds[3],
		poolSize, createFu, network.CreateListener)
	defer transportClose(t, trans4)

	node1 := runNode(t, logger, config, ps[0].ID, keys[0], p, trans1, true)
	defer node1.Shutdown()

	node2 := runNode(t, logger, config, ps[1].ID, keys[1], p, trans2, true)
	defer node2.Shutdown()

	node3 := runNode(t, logger, config, ps[2].ID, keys[2], p, trans3, true)
	defer node3.Shutdown()

	node4 := runNode(t, logger, config, ps[3].ID, keys[3], p, trans4, true)
	defer node4.Shutdown()

	nodes := []*Node{node1, node2, node3, node4}

	done := make(chan struct{})
	defer close(done)

	makeRandomTransactions(done, nodes, done)

	stopper := time.After(time.Second * 20)

loop:
	for {
		select {
		case <-stopper:
			t.Fatal("timeout")
		default:
			time.Sleep(200 * time.Millisecond)

			found := 0

			for k := range nodes {
				ce := nodes[k].core.GetLastBlockIndex()
				if ce < int64(target) {
					continue loop
				}

				targetBlock, err := nodes[k].GetBlock(int64(target))
				if err != nil {
					continue loop
				}
				if len(targetBlock.GetStateHash()) > 0 {
					found++
				}
			}

			if found >= len(nodes) {
				break loop
			}
		}
	}

	checkGossip(t, 0, target, nodes)

	//nodes := initNodes(keys, ps, 1000, 1000, "inmem", logger, t)
	//
	//
	//err := gossip(nodes, target, true, 13*time.Second)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//s := NewService("127.0.0.1:3000", nodes[0], logger)
	//
	//srv := s.Serve()
	//
	//t.Logf("serving for 3 seconds")
	//shutdownTimeout := 3 * time.Second
	//time.Sleep(shutdownTimeout)
	//ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	//defer cancel()
	//t.Logf("stopping after waiting for Serve()...")
	//if err := srv.Shutdown(ctx); err != nil {
	//	t.Fatal(err) // failure/timeout shutting down the server gracefully
	//}
	//
	//checkGossip(nodes, 0, t)
}

//
//func TestMissingNodeGossip(t *testing.T) {
//
//	logger := common.NewTestLogger(t)
//
//	keys, ps := initPeers(4)
//	nodes := initNodes(keys, ps, 1000, 1000, "inmem", logger, t)
//	defer shutdownNodes(nodes)
//
//	err := gossip(nodes[1:], 10, true, 13*time.Second)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	checkGossip(nodes[1:], 0, t)
//}
//
//func TestSyncLimit(t *testing.T) {
//
//	logger := common.NewTestLogger(t)
//
//	keys, ps := initPeers(4)
//	nodes := initNodes(keys, ps, 1000, 1000, "inmem", logger, t)
//
//	err := gossip(nodes, 10, false, 3*time.Second)
//	if err != nil {
//		t.Fatal(err)
//	}
//	defer shutdownNodes(nodes)
//
//	// create fake node[0] known to artificially reach SyncLimit
//	node0KnownEvents := nodes[0].core.KnownEvents()
//	for k := range node0KnownEvents {
//		node0KnownEvents[k] = 0
//	}
//
//	args := transport.SyncRequest{
//		FromID: nodes[0].id,
//		Known:  node0KnownEvents,
//	}
//	expectedResp := transport.SyncResponse{
//		FromID:    nodes[1].id,
//		SyncLimit: true,
//	}
//
//	var out transport.SyncResponse
//	if err := nodes[0].trans.Sync(nodes[1].localAddr, &args, &out); err != nil {
//		t.Fatalf("err: %v", err)
//	}
//
//	// Verify the response
//	if expectedResp.FromID != out.FromID {
//		t.Fatalf("SyncResponse.FromID should be %d, not %d",
//			expectedResp.FromID, out.FromID)
//	}
//	if !expectedResp.SyncLimit {
//		t.Fatal("SyncResponse.SyncLimit should be true")
//	}
//}
//
//func TestFastForward(t *testing.T) {
//
//	logger := common.NewTestLogger(t)
//
//	keys, ps := initPeers(4)
//	nodes := initNodes(keys, ps, 1000, 1000,
//		"inmem", logger, t)
//	defer shutdownNodes(nodes)
//
//	target := int64(20)
//	err := gossip(nodes[1:], target, false, 15*time.Second)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	err = nodes[0].fastForward()
//	if err != nil {
//		t.Fatalf("Error FastForwarding: %s", err)
//	}
//
//	lbi := nodes[0].core.GetLastBlockIndex()
//	if lbi <= 0 {
//		t.Fatalf("LastBlockIndex is too low: %d", lbi)
//	}
//	sBlock, err := nodes[0].GetBlock(lbi)
//	if err != nil {
//		t.Fatalf("Error retrieving latest Block"+
//			" from reset hasposetraph: %v", err)
//	}
//	expectedBlock, err := nodes[1].GetBlock(lbi)
//	if err != nil {
//		t.Fatalf("Failed to retrieve block %d from node1: %v", lbi, err)
//	}
//	if !reflect.DeepEqual(sBlock.Body, expectedBlock.Body) {
//		t.Fatalf("Blocks defer")
//	}
//}
//
//func TestCatchUp(t *testing.T) {
//	logger := common.NewTestLogger(t)
//
//	// Create  config for 4 nodes
//	keys, ps := initPeers(4)
//
//	// Initialize the first 3 nodes only
//	normalNodes := initNodes(keys[0:3], ps, 1000, 400, "inmem", logger, t)
//	defer shutdownNodes(normalNodes)
//
//	target := int64(50)
//
//	err := gossip(normalNodes, target, false, 14*time.Second)
//	if err != nil {
//		t.Fatal(err)
//	}
//	checkGossip(normalNodes, 0, t)
//
//	node4 := initNodes(keys[3:], ps, 1000, 400, "inmem", logger, t)[0]
//
//	// Run parallel routine to check node4 eventually reaches CatchingUp state.
//	timeout := time.After(10 * time.Second)
//	go func() {
//		for {
//			select {
//			case <-timeout:
//				t.Fatalf("Timeout waiting for node4 to enter CatchingUp state")
//			default:
//			}
//			if node4.getState() == CatchingUp {
//				break
//			}
//		}
//	}()
//
//	node4.RunAsync(true)
//	defer node4.Shutdown()
//
//	// Gossip some more
//	nodes := append(normalNodes, node4)
//	newTarget := target + 20
//	err = bombardAndWait(nodes, newTarget, 10*time.Second)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	start := node4.core.poset.FirstConsensusRound
//	checkGossip(nodes, *start, t)
//}
//
//func TestFastSync(t *testing.T) {
//	logger := common.NewTestLogger(t)
//
//	// Create  config for 4 nodes
//	keys, ps := initPeers(4)
//	nodes := initNodes(keys, ps, 1000, 400, "inmem", logger, t)
//	defer shutdownNodes(nodes)
//
//	var target int64 = 50
//
//	err := gossip(nodes, target, false, 13*time.Second)
//	if err != nil {
//		t.Fatal(err)
//	}
//	checkGossip(nodes, 0, t)
//
//	node4 := nodes[3]
//	node4.Shutdown()
//
//	secondTarget := target + 50
//	err = bombardAndWait(nodes[0:3], secondTarget, 6*time.Second)
//	if err != nil {
//		t.Fatal(err)
//	}
//	checkGossip(nodes[0:3], 0, t)
//
//	// Can't re-run it; have to reinstantiate a new node.
//	node4 = recycleNode(node4, logger, t)
//
//	// Run parallel routine to check node4 eventually reaches CatchingUp state.
//	timeout := time.After(6 * time.Second)
//	go func() {
//		for {
//			select {
//			case <-timeout:
//				t.Fatalf("Timeout waiting for node4 to enter CatchingUp state")
//			default:
//			}
//			if node4.getState() == CatchingUp {
//				break
//			}
//		}
//	}()
//
//	node4.RunAsync(true)
//	defer node4.Shutdown()
//
//	nodes[3] = node4
//
//	// Gossip some more
//	thirdTarget := secondTarget + 20
//	err = bombardAndWait(nodes, thirdTarget, 6*time.Second)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	start := node4.core.poset.FirstConsensusRound
//	checkGossip(nodes, *start, t)
//}
//
//func TestShutdown(t *testing.T) {
//	logger := common.NewTestLogger(t)
//
//	keys, ps := initPeers(4)
//	nodes := initNodes(keys, ps, 1000, 1000, "inmem", logger, t)
//	runNodes(nodes, false)
//
//	nodes[0].Shutdown()
//
//	err := nodes[1].gossip(nodes[0].localAddr, nil)
//	if err == nil {
//		t.Fatal("Expected Timeout Error")
//	}
//
//	nodes[1].Shutdown()
//}
//
//func TestBootstrapAllNodes(t *testing.T) {
//	logger := common.NewTestLogger(t)
//
//	os.RemoveAll("test_data")
//	os.Mkdir("test_data", os.ModeDir|0777)
//
//	// create a first network with BadgerStore
//	// and wait till it reaches 10 consensus rounds before shutting it down
//	keys, ps := initPeers(4)
//	nodes := initNodes(keys, ps, 1000, 1000, "badger", logger, t)
//
//	err := gossip(nodes, 10, false, 3*time.Second)
//	if err != nil {
//		t.Fatal(err)
//	}
//	checkGossip(nodes, 0, t)
//	shutdownNodes(nodes)
//
//	// Now try to recreate a network from the databases created
//	// in the first step and advance it to 20 consensus rounds
//	newNodes := recycleNodes(nodes, logger, t)
//	err = gossip(newNodes, 20, false, 3*time.Second)
//	if err != nil {
//		t.Fatal(err)
//	}
//	checkGossip(newNodes, 0, t)
//	shutdownNodes(newNodes)
//
//	// Check that both networks did not have
//	// completely different consensus events
//	checkGossip([]*Node{nodes[0], newNodes[0]}, 0, t)
//}

//func gossip(
//	nodes []*Node, target int64, shutdown bool, timeout time.Duration) error {
//	runNodes(nodes, true)
//	err := bombardAndWait(nodes, target, timeout)
//	if err != nil {
//		return err
//	}
//	if shutdown {
//		shutdownNodes(nodes)
//	}
//	return nil
//}

//func bombardAndWait(nodes []*Node, target int64, timeout time.Duration) error {
//
//	quit := make(chan struct{})
//	makeRandomTransactions(nodes, quit)
//
//	// wait until all nodes have at least 'target' blocks
//	stopper := time.After(timeout)
//	for {
//		select {
//		case <-stopper:
//			return fmt.Errorf("timeout")
//		default:
//		}
//		time.Sleep(10 * time.Millisecond)
//		done := true
//		for _, n := range nodes {
//			ce := n.core.GetLastBlockIndex()
//			if ce < target {
//				done = false
//				break
//			} else {
//				// wait until the target block has retrieved a state hash from
//				// the app
//				targetBlock, _ := n.core.poset.Store.GetBlock(target)
//				if len(targetBlock.GetStateHash()) == 0 {
//					done = false
//					break
//				}
//			}
//		}
//		if done {
//			break
//		}
//	}
//	close(quit)
//	return nil
//}

type Service struct {
	bindAddress string
	node        *Node
	graph       *Graph
	logger      *logrus.Logger
}

func NewService(bindAddress string, n *Node, logger *logrus.Logger) *Service {
	service := Service{
		bindAddress: bindAddress,
		node:        n,
		graph:       NewGraph(n),
		logger:      logger,
	}

	return &service
}

func (s *Service) Serve() *http.Server {
	s.logger.WithField("bind_address", s.bindAddress).Debug("Service serving")

	http.HandleFunc("/stats", s.GetStats)

	http.HandleFunc("/block/", s.GetBlock)

	http.HandleFunc("/graph", s.GetGraph)

	srv := &http.Server{Addr: s.bindAddress, Handler: nil}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			s.logger.WithField("error", err).Error("Service failed")
		}
	}()

	return srv
}

func (s *Service) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := s.node.GetStats()

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		s.logger.WithError(err).Errorf("Failed to encode stats %v", stats)
	}
}

func (s *Service) GetBlock(w http.ResponseWriter, r *http.Request) {
	param := r.URL.Path[len("/block/"):]

	blockIndex, err := strconv.ParseInt(param, 10, 64)

	if err != nil {
		s.logger.WithError(err).Errorf("Parsing block_index parameter %s", param)

		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	block, err := s.node.GetBlock(blockIndex)

	if err != nil {
		s.logger.WithError(err).Errorf("Retrieving block %d", blockIndex)

		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(block); err != nil {
		s.logger.WithError(err).Errorf("Failed to encode block %v", block)
	}
}

func (s *Service) GetGraph(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(w)

	res := s.graph.GetInfos()

	if err := encoder.Encode(res); err != nil {
		s.logger.WithError(err).Errorf("Failed to encode Infos %v", res)
	}
}

func checkGossip(t *testing.T, fromBlock, toBlock int, nodes []*Node) {
	for block := fromBlock; block <= toBlock; block++ {
		for k := range nodes {
			if k == len(nodes)-1 {
				break
			}

			block1, err := nodes[k].GetBlock(int64(block))
			if err != nil {
				t.Fatal(err)

			}

			block2, err := nodes[k+1].GetBlock(int64(block))
			if err != nil {
				t.Fatal(err)
			}

			if !block1.Body.Equals(block2.Body) {
				f1, err := nodes[k].GetFrame(int64(block))
				if err != nil {
					t.Fatal(err)
				}
				f2, err := nodes[k+1].GetFrame(int64(block))
				if err != nil {
					t.Fatal(err)
				}

				if len(f1.Roots) != len(f2.Roots) {
					t.Fatalf("expected length of root %d, got %d", len(f1.Roots), len(f2.Roots))
					continue
				}

				for i := range f1.Roots {
					if !f1.Roots[i].Equals(f2.Roots[i]) {
						t.Fatalf("expected roots\n%+v\n%+v\n", f1.Roots[i], f2.Roots[i])
					}
				}

				if len(f1.Events) != len(f2.Events) {
					t.Fatalf("expected length of events %d, got %d", len(f1.Events), len(f2.Events))
					continue
				}

				for i := range f1.Events {
					if !f1.Events[i].Equals(f2.Events[i]) {
						t.Fatalf("expected event\n%+v\n%+v\n", f1.Events[i], f2.Events[i])
					}
				}
			}
		}
	}
}

func makeRandomTransactions(
	done chan struct{}, nodes []*Node, quit chan struct{}) {
	go func() {
		seq := make(map[int]int)
		for {
			select {
			case <-quit:
				return
			default:
				n := rand.Intn(len(nodes))
				node := nodes[n]
				submitTransaction(done, node, []byte(
					fmt.Sprintf("node%d transaction %d", n, seq[n])))
				seq[n] = seq[n] + 1
				time.Sleep(3 * time.Millisecond)
			}
		}
	}()
}

// TODO: fix it
//func BenchmarkGossip(b *testing.B) {
//	logger := common.NewTestLogger(b)
//	for n := 0; n < b.N; n++ {
//		keys, ps := initPeers(4)
//		nodes := initNodes(keys, ps, 1000, 1000, "inmem", logger, b)
//		gossip(nodes, 50, true, 3*time.Second)
//	}
//}

func runNodes(nodes []*Node, gossip bool) {
	for _, n := range nodes {
		node := n
		go func() {
			node.Run(gossip)
		}()
	}
}

func shutdownNodes(nodes []*Node) {
	for _, n := range nodes {
		n.Shutdown()
	}
}

func TestMain(m *testing.M) {
	runtime.SetBlockProfileRate(1)
	runtime.SetMutexProfileFraction(1)

	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	os.Exit(m.Run())
}
