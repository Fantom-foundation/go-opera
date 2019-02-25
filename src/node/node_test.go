package node

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/common/hexutil"
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/dummy"
	"github.com/Fantom-foundation/go-lachesis/src/peer"
	"github.com/Fantom-foundation/go-lachesis/src/peer/fakenet"
	"github.com/Fantom-foundation/go-lachesis/src/peers"
	"github.com/Fantom-foundation/go-lachesis/src/poset"
)

type TestData struct {
	PoolSize   int
	Logger     *logrus.Logger
	Config     *Config
	BackConfig *peer.BackendConfig
	Network    *fakenet.Network
	CreateFu   peer.CreateSyncClientFunc
	Keys       []*ecdsa.PrivateKey
	Adds       []string
	PeersSlice []*peers.Peer
	Peers      *peers.Peers
}

func InitTestData(t *testing.T, peersCount int, poolSize int) *TestData {
	network, createFu := createNetwork()
	keys, p, adds := initPeers(peersCount, network)

	return &TestData{
		PoolSize:   poolSize,
		Logger:     common.NewTestLogger(t),
		Config:     TestConfig(t),
		BackConfig: peer.NewBackendConfig(),
		Network:    network,
		CreateFu:   createFu,
		Keys:       keys,
		Adds:       adds,
		PeersSlice: p.ToPeerSlice(),
		Peers:      p,
	}
}

