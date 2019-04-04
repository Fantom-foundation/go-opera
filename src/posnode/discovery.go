package posnode

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/api"
)

const (
	// how much time should discovry wait
	// after last failure request. 15 minutes.
	waitOnDiscoveryFailure = 60 * 15 * time.Second
)

// discovery is a network discovery process.
type discovery struct {
	tasks chan discoveryTask
	done  chan struct{}
	store *discoveries
}

// discoveryTask contains data required
// for peer discovery, source is a node
// from which we receive information
// about unknown node, host is a host
// address of the source node.
type discoveryTask struct {
	source, unknown hash.Peer
	host            string
}

// StartDiscovery starts single thread network discovery.
// It should be called once.
func (n *Node) StartDiscovery() {
	n.discovery.tasks = make(chan discoveryTask, 100) // NOTE magic buffer size.
	n.discovery.done = make(chan struct{})

	go func() {
		for {
			select {
			case task := <-n.discovery.tasks:
				n.AskPeerInfo(task.source, task.unknown, task.host)
				peerInfoAsked()
			case <-n.discovery.done:
				return
			}
		}
	}()
}

var peerInfoAsked = func() {}

// StopDiscovery stops network discovery.
// It should be called once.
func (n *Node) StopDiscovery() {
	close(n.discovery.done)
}

// CheckPeerIsKnown checks peer is known otherwise makes discovery task.
// It also checks that maybe it should skip discovery for given source and
// host because previous one request was failed.
func (n *Node) CheckPeerIsKnown(source, id hash.Peer, host string) {
	// Skip discovery if peer exists in storage.
	peerInfo := n.store.GetWirePeer(id)
	if peerInfo != nil {
		return
	}

	// Skip discovery if host was unavailable.
	discovery := n.discovery.store.GetDiscovery(source)
	if shouldSkipDiscovery(discovery, host) {
		return
	}

	select {
	case n.discovery.tasks <- discoveryTask{
		source:  source,
		unknown: id,
		host:    host,
	}:
	default:
		n.log.Warn("discovery.tasks queue is full, so skipped")
	}
}

// AskPeerInfo gets peer info (network address, public key, etc).
func (n *Node) AskPeerInfo(source, id hash.Peer, host string) {
	// Check if should skip discovery.
	discovery := n.discovery.store.GetDiscovery(source)
	if shouldSkipDiscovery(discovery, host) {
		return
	}

	// Prepare client to make request.
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	cli, err := n.ConnectTo(ctx, host)
	if err != nil {
		// TODO: read about possible errors here, can't get
		// any, status always IDLE. Should we set this
		// source as unavailable?
		n.log.Error(errors.Wrapf(err, "connect to: %s", host))
		return
	}

	// Make request to get information about peer.
	if discovery == nil {
		discovery = &Discovery{
			ID:   source,
			Host: host,
		}
	}

	peerInfo, err := requestPeerInfo(cli, id.Hex())
	if err != nil {
		// TODO: think more about what we should do when
		// info about given id is not found by source.
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			n.discovery.store.SetDiscoveryAvailability(discovery, false)
			n.log.Error(fmt.Sprintf("id: %s not found at host: %s with id: %s", id.Hex(), host, source.Hex()))
			return
		}
		// Set availability for this source as unavailable.
		n.discovery.store.SetDiscoveryAvailability(discovery, false)
		return
	}

	pKey := CalcNodeID(common.ToECDSAPub(peerInfo.PubKey))
	if pKey.Hex() != id.Hex() {
		n.discovery.store.SetDiscoveryAvailability(discovery, false)
		return
	}

	// Set availability for source as available and
	// and new peer into the store.
	n.SetPeerHost(id, peerInfo.Host)
	n.discovery.store.SetDiscoveryAvailability(discovery, true)
	peer := WireToPeer(peerInfo)
	n.store.SetPeer(peer)
}

// If last request was failure, host is same and
// not waited enough until next try - then skip
// request.
func shouldSkipDiscovery(d *Discovery, host string) bool {
	if d == nil {
		return false
	}

	if d.Available {
		return false
	}

	if d.Host != host {
		return false
	}

	if time.Now().After(d.LastRequest.Add(waitOnDiscoveryFailure)) {
		return false
	}

	return true
}

// requestPeerInfo makes GetPeerInfo using NodeClient
// with context which hash timeout.
func requestPeerInfo(cli api.NodeClient, id string) (*api.PeerInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()
	in := api.PeerRequest{
		PeerID: id,
	}
	return cli.GetPeerInfo(ctx, &in)
}

// Discovery is a representation of discovery try.
type Discovery struct {
	ID          hash.Peer
	Host        string
	LastRequest time.Time
	Available   bool
}

// ToWire converts to protobuf message.
func (p *Discovery) ToWire() *api.DiscoveryInfo {
	return &api.DiscoveryInfo{
		ID:          p.ID.Hex(),
		Host:        p.Host,
		LastRequest: p.LastRequest.UnixNano(),
		Available:   p.Available,
	}
}

// WireToDiscovery converts from protobuf message.
func WireToDiscovery(w *api.DiscoveryInfo) *Discovery {
	return &Discovery{
		ID:          hash.BytesToPeer(common.FromHex(w.ID)),
		Host:        w.Host,
		LastRequest: time.Unix(0, w.LastRequest),
		Available:   w.Available,
	}
}

type discoveries struct {
	sync.RWMutex
	discoveries map[string]*Discovery
}

func (ds *discoveries) GetDiscovery(id hash.Peer) *Discovery {
	ds.RLock()
	defer ds.RUnlock()

	for idHex, discovery := range ds.discoveries {
		if idHex == id.Hex() {
			return discovery
		}
	}

	return nil
}

func (ds *discoveries) SetDiscovery(discovery *Discovery) {
	ds.Lock()
	defer ds.Unlock()

	ds.discoveries[discovery.ID.Hex()] = discovery
}

func (ds *discoveries) SetDiscoveryAvailability(d *Discovery, available bool) {
	d.Available = available
	d.LastRequest = time.Now()
	ds.SetDiscovery(d)
}

func (ds *discoveries) Clear() {
	ds.Lock()
	defer ds.Unlock()

	ds.discoveries = make(map[string]*Discovery)
}
