package lachesis

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/dummy"
	"github.com/Fantom-foundation/go-lachesis/src/node"
	"github.com/Fantom-foundation/go-lachesis/src/peer"
	"github.com/Fantom-foundation/go-lachesis/src/peer/fakenet"
	"github.com/Fantom-foundation/go-lachesis/src/peers"
	"github.com/Fantom-foundation/go-lachesis/src/poset"
	"github.com/Fantom-foundation/go-lachesis/src/utils"
	"github.com/sirupsen/logrus"
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

func createTransport(t testing.TB, logger logrus.FieldLogger,
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

func transportClose(t testing.TB, syncPeer peer.SyncPeer) {
	if err := syncPeer.Close(); err != nil {
		t.Fatal(err)
	}
}

func runNode(t testing.TB, logger *logrus.Logger, config *node.Config,
	id uint64, key *ecdsa.PrivateKey, participants *peers.Peers,
	trans peer.SyncPeer, localAddr string, run bool) *node.Node {
	db := poset.NewInmemStore(participants, config.CacheSize, nil)
	app := dummy.NewInmemDummyApp(logger)
	selectorArgs := node.SmartPeerSelectorCreationFnArgs{
		LocalAddr: localAddr,
		GetFlagTable: nil,
	}
	node := node.NewNode(config, id, key, participants, db, trans, app, node.NewSmartPeerSelectorWrapper, selectorArgs, localAddr)
	if err := node.Init(); err != nil {
		t.Fatal(err)
	}
	go node.Run(run)
	return node
}

func TestGossip(t *testing.T) {

	poolSize := 2
	logger := common.NewTestLogger(t)
	config := node.TestConfig(t)
	backConfig := peer.NewBackendConfig()

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
		
	node1 := runNode(t, logger, config, ps[0].ID, keys[0], p, trans1, adds[0], true)
	
	node2 := runNode(t, logger, config, ps[1].ID, keys[1], p, trans2, adds[1], true)
	
	node3 := runNode(t, logger, config, ps[2].ID, keys[2], p, trans3, adds[2], true)
	
	node4 := runNode(t, logger, config, ps[3].ID, keys[3], p, trans4, adds[3], true)
	
	nodes := []*node.Node{node1, node2, node3, node4}
	
	target := int64(1)
	
	err := gossip(nodes, target, true, 30*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	srvAddr := utils.GetUnusedNetAddr(1, t)
	s := NewService(srvAddr[0], nodes[0], logger)

	srv := s.Serve()

	t.Logf("serving for 3 seconds")
	shutdownTimeout := 3 * time.Second
	time.Sleep(shutdownTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	t.Logf("stopping after waiting for Serve()...")
	if err := srv.Shutdown(ctx); err != nil {
		t.Fatal(err) // failure/timeout shutting down the server gracefully
	}

	checkGossip(nodes, 0, t)
}

func TestMissingNodeGossip(t *testing.T) {

	logger := common.NewTestLogger(t)
	config := node.TestConfig(t)

	poolSize := 2
	backConfig := peer.NewBackendConfig()

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

	node1 := runNode(t, logger, config, ps[0].ID, keys[0], p, trans1, adds[0], true)

	node2 := runNode(t, logger, config, ps[1].ID, keys[1], p, trans2, adds[1], true)

	node3 := runNode(t, logger, config, ps[2].ID, keys[2], p, trans3, adds[2], true)

	node4 := runNode(t, logger, config, ps[3].ID, keys[3], p, trans4, adds[3], true)

	nodes := []*node.Node{node1, node2, node3, node4}

	err := gossip(nodes[1:], 3, true, 120*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	checkGossip(nodes[1:], 0, t)
}

func TestSyncLimit(t *testing.T) {

	logger := common.NewTestLogger(t)
	config := node.TestConfig(t)

	poolSize := 2
	backConfig := peer.NewBackendConfig()

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

	node1 := runNode(t, logger, config, ps[0].ID, keys[0], p, trans1, adds[0], true)
	defer node1.Shutdown()

	node2 := runNode(t, logger, config, ps[1].ID, keys[1], p, trans2, adds[1], true)
	defer node2.Shutdown()

	node3 := runNode(t, logger, config, ps[2].ID, keys[2], p, trans3, adds[2], true)
	defer node3.Shutdown()

	node4 := runNode(t, logger, config, ps[3].ID, keys[3], p, trans4, adds[3], true)
	defer node4.Shutdown()

	nodes := []*node.Node{node1, node2, node3, node4}

	err := gossip(nodes, 10, false, 30*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	// create fake node[0] known to artificially reach SyncLimit
	node0KnownEvents := nodes[0].GetKnownEvents()
	for k := range node0KnownEvents {
		node0KnownEvents[k] = 0
	}

	args := peer.SyncRequest{
		FromID: nodes[0].ID(),
		Known:  node0KnownEvents,
	}
	expectedResp := peer.SyncResponse{
		FromID:    nodes[1].ID(),
		SyncLimit: true,
	}

	var out peer.SyncResponse
	if err := trans1.Sync(context.Background(), adds[1], &args, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Verify the response
	if expectedResp.FromID != out.FromID {
		t.Fatalf("SyncResponse.FromID should be %d, not %d",
			expectedResp.FromID, out.FromID)
	}
	if !expectedResp.SyncLimit {
		t.Fatal("SyncResponse.SyncLimit should be true")
	}
}

// TODO: Failed
func TestCatchUp(t *testing.T) {
	var let sync.Mutex
	caught := false
	logger := common.NewTestLogger(t)
	config := node.TestConfig(t)

	poolSize := 2
	backConfig := peer.NewBackendConfig()

	network, createFu := createNetwork()
	keys, p, adds := initPeers(4, network)
	ps := p.ToPeerSlice()

	// Create  config for 4 nodes
	trans1 := createTransport(t, logger, backConfig, adds[0],
		poolSize, createFu, network.CreateListener)
	defer transportClose(t, trans1)

	trans2 := createTransport(t, logger, backConfig, adds[1],
		poolSize, createFu, network.CreateListener)
	defer transportClose(t, trans2)

	trans3 := createTransport(t, logger, backConfig, adds[2],
		poolSize, createFu, network.CreateListener)
	defer transportClose(t, trans3)

	// Initialize the first 3 nodes only
	node1 := runNode(t, logger, config, ps[0].ID, keys[0], p, trans1, adds[0], true)
	defer node1.Shutdown()

	node2 := runNode(t, logger, config, ps[1].ID, keys[1], p, trans2, adds[1], true)
	defer node2.Shutdown()

	node3 := runNode(t, logger, config, ps[2].ID, keys[2], p, trans3, adds[2], true)
	defer node3.Shutdown()

	normalNodes := []*node.Node{node1, node2, node3}

	target := int64(3)

	err := gossip(normalNodes, target, false, 30*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	checkGossip(normalNodes, 0, t)

	trans4 := createTransport(t, logger, backConfig, adds[3],
		poolSize, createFu, network.CreateListener)
	defer transportClose(t, trans4)

	node4 := runNode(t, logger, config, ps[3].ID, keys[3], p, trans4, adds[3], false)
	node4.Shutdown()

	// Run parallel routine to check node4 eventually reaches CatchingUp state.
	timeout := time.After(30 * time.Second)
	go func() {
		let.Lock()
		defer let.Unlock()
		for {
			select {
			case <-timeout:
				t.Logf("Timeout waiting for node4 to enter CatchingUp state")
				break
			default:
			}
			if node4.GetState() == node.CatchingUp {
				caught = true
				break
			}
		}
	}()

	node4.RunAsync(true)
	defer node4.Shutdown()

	// Gossip some more
	nodes := append(normalNodes, node4)
	newTarget := target + 4

	err = bombardAndWait(nodes, newTarget, 20*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	start := node4.GetFirstConsensusRound()
	checkGossip(nodes, *start, t)
	let.Lock()
	let.Unlock()
	if !caught {
		t.Fatalf("Node4 didn't reach CatchingUp state")
	}
}

func gossip(
	nodes []*node.Node, target int64, shutdown bool, timeout time.Duration) error {
	err := bombardAndWait(nodes, target, timeout)
	if err != nil {
		return err
	}
	if shutdown {
		for _, n := range nodes {
			n.Shutdown()
		}
	}
	return nil
}

func bombardAndWait(nodes []*node.Node, target int64, timeout time.Duration) error {

	quit := make(chan struct{})
	makeRandomTransactions(nodes, quit)
	tag := "beginning"

	// wait until all nodes have at least 'target' blocks
	stopper := time.After(timeout)
	for {
		select {
		case <-stopper:
			return fmt.Errorf("timeout in %v", tag)
		default:
		}
		time.Sleep(10 * time.Millisecond)
		done := true
		for _, n := range nodes {
			ce := n.GetLastBlockIndex()
			if ce < target {
				done = false
				tag = fmt.Sprintf("ce<target:%v<%v", ce, target)
				break
			} else {
				// wait until the target block has retrieved a state hash from
				// the app
				targetBlock, _ := n.GetBlock(target)
				if len(targetBlock.GetStateHash()) == 0 {
					done = false
					tag = "stateHash==0"
					break
				}
			}
		}
		if done {
			break
		}
	}
	close(quit)
	return nil
}

type Service struct {
	bindAddress string
	node        *node.Node
	graph       *node.Graph
	logger      *logrus.Logger
}

func NewService(bindAddress string, n *node.Node, logger *logrus.Logger) *Service {
	service := Service{
		bindAddress: bindAddress,
		node:        n,
		graph:       node.NewGraph(n),
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

func checkGossip(nodes []*node.Node, fromBlock int64, t *testing.T) {

	nodeBlocks := map[uint64][]poset.Block{}
	for _, n := range nodes {
		var blocks []poset.Block
		lastIndex := n.GetLastBlockIndex()
		for i := fromBlock; i < lastIndex; i++ {
			block, err := n.GetBlock(i)
			if err != nil {
				t.Fatalf("checkGossip: %v ", err)
			}
			blocks = append(blocks, block)
		}
		nodeBlocks[n.ID()] = blocks
	}

	minB := len(nodeBlocks[0])
	for k := uint64(1); k < uint64(len(nodes)); k++ {
		if len(nodeBlocks[k]) < minB {
			minB = len(nodeBlocks[k])
		}
	}

	for i, block := range nodeBlocks[0][:minB] {
		for k := uint64(1); k < uint64(len(nodes)); k++ {
			oBlock := nodeBlocks[k][i]
			if !reflect.DeepEqual(block.Body, oBlock.Body) {
				t.Fatalf("check gossip: difference in block %d."+
					" node 0: %v, node %d: %v",
					block.Index(), block.Body, k, oBlock.Body)
			}
		}
	}
}

func makeRandomTransactions(nodes []*node.Node, quit chan struct{}) {
	go func() {
		seq := make(map[int]int)
		for {
			select {
			case <-quit:
				return
			default:
				n := rand.Intn(len(nodes))
				node := nodes[n]
				if err := submitTransaction(node, []byte(
					fmt.Sprintf("node%d transaction %d", n, seq[n]))); err != nil {
					panic(err)
				}
				seq[n] = seq[n] + 1
				time.Sleep(3 * time.Millisecond)
			}
		}
	}()
}

func submitTransaction(n *node.Node, tx []byte) error {
	return n.SubmitCh(tx)
}


func BenchmarkGossip(b *testing.B) {
	logger := common.NewTestLogger(b)
	config := node.TestConfig(b)
	poolSize := 2
	backConfig := peer.NewBackendConfig()
	network, createFu := createNetwork()
	
	for n := 0; n < b.N; n++ {
		keys, p, adds := initPeers(4, network)
		ps := p.ToPeerSlice()

		trans1 := createTransport(b, logger, backConfig, adds[0],
			poolSize, createFu, network.CreateListener)
		defer transportClose(b, trans1)

		trans2 := createTransport(b, logger, backConfig, adds[1],
			poolSize, createFu, network.CreateListener)
		defer transportClose(b, trans2)

		trans3 := createTransport(b, logger, backConfig, adds[2],
			poolSize, createFu, network.CreateListener)
		defer transportClose(b, trans3)

		trans4 := createTransport(b, logger, backConfig, adds[3],
			poolSize, createFu, network.CreateListener)
		defer transportClose(b, trans4)

		node1 := runNode(b, logger, config, ps[0].ID, keys[0], p, trans1, adds[0], true)

		node2 := runNode(b, logger, config, ps[1].ID, keys[1], p, trans2, adds[1], true)

		node3 := runNode(b, logger, config, ps[2].ID, keys[2], p, trans3, adds[2], true)

		node4 := runNode(b, logger, config, ps[3].ID, keys[3], p, trans4, adds[3], true)

		nodes := []*node.Node{node1, node2, node3, node4}
		if err := gossip(nodes, 50, true, 3*time.Second); err != nil {
			b.Fatal(err)
		}
	}
}