func initPeers(
	number int, network *fakenet.Network) ([]*ecdsa.PrivateKey, *peers.Peers, []string) {

	var keys []*ecdsa.PrivateKey
	var adds []string

	ps := peers.NewPeers()

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

func createTransport(t *testing.T, logger logrus.FieldLogger,
	backConf *peer.BackendConfig, addr string, poolSize int,
	clientFu peer.CreateSyncClientFunc,
	listenerFu peer.CreateListenerFunc) peer.SyncPeer {

	producer := peer.NewProducer(poolSize, time.Second, clientFu)

	backend := peer.NewBackend(backConf, logger, listenerFu)
	if err := backend.ListenAndServe(peer.TCP, addr); err != nil {
		t.Fatal(err)
	}

	return peer.NewTransport(logger, producer, backend)
}

func transportClose(t *testing.T, syncPeer peer.SyncPeer) {
	if err := syncPeer.Close(); err != nil {
		t.Fatal(err)
	}
}

func createNode(t *testing.T, logger *logrus.Logger, config *Config,
	id uint64, key *ecdsa.PrivateKey, participants *peers.Peers,
	trans peer.SyncPeer, localAddr string, run bool) *Node {

	db := poset.NewInmemStore(participants, config.CacheSize, nil)
	app := dummy.NewInmemDummyApp(logger)

	selectorArgs := SmartPeerSelectorCreationFnArgs{
		LocalAddr: localAddr,
		GetFlagTable: nil,
	}

	node := NewNode(config, id, key, participants, db, trans, app, NewSmartPeerSelectorWrapper, selectorArgs, localAddr)
	if err := node.Init(); err != nil {
		t.Fatal(err)
	}

	go node.Run(run)

	return node
}

func gossip(
	nodes []*Node, target int64, shutdown bool, timeout time.Duration) error {
	for _, n := range nodes {
		node := n
		go func() {
			node.Run(true)
		}()
	}
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

func bombardAndWait(nodes []*Node, target int64, timeout time.Duration) error {

	quit := make(chan struct{})
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

func submitTransaction(n *Node, tx []byte) error {
	n.proxy.SubmitCh() <- []byte(tx)
	return nil
}

func recycleNodes(
	oldNodes []*Node, logger *logrus.Logger, t *testing.T) []*Node {
	var newNodes []*Node
	for _, oldNode := range oldNodes {
		newNode := recycleNode(oldNode, logger, t)
		newNodes = append(newNodes, newNode)
	}
	return newNodes
}

func recycleNode(oldNode *Node, logger *logrus.Logger, t *testing.T) *Node {
	conf := oldNode.conf
	id := oldNode.id
	key := oldNode.core.key
	ps := oldNode.peerSelector.Peers()

	var store poset.Store
	var err error
	if _, ok := oldNode.core.poset.Store.(*poset.BadgerStore); ok {
		store, err = poset.LoadBadgerStore(
			conf.CacheSize, oldNode.core.poset.Store.StorePath())
		if err != nil {
			t.Fatal(err)
		}
	} else {
		store = poset.NewInmemStore(oldNode.core.participants, conf.CacheSize, nil)
	}

	backConfig := peer.NewBackendConfig()
	network, createFu := createNetwork()
	p := ps.ToPeerSlice()

	// Create transport
	trans := createTransport(t, logger, backConfig, p[0].NetAddr,
		2, createFu, network.CreateListener)
	defer transportClose(t, trans)

	prox := dummy.NewInmemDummyApp(logger)

	selectorArgs := SmartPeerSelectorCreationFnArgs{
		LocalAddr: p[0].NetAddr,
		GetFlagTable: nil,
	}

	// Create & Init node
	newNode := NewNode(conf, id, key, ps, store, trans, prox, NewSmartPeerSelectorWrapper, selectorArgs, p[0].NetAddr)
	if err := newNode.Init(); err != nil {
		t.Fatal(err)
	}

	return newNode
}

func checkGossip(nodes []*Node, fromBlock int64, t *testing.T) {

	nodeBlocks := map[uint64][]poset.Block{}
	for _, n := range nodes {
		var blocks []poset.Block
		lastIndex := n.core.poset.Store.LastBlockIndex()
		for i := fromBlock; i < lastIndex; i++ {
			block, err := n.core.poset.Store.GetBlock(i)
			if err != nil {
				t.Fatalf("checkGossip: %v ", err)
			}
			blocks = append(blocks, block)
		}
		nodeBlocks[n.id] = blocks
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

func TestCreateAndInitNode(t *testing.T) {
	// Init data
	data := InitTestData(t, 1, 2)

	// Create transport
	trans := createTransport(t, data.Logger, data.BackConfig, data.Adds[0],
		data.PoolSize, data.CreateFu, data.Network.CreateListener)
	defer transportClose(t, trans)

	// Create & Init node
	node := createNode(t, data.Logger, data.Config, data.PeersSlice[0].ID, data.Keys[0], data.Peers, trans, data.Adds[0], false)

	// Check status
	nodeState := node.getState()
	if nodeState != Gossiping {
		t.Fatal(nodeState)
	}

	// Check ID
	if node.ID() != data.PeersSlice[0].ID {
		t.Fatal(node.id)
	}

	// Stop node & check status
	node.Stop()
	nodeState = node.getState()
	if nodeState != Stop {
		t.Fatal(nodeState)
	}

	// Shutdown node & check status
	node.Shutdown()
	nodeState = node.getState()
	if nodeState != Shutdown {
		t.Fatal(nodeState)
	}
}

func TestAddTransaction(t *testing.T) {
	// Init data
	data := InitTestData(t, 1, 2)

	// Create transport
	trans := createTransport(t, data.Logger, data.BackConfig, data.Adds[0],
		data.PoolSize, data.CreateFu, data.Network.CreateListener)
	defer transportClose(t, trans)

	// Create & Init node
	node := createNode(t, data.Logger, data.Config, data.PeersSlice[0].ID, data.Keys[0], data.Peers, trans, data.Adds[0], false)
	defer node.Shutdown()

	// Add new Tx
	message := "Test"
	err := node.addTransaction([]byte(message))
	if err != nil {
		t.Fatal(err.Error())
	}

	// Check tx pool
	txPoolCount := node.core.GetTransactionPoolCount()
	if txPoolCount != 1 {
		t.Fatal("Transaction pool count wrong")
	}

	// Add new Internal Tx
	internalTx := poset.InternalTransaction{}
	node.addInternalTransaction(internalTx)

	// Check internal tx pool
	txPoolCount = node.core.GetInternalTransactionPoolCount()
	if txPoolCount != 1 {
		t.Fatal("Transaction pool count wrong")
	}
}

func TestCommit(t *testing.T) {
	// Init data
	data := InitTestData(t, 1, 2)

	// Create transport
	trans := createTransport(t, data.Logger, data.BackConfig, data.Adds[0],
		data.PoolSize, data.CreateFu, data.Network.CreateListener)
	defer transportClose(t, trans)

	// Create & Init node
	node := createNode(t, data.Logger, data.Config, data.PeersSlice[0].ID, data.Keys[0], data.Peers, trans, data.Adds[0], false)
	defer node.Shutdown()

	// Create block
	block := poset.NewBlock(0, 1,
		[]byte("framehash"),
		[][]byte{
			[]byte("test1"),
			[]byte("test2"),
			[]byte("test3"),
		})

	// Make a commit
	err := node.commit(block)
	if err != nil {
		t.Fatal(err.Error())
	}

	testBlock, err := node.GetBlock(0)
	if err != nil {
		t.Fatal(err.Error())
	}

	blockhex := len(block.GetBody().Transactions)
	testBlockHex := len(testBlock.GetBody().Transactions)

	if blockhex != testBlockHex {
		t.Fatal(testBlockHex)
	}
}

func TestDoBackgroundWork(t *testing.T) {
	// Init data
	data := InitTestData(t, 1, 2)

	// Create transport
	trans := createTransport(t, data.Logger, data.BackConfig, data.Adds[0],
		data.PoolSize, data.CreateFu, data.Network.CreateListener)
	defer transportClose(t, trans)

	// Create & Init node
	node := createNode(t, data.Logger, data.Config, data.PeersSlice[0].ID, data.Keys[0], data.Peers, trans, data.Adds[0], false)
	defer node.Shutdown()

	// Check submitCh case
	message := "Test"
	node.submitCh <- []byte(message)

	// Because of we need to wait to complete submitCh.
	time.Sleep(3 * time.Second)

	txPoolCount := node.core.GetTransactionPoolCount()
	if txPoolCount != 1 {
		t.Fatal("Transaction pool count wrong")
	}

	// Check submitInternalCh case
	internalTx := poset.InternalTransaction{}
	node.submitInternalCh <- internalTx

	// Because of we need to wait to complete submitInternalCh.
	time.Sleep(3 * time.Second)

	txPoolCount = node.core.GetInternalTransactionPoolCount()
	if txPoolCount != 1 {
		t.Fatal("Transaction pool count wrong")
	}

	// Check commitCh case
	block := poset.NewBlock(0, 1,
		[]byte("framehash"),
		[][]byte{
			[]byte("test1"),
			[]byte("test2"),
			[]byte("test3"),
		})
	node.commitCh <- block

	// Because of we need to wait to complete commit.
	time.Sleep(3 * time.Second)

	testBlock, err := node.GetBlock(0)
	if err != nil {
		t.Fatal(err.Error())
	}

	blockhex := len(block.GetBody().Transactions)
	testBlockHex := len(testBlock.GetBody().Transactions)

	if blockhex != testBlockHex {
		t.Fatal(testBlockHex)
	}

	// Check signalTERMch case
	node.signalTERMch <- nil

	// Because of we need to wait to complete Shutdown.
	time.Sleep(3 * time.Second)

	if node.getState() != Shutdown {
		t.Fatal(node.getState())
	}
}

func TestSyncAndRequestSync(t *testing.T) {
	// Init data
	data := InitTestData(t, 2, 2)

	// Create transport
	trans1 := createTransport(t, data.Logger, data.BackConfig, data.Adds[0],
		data.PoolSize, data.CreateFu, data.Network.CreateListener)
	defer transportClose(t, trans1)

	trans2 := createTransport(t, data.Logger, data.BackConfig, data.Adds[1],
		data.PoolSize, data.CreateFu, data.Network.CreateListener)
	defer transportClose(t, trans2)

	// Create & Init node
	node1 := createNode(t, data.Logger, data.Config, data.PeersSlice[0].ID, data.Keys[0], data.Peers, trans1, data.Adds[0], false)
	defer node1.Shutdown()

	node2 := createNode(t, data.Logger, data.Config, data.PeersSlice[1].ID, data.Keys[1], data.Peers, trans2, data.Adds[1], false)
	defer node2.Shutdown()

	// Submit transaction for node
	message := "Test"
	node1.submitCh <- []byte(message)

	// Get known events from node1 & make sync request object
	node1KnownEvents := node1.core.KnownEvents()

	// Sync request
	resp, err := node1.requestSync(data.Adds[1], node1KnownEvents)
	if err != nil {
		t.Fatal(err)
	}

	// Sync events
	if err := node1.sync(data.PeersSlice[1], resp.Events); err != nil {
		t.Fatal(err)
	}

	// Check pool
	if l := len(node1.core.transactionPool); l > 0 {
		t.Fatalf("expected %d, got %d", 0, l)
	}

	// Get head & check tx count
	node1Head, err := node1.core.GetHead()
	if err != nil {
		t.Fatal(err)
	}

	if l := len(node1Head.Transactions()); l != 1 {
		t.Fatalf("expected %d, got %d", 1, l)
	}

	// Check message
	if m := string(node1Head.Transactions()[0]); m != message {
		t.Fatalf("expected message %s, got %s", message, m)
	}
}

func TestRequestEagerSyncAndEventDiff(t *testing.T) {
	// Init data
	data := InitTestData(t, 2, 2)

	// Create transport
	trans1 := createTransport(t, data.Logger, data.BackConfig, data.Adds[0],
		data.PoolSize, data.CreateFu, data.Network.CreateListener)
	defer transportClose(t, trans1)

	trans2 := createTransport(t, data.Logger, data.BackConfig, data.Adds[1],
		data.PoolSize, data.CreateFu, data.Network.CreateListener)
	defer transportClose(t, trans2)

	// Create & Init node
	node1 := createNode(t, data.Logger, data.Config, data.PeersSlice[0].ID, data.Keys[0], data.Peers, trans1, data.Adds[0], false)
	defer node1.Shutdown()

	node2 := createNode(t, data.Logger, data.Config, data.PeersSlice[1].ID, data.Keys[1], data.Peers, trans2, data.Adds[1], false)
	defer node2.Shutdown()

	// Get known events
	node1KnownEvents := node1.core.KnownEvents()

	// Get unknown events & convert it into wire
	unknownEvents, err := node1.EventDiff(node1KnownEvents)
	if err != nil {
		t.Fatal(err)
	}

	unknownWireEvents, err := node1.core.ToWire(unknownEvents)
	if err != nil {
		t.Fatal(err)
	}

	// Eager Sync
	result, err := node1.requestEagerSync(data.Adds[1], unknownWireEvents)
	if err != nil {
		t.Fatal(err)
	}

	expected := &peer.ForceSyncResponse{
		FromID:  node2.id,
		Success: true,
	}

	// Check result
	if expected.Success != result.Success {
		t.Fatalf("expected %v, got %v", expected.Success, result.Success)
	}

	if expected.FromID != result.FromID {
		t.Fatalf("expected %v, got %v", expected.FromID, result.FromID)
	}
}

func TestRequestFastForward(t *testing.T) {
	// Init data
	data := InitTestData(t, 2, 2)

	// Create transport
	trans1 := createTransport(t, data.Logger, data.BackConfig, data.Adds[0],
		data.PoolSize, data.CreateFu, data.Network.CreateListener)
	defer transportClose(t, trans1)

	trans2 := createTransport(t, data.Logger, data.BackConfig, data.Adds[1],
		data.PoolSize, data.CreateFu, data.Network.CreateListener)
	defer transportClose(t, trans2)

	// Create & Init node
	node1 := createNode(t, data.Logger, data.Config, data.PeersSlice[0].ID, data.Keys[0], data.Peers, trans1, data.Adds[0], false)
	defer node1.Shutdown()

	node2 := createNode(t, data.Logger, data.Config, data.PeersSlice[1].ID, data.Keys[1], data.Peers, trans2, data.Adds[1], false)
	defer node2.Shutdown()

	// Create frame
	frame := poset.Frame{
		Round: 1,
		Events: []*poset.EventMessage{
			&poset.EventMessage{
				Body: &poset.EventBody{
					Transactions: [][]byte{
						[]byte("test1"),
						[]byte("test2"),
						[]byte("test3"),
					},
				},
			},
		},
	}

	if err := node2.core.poset.Store.SetFrame(frame); err != nil {
		t.Fatal(err)
	}

	// Commit block
	block0 := poset.NewBlock(0, 1,
		[]byte("framehash"),
		[][]byte{
			[]byte("test1"),
			[]byte("test2"),
			[]byte("test3"),
		})

	node2.commitCh <- block0

	// Because of we need to wait to complete commit.
	time.Sleep(3 * time.Second)

	// Assign AnchorBlock
	node2.core.poset.AnchorBlock = new(int64)
	*node2.core.poset.AnchorBlock = 0

	block, _, err := node2.core.GetAnchorBlockWithFrame()
	if err != nil {
		t.Fatal(err)
	}

	// Fast forward request
	result, err := node1.requestFastForward(data.Adds[1])
	if err != nil {
		t.Fatal(err)
	}

	// Create expected result
	block, err = poset.NewBlockFromFrame(0, frame)
	if err != nil {
		t.Fatal(err)
	}

	// Assign test values
	b, err := node2.core.poset.Store.GetBlock(0)
	if err != nil {
		t.Fatal(err)
	}

	// TODO: Perhaps we should re-calculate the value before assign the value.
	block.Signatures = b.Signatures
	block.FrameHash = []byte("framehash")
	block.CreatedTime = b.CreatedTime

	// Because of we have the value in src/node/commit func with explanation.
	block.StateHash = []byte{0, 1, 2}

	// Get snapshot
	snapshot, err := node2.proxy.GetSnapshot(block.Index())
	if err != nil {
		t.Fatal(err)
	}

	// Create expected object
	expected := peer.FastForwardResponse{
		FromID:   node2.id,
		Block:    block,
		Frame:    frame,
		Snapshot: snapshot,
	}

	// Check actual result
	if !result.Block.Equals(&expected.Block) || !result.Frame.Equals(&expected.Frame) ||
		result.FromID != expected.FromID || !bytes.Equal(result.Snapshot, expected.Snapshot) {
		t.Fatalf("bad response, expected: %+v, got: %+v", expected, result)
	}

	hash1, err := expected.Frame.Hash()
	if err != nil {
		t.Fatal(err)
	}

	hash2, err := result.Frame.Hash()
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(hash1, hash2) {
		t.Fatalf("expected hash %s, got %s", hexutil.Encode(hash1), hexutil.Encode(hash2))
	}
}

func TestFastForward(t *testing.T) {

	data := InitTestData(t, 4, 2)

	// Create transport
	trans1 := createTransport(t, data.Logger, data.BackConfig, data.Adds[0],
		data.PoolSize, data.CreateFu, data.Network.CreateListener)
	defer transportClose(t, trans1)

	trans2 := createTransport(t, data.Logger, data.BackConfig, data.Adds[1],
		data.PoolSize, data.CreateFu, data.Network.CreateListener)
	defer transportClose(t, trans2)

	trans3 := createTransport(t, data.Logger, data.BackConfig, data.Adds[2],
		data.PoolSize, data.CreateFu, data.Network.CreateListener)
	defer transportClose(t, trans3)

	trans4 := createTransport(t, data.Logger, data.BackConfig, data.Adds[3],
		data.PoolSize, data.CreateFu, data.Network.CreateListener)
	defer transportClose(t, trans4)

	// Create & Init node
	node1 := createNode(t, data.Logger, data.Config, data.PeersSlice[0].ID, data.Keys[0], data.Peers, trans1, data.Adds[0], false)
	defer node1.Shutdown()

	node2 := createNode(t, data.Logger, data.Config, data.PeersSlice[1].ID, data.Keys[1], data.Peers, trans2, data.Adds[1], false)
	defer node2.Shutdown()

	node3 := createNode(t, data.Logger, data.Config, data.PeersSlice[2].ID, data.Keys[2], data.Peers, trans3, data.Adds[2], false)
	defer node3.Shutdown()

	node4 := createNode(t, data.Logger, data.Config, data.PeersSlice[3].ID, data.Keys[3], data.Peers, trans4, data.Adds[3], false)
	defer node4.Shutdown()

	nodes := []*Node{node1, node2, node3, node4}

	target := int64(3)
	err := gossip(nodes[1:], target, false, 60*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Second)

	err = nodes[0].fastForward()
	if err != nil {
		t.Fatalf("Error FastForwarding: %s", err)
	}

	lbi := nodes[0].GetLastBlockIndex()
	if lbi < 0 {
		t.Fatalf("LastBlockIndex is too low: %d", lbi)
	}
	sBlock, err := nodes[0].GetBlock(lbi)
	if err != nil {
		t.Fatalf("Error retrieving latest Block"+
			" from reset hasposetraph: %v", err)
	}
	expectedBlock, err := nodes[1].GetBlock(lbi)
	if err != nil {
		t.Fatalf("Failed to retrieve block %d from node1: %v", lbi, err)
	}
	if !reflect.DeepEqual(sBlock.Body, expectedBlock.Body) {
		t.Fatalf("Blocks defer")
	}
}

// TODO: Failed
func TestFastSync(t *testing.T) {
	var let sync.Mutex
	caught := false
	logger := common.NewTestLogger(t)
	
	poolSize := 2
	config := TestConfig(t)
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

	node1 := createNode(t, logger, config, ps[0].ID, keys[0], p, trans1, adds[0], true)
	defer node1.Shutdown()

	node2 := createNode(t, logger, config, ps[1].ID, keys[1], p, trans2, adds[1], true)
	defer node2.Shutdown()

	node3 := createNode(t, logger, config, ps[2].ID, keys[2], p, trans3, adds[2], true)
	defer node3.Shutdown()

	node4 := createNode(t, logger, config, ps[3].ID, keys[3], p, trans4, adds[3], true)
	defer node4.Shutdown()

	nodes := []*Node{node1, node2, node3, node4}

	var target int64 = 10

	err := gossip(nodes, target, false, 30*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	checkGossip(nodes, 0, t)

	node4 = nodes[3]
	node4.Shutdown()

	secondTarget := target + 10
	err = bombardAndWait(nodes[0:3], secondTarget, 30*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	checkGossip(nodes[0:3], 0, t)

	// Can't re-run it; have to reinstantiate a new node.
	node4 = recycleNode(node4, logger, t)

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
			if node4.getState() == CatchingUp {
				caught = true
				break
			}
		}
	}()

	node4.RunAsync(true)

	nodes[3] = node4

	// Gossip some more
	thirdTarget := secondTarget + 10
	err = bombardAndWait(nodes, thirdTarget, 25*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	start := node4.core.poset.FirstConsensusRound
	checkGossip(nodes, *start, t)
	let.Lock()
	let.Unlock()
	if !caught {
		t.Fatalf("Node4 didn't reach CatchingUp state")
	}
}

// TODO: Failed
func TestBootstrapAllNodes(t *testing.T) {
	logger := common.NewTestLogger(t)

	if err := os.RemoveAll("test_data"); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir("test_data", os.ModeDir|0777); err != nil {
		t.Fatal(err)
	}

	// create a first network with BadgerStore
	// and wait till it reaches 10 consensus rounds before shutting it down
	config := TestConfig(t)

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

	node1 := createNode(t, logger, config, ps[0].ID, keys[0], p, trans1, adds[0], true)
	defer node1.Shutdown()

	node2 := createNode(t, logger, config, ps[1].ID, keys[1], p, trans2, adds[1], true)
	defer node2.Shutdown()

	node3 := createNode(t, logger, config, ps[2].ID, keys[2], p, trans3, adds[2], true)
	defer node3.Shutdown()

	node4 := createNode(t, logger, config, ps[3].ID, keys[3], p, trans4, adds[3], true)
	defer node4.Shutdown()

	nodes := []*Node{node1, node2, node3, node4}

	err := gossip(nodes, 10, false, 3*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	checkGossip(nodes, 0, t)
	for _, n := range nodes {
		n.Shutdown()
	}

	// Now try to recreate a network from the databases created
	// in the first step and advance it to 20 consensus rounds
	newNodes := recycleNodes(nodes, logger, t)
	for _, n := range nodes {
		n.RunAsync(true)
	}
	err = gossip(newNodes, 20, false, 3*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	checkGossip(newNodes, 0, t)
	for _, n := range nodes {
		n.Shutdown()
	}

	// Check that both networks did not have
	// completely different consensus events
	checkGossip([]*Node{nodes[0], newNodes[0]}, 0, t)
}

func TestShutdown(t *testing.T) {
	logger := common.NewTestLogger(t)

	config := TestConfig(t)

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

	node1 := createNode(t, logger, config, ps[0].ID, keys[0], p, trans1, adds[0], false)
	defer node1.Shutdown()

	node2 := createNode(t, logger, config, ps[1].ID, keys[1], p, trans2, adds[1], false)
	defer node2.Shutdown()

	node3 := createNode(t, logger, config, ps[2].ID, keys[2], p, trans3, adds[2], false)
	defer node3.Shutdown()

	node4 := createNode(t, logger, config, ps[3].ID, keys[3], p, trans4, adds[3], false)
	defer node4.Shutdown()

	nodes := []*Node{node1, node2, node3, node4}

	nodes[0].Shutdown()

	// That modification of used counters should force SmartPeerSelector
	// to choose nodes[0] to gossip to
	// must be changed accordingly if PeerSelector is changed
	nodes[1].peerSelector.Peers().ByID[nodes[1].id].Used = 2
	nodes[1].peerSelector.Peers().ByID[nodes[2].id].Used = 2
	nodes[1].peerSelector.Peers().ByID[nodes[0].id].Used = 1

	err := nodes[1].gossip(nil)
	if err == nil {
		t.Fatal("Expected Timeout Error")
	}

	nodes[1].Shutdown()
}