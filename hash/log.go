package hash

import (
	"sync"

	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

var (
	nodeNameDictMu  sync.RWMutex
	eventNameDictMu sync.RWMutex

	// nodeNameDict is an optional dictionary to make node address human readable in log.
	nodeNameDict = make(map[idx.StakerID]string)

	// eventNameDict is an optional dictionary to make events human readable in log.
	eventNameDict = make(map[Event]string)
)

// SetNodeName sets an optional human readable alias of node address in log.
func SetNodeName(n idx.StakerID, name string) {
	nodeNameDictMu.Lock()
	defer nodeNameDictMu.Unlock()

	nodeNameDict[n] = name
}

// SetEventName sets an optional human readable alias of event hash in log.
func SetEventName(e Event, name string) {
	eventNameDictMu.Lock()
	defer eventNameDictMu.Unlock()

	eventNameDict[e] = name
}

// GetNodeName gets an optional human readable alias of node address.
func GetNodeName(n idx.StakerID) string {
	nodeNameDictMu.RLock()
	defer nodeNameDictMu.RUnlock()

	return nodeNameDict[n]
}

// GetEventName gets an optional human readable alias of event hash.
func GetEventName(e Event) string {
	eventNameDictMu.RLock()
	defer eventNameDictMu.RUnlock()

	return eventNameDict[e]
}
