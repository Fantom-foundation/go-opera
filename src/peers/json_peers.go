package peers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

// JSONPeers is used to provide peer persistence on disk in the form
// of a JSON file. This allows human operators to manipulate the file.
type JSONPeers struct {
	l    sync.Mutex
	path string
}

// NewJSONPeers creates a new JSONPeers store.
func NewJSONPeers(base string) *JSONPeers {
	path := filepath.Join(base, jsonPeerPath)
	store := &JSONPeers{
		path: path,
	}
	return store
}

// Peers implements the PeerStore interface.
func (j *JSONPeers) Peers() (*Peers, error) {
	j.l.Lock()
	defer j.l.Unlock()

	// Read the file or create empty
	buf, err := ioutil.ReadFile(j.path)
	if err != nil {
		err = os.MkdirAll(filepath.Dir(j.path), 0750)
		if err != nil {
			return nil, err
		}
		f, err := os.OpenFile(j.path, os.O_CREATE|os.O_WRONLY, 0640)
		if err != nil {
			return nil, err
		}
		if err := f.Close(); err != nil {
			return nil, err
		}
	}

	// Decode the peers
	peerSet := make([]*Peer, len(buf))
	if len(buf) > 0 {
		dec := json.NewDecoder(bytes.NewReader(buf))
		if err := dec.Decode(&peerSet); err != nil {
			return nil, err
		}
	}

	if len(peerSet) == 0 {
		return nil, fmt.Errorf("peers not found")
	}

	return NewPeersFromSlice(peerSet), nil
}

// SetPeers implements the PeerStore interface.
func (j *JSONPeers) SetPeers(peers []*Peer) error {
	j.l.Lock()
	defer j.l.Unlock()

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(peers); err != nil {
		return err
	}

	// Write out as JSON
	return ioutil.WriteFile(j.path, buf.Bytes(), 0755)
}
