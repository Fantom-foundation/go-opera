package gossip

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

func init() {
	// log.Root().SetHandler(log.LvlFilterHandler(log.LvlTrace, log.StreamHandler(os.Stderr, log.TerminalFormat(false))))
}

var testAccount, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")

// Tests that handshake failures are detected and reported correctly.
func TestStatusMsgErrors62(t *testing.T) { testStatusMsgErrors(t, fantom62) }

func testStatusMsgErrors(t *testing.T, protocol int) {
	pm, _ := newTestProtocolManagerMust(t, 5, 5, nil)
	var (
		genesis  = pm.engine.GetGenesisHash()
		progress = PeerProgress{
			Epoch: pm.engine.CurrentSuperFrameN(),
		}
		networkId = uint64(1)
	)
	defer pm.Stop()

	tests := []struct {
		code      uint64
		data      interface{}
		wantError error
	}{
		{
			code: TxMsg, data: []interface{}{},
			wantError: errResp(ErrNoStatusMsg, "first msg has code 2 (!= 0)"),
		},
		{
			code: StatusMsg, data: statusData{10, networkId, progress, genesis},
			wantError: errResp(ErrProtocolVersionMismatch, "10 (!= %d)", protocol),
		},
		{
			code: StatusMsg, data: statusData{uint32(protocol), 999, progress, genesis},
			wantError: errResp(ErrNetworkIdMismatch, "999 (!= %d)", networkId),
		},
		{
			code: StatusMsg, data: statusData{uint32(protocol), networkId, progress, hash.Hash{3}},
			wantError: errResp(ErrGenesisMismatch, "0300000000000000 (!= %x)", genesis.Bytes()[:8]),
		},
	}

	for i, test := range tests {
		p, errc := newTestPeer("peer", protocol, pm, false)
		// The send call might hang until reset because
		// the protocol might not read the payload.
		go p2p.Send(p.app, test.code, test.data)

		select {
		case err := <-errc:
			if err == nil {
				t.Errorf("test %d: protocol returned nil error, want %q", i, test.wantError)
			} else if err.Error() != test.wantError.Error() {
				t.Errorf("test %d: wrong error: got %q, want %q", i, err, test.wantError)
			}
		case <-time.After(2 * time.Second):
			t.Errorf("protocol did not shut down within 2 seconds")
		}
		p.close()
	}
}
