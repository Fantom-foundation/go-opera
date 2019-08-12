package gossip

import (
	"net"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rlp"
)

const (
	test_networkid = 10
)

var (
	genesis = common.HexToHash("genesis hash")
)

func newNodeID(t *testing.T) *enode.Node {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal("generate key err:", err)
	}
	return enode.NewV4(&key.PublicKey, net.IP{}, 35000, 35000)
}

func TestPeerHandshakeOK(t *testing.T) {
	id := newNodeID(t).ID()

	//peer to connect(on ulc side)
	p := peer{
		Peer:    p2p.NewPeer(id, "test peer", []p2p.Cap{}),
		version: protocol_version,
		network: test_networkid,
		rw: &rwStub{
			WriteHook: func(recvList keyValueList) {
				/*recv, _ := recvList.decode()
				var reqType uint64

				err := recv.get("announceType", &reqType)
				if err != nil {
					t.Fatal(err)
				}*/
			},
			ReadHook: func(l keyValueList) keyValueList {
				return l
			},
		},
	}

	err := p.Handshake(genesis, nil)
	if err != nil {
		t.Fatalf("Handshake error: %s", err)
	}
}

func TestPeerHandshakeErr(t *testing.T) {
	id := newNodeID(t).ID()
	p := peer{
		Peer:    p2p.NewPeer(id, "test peer", []p2p.Cap{}),
		version: protocol_version,
		network: test_networkid + 1,
		rw: &rwStub{
			ReadHook: func(l keyValueList) keyValueList {
				return l
			},
		},
	}

	err := p.Handshake(genesis, nil)
	if err == nil {
		t.FailNow()
	}
}

type rwStub struct {
	ReadHook  func(l keyValueList) keyValueList
	WriteHook func(l keyValueList)
}

func (s *rwStub) ReadMsg() (p2p.Msg, error) {
	payload := keyValueList{}
	payload = payload.add("protocolVersion", uint64(protocol_version))
	payload = payload.add("networkId", uint64(test_networkid))
	payload = payload.add("genesisHash", genesis)

	if s.ReadHook != nil {
		payload = s.ReadHook(payload)
	}

	size, p, err := rlp.EncodeToReader(payload)
	if err != nil {
		return p2p.Msg{}, err
	}

	return p2p.Msg{
		Size:    uint32(size),
		Payload: p,
	}, nil
}

func (s *rwStub) WriteMsg(m p2p.Msg) error {
	recvList := keyValueList{}
	if err := m.Decode(&recvList); err != nil {
		return err
	}

	if s.WriteHook != nil {
		s.WriteHook(recvList)
	}

	return nil
}
